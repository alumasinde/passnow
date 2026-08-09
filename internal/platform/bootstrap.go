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

	"gatepass/internal/auth"
	"gatepass/internal/httpx"
	"gatepass/internal/roles"
	"gatepass/internal/tenants"
	"gatepass/internal/users"
)

var ErrSlugTaken = errors.New("platform: tenant slug already in use")

type Service struct {
	tenantRepo *tenants.Repository
	userRepo   *users.Repository
	roleRepo   *roles.Repository
	bcryptCost int
}

func NewService(tenantRepo *tenants.Repository, userRepo *users.Repository, roleRepo *roles.Repository, bcryptCost int) *Service {
	return &Service{tenantRepo: tenantRepo, userRepo: userRepo, roleRepo: roleRepo, bcryptCost: bcryptCost}
}

type BootstrapInput struct {
	TenantName     string `json:"tenant_name"`
	TenantSlug     string `json:"tenant_slug"`
	AdminEmail     string `json:"admin_email"`
	AdminPassword  string `json:"admin_password"`
	AdminFirstName string `json:"admin_first_name"`
	AdminLastName  string `json:"admin_last_name"`
}

type BootstrapResult struct {
	TenantID int64  `json:"tenant_id"`
	Slug     string `json:"slug"`
	AdminID  int64  `json:"admin_user_id"`
	RoleID   int64  `json:"role_id"`
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

	return &BootstrapResult{TenantID: tenantID, Slug: in.TenantSlug, AdminID: userID, RoleID: roleID}, nil
}

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

	var in BootstrapInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.TenantName == "" || in.TenantSlug == "" || in.AdminEmail == "" || len(in.AdminPassword) < 12 {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("tenant_name, tenant_slug, admin_email, and admin_password (12+ chars) are required"))
		return
	}

	result, err := h.svc.Bootstrap(r.Context(), in)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}
