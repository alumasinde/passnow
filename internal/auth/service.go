package auth

import (
	"context"
	"errors"
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
	dummyHashOnce.Do(func() {
		h, err := HashPassword("dummy-password-for-timing-parity", cost)
		if err != nil {
			h = "$2a$10$CwTycUXWue0Thq9StjUM0uJ8Q4Zi/tRJXQZQzY7X6O0Z6Y8n8f8f8"
		}
		dummyHash = h
	})
	return dummyHash
}

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrAccountLocked      = errors.New("auth: account temporarily locked")
	ErrAccountDisabled    = errors.New("auth: account disabled")
)

const (
	maxFailedLogins  = 5
	lockDurationSecs = 15 * 60
)

type Service struct {
	users           *users.Repository
	memberships     *roles.Repository
	refresh         *RefreshTokenRepository
	jwtSecret       []byte
	bcryptCost      int
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(userRepo *users.Repository, roleRepo *roles.Repository, refreshRepo *RefreshTokenRepository, jwtSecret []byte, bcryptCost int, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{users: userRepo, memberships: roleRepo, refresh: refreshRepo, jwtSecret: jwtSecret, bcryptCost: bcryptCost, accessTokenTTL: accessTTL, refreshTokenTTL: refreshTTL}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *Service) Login(ctx context.Context, tenantID int64, email, password string) (*TokenPair, *users.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		VerifyPassword(getDummyHash(s.bcryptCost), password)
		return nil, nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if u.IsLocked(now) {
		return nil, nil, ErrAccountLocked
	}
	if u.Status != users.StatusActive {
		return nil, nil, ErrAccountDisabled
	}

	tx, err := s.users.BeginTx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	if !VerifyPassword(u.PasswordHash, password) {
		if err := s.users.RegisterFailedLogin(ctx, tx, u.ID, lockDurationSecs, maxFailedLogins); err != nil {
			return nil, nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
		return nil, nil, ErrInvalidCredentials
	}

	if err := s.users.ResetFailedLogins(ctx, tx, u.ID); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	membership, err := s.memberships.MembershipFor(ctx, u.ID)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	if !membership.IsActive() {
		return nil, nil, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(ctx, u.ID, tenantID, membership.RoleID)
	if err != nil {
		return nil, nil, err
	}

	return pair, u, nil
}

func (s *Service) Refresh(ctx context.Context, tenantID int64, rawRefreshToken string) (*TokenPair, error) {
	userID, err := s.refresh.Consume(ctx, rawRefreshToken)
	if err != nil {
		return nil, err
	}
	u, err := s.users.ByID(ctx, userID)
	if err != nil || u.Status != users.StatusActive {
		return nil, ErrInvalidCredentials
	}
	membership, err := s.memberships.MembershipFor(ctx, userID)
	if err != nil || !membership.IsActive() {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokenPair(ctx, userID, tenantID, membership.RoleID)
}

func (s *Service) Logout(ctx context.Context, userID int64) error {
	return s.refresh.RevokeAllForUser(ctx, userID)
}

func (s *Service) issueTokenPair(ctx context.Context, userID, tenantID, roleID int64) (*TokenPair, error) {
	access, err := IssueAccessToken(s.jwtSecret, userID, tenantID, roleID, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}
	rawRefresh, hash, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.refresh.Store(ctx, userID, hash, s.refreshTokenTTL); err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: rawRefresh, ExpiresIn: int64(s.accessTokenTTL.Seconds())}, nil
}

func (s *Service) Register(ctx context.Context, in users.CreateInput) (*users.User, error) {
	hash, err := HashPassword(in.Password, s.bcryptCost)
	if err != nil {
		return nil, err
	}
	u := &users.User{Email: in.Email, PasswordHash: hash, FirstName: in.FirstName, LastName: in.LastName}
	id, err := s.users.Create(ctx, u)
	if err != nil {
		return nil, err
	}
	u.ID = id
	return u, nil
}


func (s *Service) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	if len(newPassword) < 8 { return errors.New("auth: password must be at least 8 characters") }
	u, err := s.users.ByID(ctx, userID); if err != nil { return ErrInvalidCredentials }
	if !VerifyPassword(u.PasswordHash, currentPassword) { return ErrInvalidCredentials }
	hash, err := HashPassword(newPassword, s.bcryptCost); if err != nil { return err }
	if err := s.users.ChangePassword(ctx, userID, hash); err != nil { return err }
	return s.refresh.RevokeAllForUser(ctx, userID)
}


func (s *Service) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	if len(newPassword) < 8 { return errors.New("auth: password must be at least 8 characters") }
	u, err := s.users.ByID(ctx, userID); if err != nil { return ErrInvalidCredentials }
	if !VerifyPassword(u.PasswordHash, currentPassword) { return ErrInvalidCredentials }
	hash, err := HashPassword(newPassword, s.bcryptCost); if err != nil { return err }
	if err := s.users.ChangePassword(ctx, userID, hash); err != nil { return err }
	return s.refresh.RevokeAllForUser(ctx, userID)
}
