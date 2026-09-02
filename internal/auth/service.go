package auth

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"gatepass/internal/roles"
	"gatepass/internal/users"
)

var (
	dummyHashOnce sync.Once
	dummyHash string
)

func getDummyHash(cost int) string {
	dummyHashOnce.Do(func(){
		h,err:=HashPassword("dummy-password-for-timing-parity",cost)
		if err!=nil { h="$2a$10$CwTycUXWue0Thq9StjUM0uJ8Q4Zi/tRJXQZQzY7X6O0Z6Y8n8f8f8" }
		dummyHash=h
	})
	return dummyHash
}

var (
	ErrInvalidCredentials=errors.New("auth: invalid credentials")
	ErrAccountLocked=errors.New("auth: account temporarily locked")
	ErrAccountDisabled=errors.New("auth: account disabled")
)

const ( maxFailedLogins=5; lockDurationSecs=15*60 )

type Service struct {
	users *users.Repository
	memberships *roles.Repository
	refresh *RefreshTokenRepository
	jwtSecret []byte
	bcryptCost int
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
}

func NewService(userRepo *users.Repository,roleRepo *roles.Repository,refreshRepo *RefreshTokenRepository,jwtSecret []byte,bcryptCost int,accessTTL,refreshTTL time.Duration)*Service{
	return &Service{users:userRepo,memberships:roleRepo,refresh:refreshRepo,jwtSecret:jwtSecret,bcryptCost:bcryptCost,accessTokenTTL:accessTTL,refreshTokenTTL:refreshTTL}
}

type TokenPair struct { AccessToken string; RefreshToken string; ExpiresIn int64 }

func (s *Service) Login(ctx context.Context, tenantID int64,email,password string)(*TokenPair,*users.User,error){
	email = strings.ToLower(strings.TrimSpace(email))
	log.Printf("AUTH LOGIN START: tenant_id=%d email=%q", tenantID, email)

	u,err:=s.users.ByEmail(ctx,email)
	if err!=nil {
		log.Printf("AUTH LOGIN USER LOOKUP FAILED: tenant_id=%d email=%q error=%v; running dummy password verification", tenantID, email, err)
		VerifyPassword(getDummyHash(s.bcryptCost),password)
		return nil,nil,ErrInvalidCredentials
	}

	log.Printf("AUTH LOGIN USER FOUND: tenant_id=%d user_id=%d email=%q status=%q failed_login_count=%d locked_until=%v password_hash_length=%d", tenantID, u.ID, u.Email, u.Status, u.FailedLoginCount, u.LockedUntil, len(u.PasswordHash))

	now:=time.Now().UTC()
	if u.IsLocked(now){
		log.Printf("AUTH LOGIN REJECTED: account locked tenant_id=%d user_id=%d locked_until=%v", tenantID, u.ID, u.LockedUntil)
		return nil,nil,ErrAccountLocked
	}
	if u.Status!=users.StatusActive{
		log.Printf("AUTH LOGIN REJECTED: account disabled/inactive tenant_id=%d user_id=%d status=%q", tenantID, u.ID, u.Status)
		return nil,nil,ErrAccountDisabled
	}

	tx,err:=s.users.BeginTx(ctx)
	if err!=nil{
		log.Printf("AUTH LOGIN FAILED: begin transaction tenant_id=%d user_id=%d error=%v", tenantID, u.ID, err)
		return nil,nil,err
	}
	defer tx.Rollback()

	if !VerifyPassword(u.PasswordHash,password){
		log.Printf("AUTH LOGIN PASSWORD FAILED: tenant_id=%d user_id=%d email=%q", tenantID, u.ID, u.Email)
		if err:=s.users.RegisterFailedLogin(ctx,tx,u.ID,lockDurationSecs,maxFailedLogins);err!=nil{
			log.Printf("AUTH LOGIN FAILED: RegisterFailedLogin tenant_id=%d user_id=%d error=%v", tenantID, u.ID, err)
		}
		if err:=tx.Commit();err!=nil{
			log.Printf("AUTH LOGIN FAILED: commit failed after bad password tenant_id=%d user_id=%d error=%v", tenantID, u.ID, err)
		}
		return nil,nil,ErrInvalidCredentials
	}

	log.Printf("AUTH LOGIN PASSWORD PASSED: tenant_id=%d user_id=%d email=%q", tenantID, u.ID, u.Email)

	if err:=s.users.ResetFailedLogins(ctx,tx,u.ID);err!=nil{
		log.Printf("AUTH LOGIN FAILED: ResetFailedLogins tenant_id=%d user_id=%d error=%v", tenantID, u.ID, err)
		return nil,nil,err
	}
	if err:=tx.Commit();err!=nil{
		log.Printf("AUTH LOGIN FAILED: commit failed after successful password verification tenant_id=%d user_id=%d error=%v", tenantID, u.ID, err)
		return nil,nil,err
	}

	membership,err:=s.memberships.MembershipFor(ctx,u.ID)
	if err!=nil{
		log.Printf("AUTH LOGIN MEMBERSHIP LOOKUP FAILED: tenant_id=%d user_id=%d error=%v", tenantID, u.ID, err)
		return nil,nil,ErrInvalidCredentials
	}
	if !membership.IsActive(){
		log.Printf("AUTH LOGIN MEMBERSHIP INACTIVE: tenant_id=%d user_id=%d membership_id=%d role_id=%d status=%q", tenantID, u.ID, membership.ID, membership.RoleID, membership.Status)
		return nil,nil,ErrInvalidCredentials
	}

	log.Printf("AUTH LOGIN MEMBERSHIP PASSED: tenant_id=%d user_id=%d membership_id=%d role_id=%d", tenantID, u.ID, membership.ID, membership.RoleID)

	pair,err:=s.issueTokenPair(ctx,u.ID,tenantID,membership.RoleID)
	if err!=nil{
		log.Printf("AUTH LOGIN TOKEN ISSUANCE FAILED: tenant_id=%d user_id=%d role_id=%d error=%v", tenantID, u.ID, membership.RoleID, err)
		return nil,nil,err
	}

	log.Printf("AUTH LOGIN SUCCESS: tenant_id=%d user_id=%d role_id=%d", tenantID, u.ID, membership.RoleID)
	return pair,u,nil
}

