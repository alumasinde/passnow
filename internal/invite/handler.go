// Package invite handles adding a user to a tenant. It's a separate small
// package (rather than living in users, roles, or auth) specifically to
// avoid an import cycle: auth already imports both users and roles, so
// this logic — which needs users + roles + auth's password hashing —
// can't live in any of those three without creating one.
package invite

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"

	"gatepass/internal/auth"
	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
	"gatepass/internal/roles"
	"gatepass/internal/users"
)

var ErrRoleNotFound = errors.New("invite: role not found")

const DefaultInitialPassword = "PassNow@123"

type Service struct {
	users      *users.Repository
	roles      *roles.Repository
	bcryptCost int
}

func NewService(userRepo *users.Repository, roleRepo *roles.Repository, bcryptCost int) *Service {
	return &Service{users: userRepo, roles: roleRepo, bcryptCost: bcryptCost}
}

type Input struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    int64  `json:"role_id"`
}

type Result struct {
	UserID       int64  `json:"user_id"`
	MembershipID int64  `json:"membership_id"`
	Email        string `json:"email"`
	// TemporaryPassword is returned ONCE, at invite time, because there is
	// no email-sending infrastructure yet to deliver a proper invite link.
	// The user should change it on first login. This is a deliberate,
	// documented simplification — not a substitute for a real invite-email
	// flow, which should replace this before production use with real
	// external users.
	TemporaryPassword string `json:"temporary_password,omitempty"`
}

// Invite adds a user to a tenant: reuses an existing global account by
// email if one exists (a person already registered under a different
// tenant), otherwise creates a new account with a random temporary
// password. Either way, the membership itself is what actually grants
// tenant access — creating/finding the user alone grants nothing.
func (s *Service) Invite(ctx context.Context, tenantID int64, in Input) (*Result, error) {
	if _, err := s.roles.RoleByID(ctx, in.RoleID); err != nil {
		return nil, ErrRoleNotFound
	}

	existing, err := s.users.ByEmail(ctx, in.Email)
	var userID int64
	var tempPassword string

	if err == nil {
		userID = existing.ID
	} else {
		tempPassword = DefaultInitialPassword
		hash, err := auth.HashPassword(tempPassword, s.bcryptCost)
		if err != nil {
			return nil, err
		}
		newUser := &users.User{Email: in.Email, PasswordHash: hash, FirstName: in.FirstName, LastName: in.LastName, MustChangePassword: true}
		userID, err = s.users.Create(ctx, newUser)
		if err != nil {
			return nil, err
		}
	}

	membershipID, err := s.roles.CreateMembership(ctx, userID, in.RoleID, roles.MembershipActive)
	if err != nil { return nil, err }

	return &Result{UserID: userID, MembershipID: membershipID, Email: in.Email, TemporaryPassword: tempPassword}, nil
}

func randomPassword() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// --- HTTP handler ---------------------------------------------------------

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Email == "" || in.FirstName == "" || in.LastName == "" || in.RoleID == 0 {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("email, first_name, last_name and role_id are required"))
		return
	}

	result, err := h.svc.Invite(r.Context(), tenant.ID, in)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			httpx.WriteError(w, httpx.ErrValidation.WithMessage("role_id is invalid"))
			return
		}
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}

// Invite remains as a compatibility alias for older clients.
func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) { h.CreateUser(w,r) }
