package roles

import (
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
	"gatepass/internal/tenants"
)

func tenantFromCtx(r *http.Request) (*tenants.Tenant, bool) {
	return reqctx.TenantFromContext(r.Context())
}

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// --- Permissions catalog ---------------------------------------------

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.AllPermissions(r.Context())
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	type dto struct {
		Code  string `json:"code"`
		Label string `json:"label"`
	}
	out := make([]dto, 0, len(items))
	for _, p := range items {
		out = append(out, dto{Code: p.Code, Label: p.Label})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// --- Roles -------------------------------------------------------------

type roleDTO struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsSystem bool   `json:"is_system"`
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenant, ok := tenantFromCtx(r)
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	items, err := h.repo.ListRoles(r.Context(), tenant.ID)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	out := make([]roleDTO, 0, len(items))
	for _, role := range items {
		out = append(out, roleDTO{ID: role.ID, Name: role.Name, IsSystem: role.IsSystem})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

type createRoleInput struct {
	Name string `json:"name"`
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	tenant, ok := tenantFromCtx(r)
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in createRoleInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("name is required"))
		return
	}
	id, err := h.repo.CreateRole(r.Context(), tenant.ID, in.Name, false)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	role, _ := h.repo.RoleByID(r.Context(), tenant.ID, id)
	httpx.WriteJSON(w, http.StatusCreated, roleDTO{ID: role.ID, Name: role.Name, IsSystem: role.IsSystem})
}

type setPermissionsInput struct {
	PermissionCodes []string `json:"permission_codes"`
}

// SetRolePermissions replaces a role's permission set — this is the
// screen where an admin ticks/unticks checkboxes for what a role (e.g.
// "HOD", "Security Manager") can do, and everything downstream
// (RequirePermission middleware, gatepass approval eligibility) reflects
// it on the next request, not after a token refresh.
func (h *Handler) SetRolePermissions(w http.ResponseWriter, r *http.Request) {
	tenant, ok := tenantFromCtx(r)
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	roleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	var in setPermissionsInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.repo.SetRolePermissions(r.Context(), tenant.ID, roleID, in.PermissionCodes); err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Users (tenant memberships) ----------------------------------------

type membershipDTO struct {
	MembershipID int64  `json:"membership_id"`
	UserID       int64  `json:"user_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	RoleID       int64  `json:"role_id"`
	RoleName     string `json:"role_name"`
	Status       string `json:"status"`
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenant, ok := tenantFromCtx(r)
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	items, err := h.repo.ListMemberships(r.Context(), tenant.ID)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	out := make([]membershipDTO, 0, len(items))
	for _, m := range items {
		out = append(out, membershipDTO{
			MembershipID: m.MembershipID, UserID: m.UserID, Email: m.Email,
			FirstName: m.FirstName, LastName: m.LastName, RoleID: m.RoleID,
			RoleName: m.RoleName, Status: string(m.Status),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

type updateMembershipInput struct {
	RoleID *int64  `json:"role_id"`
	Status *string `json:"status"` // "active" | "invited" | "disabled"
}

func (h *Handler) UpdateUserMembership(w http.ResponseWriter, r *http.Request) {
	tenant, ok := tenantFromCtx(r)
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	membershipID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	var in updateMembershipInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	var status *MembershipStatus
	if in.Status != nil {
		s := MembershipStatus(*in.Status)
		status = &s
	}
	if err := h.repo.UpdateMembership(r.Context(), tenant.ID, membershipID, in.RoleID, status); err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
