// Package platform handles platform-level (cross-tenant) operations —
// currently just bootstrapping a brand new tenant with its first admin
// user. This is the ONLY endpoint in the system that runs outside tenant
// resolution (there's no tenant yet) and outside normal JWT auth (there's
// no user yet) — it is gated instead by a static bootstrap token that only
// the platform operator (you) holds. Rotate/disable this token in
// production once initial tenants are provisioned, or replace this with a
// proper platform-admin authentication scheme if self-service tenant
// signup is ever needed.
package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"gatepass/internal/auth"
	"gatepass/internal/httpx"
	"gatepass/internal/roles"
	"gatepass/internal/tenants"
	"gatepass/internal/tenantdb"
	"gatepass/internal/users"
)

var ErrSlugTaken = errors.New("platform: tenant slug already in use")

type Service struct {
	tenantRepo *tenants.Repository
	userRepo   *users.Repository
	roleRepo   *roles.Repository
	dbRepo     *tenantdb.Repository
	cipher     *tenantdb.Cipher
	installer  *tenantdb.Installer
	provisioner *tenantdb.Provisioner
	bcryptCost int
}

func NewService(tenantRepo *tenants.Repository, userRepo *users.Repository, roleRepo *roles.Repository, bcryptCost int) *Service {
	return &Service{tenantRepo: tenantRepo, userRepo: userRepo, roleRepo: roleRepo, bcryptCost: bcryptCost}
}

func (s *Service) WithTenantDatabase(repo *tenantdb.Repository, cipher *tenantdb.Cipher, installer *tenantdb.Installer, provisioner *tenantdb.Provisioner) *Service {
	s.dbRepo, s.cipher, s.installer, s.provisioner = repo, cipher, installer, provisioner
	return s
}

type BootstrapInput struct {
	TenantName     string `json:"tenant_name"`
	TenantSlug     string `json:"tenant_slug"`
	AdminEmail     string `json:"admin_email"`
	AdminPassword  string `json:"admin_password"`
	AdminFirstName string `json:"admin_first_name"`
	AdminLastName  string `json:"admin_last_name"`
	DatabaseMode string `json:"database_mode"`
	DatabaseHost string `json:"database_host"`
	DatabasePort string `json:"database_port"`
	DatabaseName string `json:"database_name"`
	DatabaseUsername string `json:"database_username"`
	DatabasePassword string `json:"database_password"`
}

type BootstrapResult struct {
	TenantID int64  `json:"tenant_id"`
	Slug     string `json:"slug"`
	AdminID  int64  `json:"admin_user_id"`
	RoleID   int64  `json:"role_id"`
	DatabaseStatus string `json:"database_status"`
}

