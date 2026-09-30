package platform

import (
    "context"
    "errors"
    "time"

    "gatepass/internal/auth"
    "gatepass/internal/users"
)

var (
    ErrInvalidPlatformCredentials = errors.New("platform: invalid credentials")
    ErrPlatformAdminDisabled = errors.New("platform: admin disabled")
)

type AdminService struct {
    admins *AdminRepository
    users *users.Repository
    jwtSecret []byte
    accessTTL time.Duration
}

func NewAdminService(admins *AdminRepository, users *users.Repository, jwtSecret []byte, accessTTL time.Duration) *AdminService {
    return &AdminService{admins: admins, users: users, jwtSecret: jwtSecret, accessTTL: accessTTL}
}

type AdminLoginResult struct {
    AccessToken string
    ExpiresIn int64
    User *users.User
    Role string
}

func (s *AdminService) Login(ctx context.Context, email, password string) (*AdminLoginResult, error) {
    u, err := s.users.ByEmail(ctx, email)
    if err != nil || u.Status != users.StatusActive || !auth.VerifyPassword(u.PasswordHash, password) {
        return nil, ErrInvalidPlatformCredentials
    }
    admin, err := s.admins.ByUserID(ctx, u.ID)
    if err != nil || admin.Status != "active" {
        return nil, ErrInvalidPlatformCredentials
    }

    roleID := int64(1)
    if admin.Role == "admin" { roleID = 2 }

    token, err := auth.IssueAccessToken(s.jwtSecret, u.ID, 0, roleID, s.accessTTL)
    if err != nil { return nil, err }

    return &AdminLoginResult{
        AccessToken: token,
        ExpiresIn: int64(s.accessTTL.Seconds()),
        User: u,
        Role: admin.Role,
    }, nil
}
