package platform

import (
	"context"
	"errors"
	"strings"
	"time"

	"gatepass/internal/auth"
	"gatepass/internal/users"
)

var (
	ErrInvalidPlatformCredentials = errors.New("platform: invalid credentials")
	ErrPlatformAdminDisabled      = errors.New("platform: admin disabled")
)

const (
	platformMaxFailedLogins = 5
	platformLockSeconds     = 15 * 60
)

type AdminService struct {
	admins     *AdminRepository
	users      *users.Repository
	jwtSecret  []byte
	accessTTL  time.Duration
	bcryptCost int
}

func NewAdminService(admins *AdminRepository, users *users.Repository, jwtSecret []byte, accessTTL time.Duration) *AdminService {
	return &AdminService{admins: admins, users: users, jwtSecret: jwtSecret, accessTTL: accessTTL, bcryptCost: 12}
}

// WithBcryptCost sets the cost used for the timing-parity dummy hash. Pass the
// same BCRYPT_COST the passwords were hashed with.
func (s *AdminService) WithBcryptCost(cost int) *AdminService {
	if cost > 0 {
		s.bcryptCost = cost
	}
	return s
}

type AdminLoginResult struct {
	AccessToken string
	ExpiresIn   int64
	User        *users.User
	Role        string
}

// Login authenticates a platform operator. Every failure path returns the same
// generic error AND does one bcrypt comparison, so neither the response nor its
// timing reveals whether the account exists, is locked, or is a platform admin.
func (s *AdminService) Login(ctx context.Context, email, password string) (*AdminLoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	dummy := func() { auth.VerifyPassword(auth.DummyPasswordHash(s.bcryptCost), password) }

	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		dummy()
		return nil, ErrInvalidPlatformCredentials
	}
	if u.IsLocked(time.Now().UTC()) {
		dummy()
		return nil, ErrInvalidPlatformCredentials
	}

	passwordOK := auth.VerifyPassword(u.PasswordHash, password)
	admin, adminErr := s.admins.ByUserID(ctx, u.ID)
	isAdmin := adminErr == nil && admin.Status == "active"

	if !passwordOK || u.Status != users.StatusActive || !isAdmin {
		// Count failures only against real platform admins.
		if !passwordOK && adminErr == nil {
			s.recordFailure(ctx, u.ID)
		}
		return nil, ErrInvalidPlatformCredentials
	}
	s.resetFailures(ctx, u.ID)

	roleID := int64(1)
	if admin.Role == "admin" {
		roleID = 2
	}

	token, err := auth.IssueAccessToken(s.jwtSecret, u.ID, 0, roleID, s.accessTTL)
	if err != nil {
		return nil, err
	}

	return &AdminLoginResult{
		AccessToken: token,
		ExpiresIn:   int64(s.accessTTL.Seconds()),
		User:        u,
		Role:        admin.Role,
	}, nil
}

func (s *AdminService) recordFailure(ctx context.Context, userID int64) {
	tx, err := s.users.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()
	if err := s.users.RegisterFailedLogin(ctx, tx, userID, platformLockSeconds, platformMaxFailedLogins); err != nil {
		return
	}
	_ = tx.Commit()
}

func (s *AdminService) resetFailures(ctx context.Context, userID int64) {
	tx, err := s.users.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()
	if err := s.users.ResetFailedLogins(ctx, tx, userID); err != nil {
		return
	}
	_ = tx.Commit()
}