// Bootstrap atomically creates: the tenant, a "Tenant Admin" system role
// granted every current permission, the first admin user (password
// supplied by the caller, hashed here — never generated/returned by the
// server, since this account isn't a "temporary password, change it
// later" case, it's the account setting everything else up), and the
// membership linking them. All in one transaction: a failure at any step
// leaves no partial tenant behind.
func (s *Service) Bootstrap(ctx context.Context, in BootstrapInput) (*BootstrapResult, error) {
	token, err := randomHex()
	if err != nil {
		return nil, err
	}
	hash, err := auth.HashPassword(in.AdminPassword, s.bcryptCost)
	if err != nil {
		return nil, err
	}

	tx, err := s.userRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	tenantID, err := s.tenantRepo.CreateTx(ctx, tx, &tenants.Tenant{
		Name: in.TenantName, Slug: in.TenantSlug, CustomDomainToken: token,
	})
	if err != nil {
		return nil, err
	}

	roleID, err := s.roleRepo.CreateRoleTx(ctx, tx, tenantID, "Tenant Admin", true)
	if err != nil {
		return nil, err
	}
	if err := s.roleRepo.GrantAllPermissions(ctx, tx, roleID); err != nil {
		return nil, err
	}

	userID, err := s.userRepo.CreateTx(ctx, tx, &users.User{
		Email: in.AdminEmail, PasswordHash: hash, FirstName: in.AdminFirstName, LastName: in.AdminLastName,
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.roleRepo.CreateMembershipTx(ctx, tx, tenantID, userID, roleID, roles.MembershipActive); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if s.dbRepo != nil && s.cipher != nil && s.installer != nil {
		if err := s.configureTenantDatabase(ctx, tenantID, in.TenantName, in.TenantSlug, token, in); err != nil {
			return nil, err
		}
	}

	return &BootstrapResult{TenantID: tenantID, Slug: in.TenantSlug, AdminID: userID, RoleID: roleID, DatabaseStatus: "ready"}, nil
}

func (s *Service) configureTenantDatabase(ctx context.Context, tenantID int64, name, slug, token string, in BootstrapInput) error {
	mode := strings.ToLower(strings.TrimSpace(in.DatabaseMode))
	if mode == "" { mode = "existing" }
	creds := tenantdb.Credentials{Host: strings.TrimSpace(in.DatabaseHost), Port: strings.TrimSpace(in.DatabasePort), Database: strings.TrimSpace(in.DatabaseName), Username: strings.TrimSpace(in.DatabaseUsername), Password: in.DatabasePassword}
	if creds.Port == "" { creds.Port = "3306" }
	if mode == "create" {
		if s.provisioner == nil || !s.provisioner.Enabled() { return errors.New("tenant database provisioning is not configured") }
		if creds.Host == "" { creds.Host = s.provisionerHost() }
		if err := s.provisioner.CreateDatabase(ctx, creds.Database); err != nil { return err }
	}
	if err := tenantdb.Verify(ctx, creds); err != nil { return err }
	secret, err := s.cipher.Encrypt(creds.Password); if err != nil { return err }
	conn := &tenantdb.Connection{TenantID: tenantID, Host: creds.Host, Port: creds.Port, DatabaseName: creds.Database, Username: creds.Username, EncryptedPassword: secret, Status: tenantdb.StatusVerified}
	if err := s.dbRepo.Upsert(ctx, conn); err != nil { return err }
	if err := s.installer.Install(ctx, creds, tenantID, name, slug, token); err != nil {
		msg := err.Error(); _ = s.dbRepo.MarkStatus(ctx, tenantID, tenantdb.StatusError, false, &msg); return err
	}
	return s.dbRepo.MarkStatus(ctx, tenantID, tenantdb.StatusReady, true, nil)
}

func (s *Service) provisionerHost() string { return "" }

func randomHex() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// --- HTTP handler ---------------------------------------------------------

type Handler struct {
	svc            *Service
	bootstrapToken string
}

func NewHandler(svc *Service, bootstrapToken string) *Handler {
	return &Handler{svc: svc, bootstrapToken: bootstrapToken}
}

func (h *Handler) BootstrapTenant(w http.ResponseWriter, r *http.Request) {
	// Constant-effort comparison isn't strictly necessary here (this token
	// is meant to be rotated/removed after initial setup, not a
	// long-lived secret defending live traffic), but requiring an exact
	// header match with no fallback keeps this closed-by-default: if
	// bootstrapToken is empty (unset in config), every request is
	// rejected rather than accidentally left open.
	if h.bootstrapToken == "" || r.Header.Get("X-Bootstrap-Token") != h.bootstrapToken {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	h.createTenant(w, r)
}

// CreateTenant is the platform-admin authenticated onboarding endpoint.
// The caller is already authenticated by PlatformAdmin middleware.
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	h.createTenant(w, r)
}

func (h *Handler) createTenant(w http.ResponseWriter, r *http.Request) {
	var in BootstrapInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.TenantName = strings.TrimSpace(in.TenantName)
	in.TenantSlug = strings.ToLower(strings.TrimSpace(in.TenantSlug))
	in.AdminEmail = strings.ToLower(strings.TrimSpace(in.AdminEmail))
	in.AdminFirstName = strings.TrimSpace(in.AdminFirstName)
	in.AdminLastName = strings.TrimSpace(in.AdminLastName)

	if in.TenantName == "" || in.TenantSlug == "" || in.AdminEmail == "" || in.AdminFirstName == "" || in.AdminLastName == "" || len(in.AdminPassword) < 12 {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("organization, slug, admin name, email, and a 12+ character password are required"))
		return
	}
	if !validSlug(in.TenantSlug) {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("slug must be 3-50 lowercase letters, numbers, or hyphens"))
		return
	}

	result, err := h.svc.Bootstrap(r.Context(), in)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			httpx.WriteError(w, httpx.ErrValidation.WithMessage("organization slug or administrator email is already in use"))
			return
		}
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}