func (s *Service) Refresh(ctx context.Context,tenantID int64,rawRefreshToken string)(*TokenPair,error){
	userID,err:=s.refresh.Consume(ctx,rawRefreshToken);if err!=nil{return nil,err}
	u,err:=s.users.ByID(ctx,userID);if err!=nil||u.Status!=users.StatusActive{return nil,ErrInvalidCredentials}
	membership,err:=s.memberships.MembershipFor(ctx,userID);if err!=nil||!membership.IsActive(){return nil,ErrInvalidCredentials}
	return s.issueTokenPair(ctx,userID,tenantID,membership.RoleID)
}

func (s *Service) Logout(ctx context.Context,userID int64)error{return s.refresh.RevokeAllForUser(ctx,userID)}

func (s *Service) issueTokenPair(ctx context.Context,userID,tenantID,roleID int64)(*TokenPair,error){
	access,err:=IssueAccessToken(s.jwtSecret,userID,tenantID,roleID,s.accessTokenTTL);if err!=nil{return nil,err}
	rawRefresh,hash,err:=NewRefreshToken();if err!=nil{return nil,err}
	if err:=s.refresh.Store(ctx,userID,hash,s.refreshTokenTTL);err!=nil{return nil,err}
	return &TokenPair{AccessToken:access,RefreshToken:rawRefresh,ExpiresIn:int64(s.accessTokenTTL.Seconds())},nil
}

func (s *Service) Register(ctx context.Context,in users.CreateInput)(*users.User,error){
	hash,err:=HashPassword(in.Password,s.bcryptCost);if err!=nil{return nil,err}
	u:=&users.User{Email:in.Email,PasswordHash:hash,FirstName:in.FirstName,LastName:in.LastName}
	id,err:=s.users.Create(ctx,u);if err!=nil{return nil,err};u.ID=id;return u,nil
}
