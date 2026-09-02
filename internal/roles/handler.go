package roles

import (
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func requireTenant(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return false
	}
	return true
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w, r) { return }
	items, err := h.repo.AllPermissions(r.Context())
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	type dto struct { Code string `json:"code"`; Label string `json:"label"` }
	out := make([]dto, 0, len(items))
	for _, p := range items { out = append(out, dto{Code:p.Code, Label:p.Label}) }
	httpx.WriteJSON(w, http.StatusOK, out)
}

type roleDTO struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	IsSystem bool `json:"is_system"`
}

func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w, r) { return }
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 { httpx.WriteError(w, httpx.ErrNotFound); return }
	role, err := h.repo.RoleByID(r.Context(), id)
	if err != nil { httpx.WriteError(w, httpx.ErrNotFound); return }
	codes, err := h.repo.PermissionCodesForRole(r.Context(), id)
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	selected := make([]string, 0, len(codes)); for code := range codes { selected = append(selected, code) }
	httpx.WriteJSON(w, http.StatusOK, struct { ID int64 `json:"id"`; Name string `json:"name"`; IsSystem bool `json:"is_system"`; PermissionCodes []string `json:"permission_codes"` }{role.ID, role.Name, role.IsSystem, selected})
}

type updateRoleInput struct { Name string `json:"name"` }
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w, r) { return }
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64); if err != nil || id < 1 { httpx.WriteError(w,httpx.ErrNotFound); return }
	var in updateRoleInput; if !httpx.DecodeJSON(w,r,&in) { return }
	if in.Name == "" { httpx.WriteError(w,httpx.ErrValidation.WithMessage("name is required")); return }
	if err := h.repo.UpdateRole(r.Context(),id,in.Name); err != nil { httpx.WriteError(w,httpx.ErrValidation.WithMessage(err.Error())); return }
	httpx.WriteJSON(w,http.StatusOK,map[string]any{"id":id,"name":in.Name})
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w, r) { return }
	items, err := h.repo.ListRoles(r.Context())
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	out := make([]roleDTO,0,len(items))
	for _, role := range items { out=append(out, roleDTO{ID:role.ID,Name:role.Name,IsSystem:role.IsSystem}) }
	httpx.WriteJSON(w,http.StatusOK,out)
}

type createRoleInput struct { Name string `json:"name"` }

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w,r) { return }
	var in createRoleInput
	if !httpx.DecodeJSON(w,r,&in){return}
	if in.Name=="" { httpx.WriteError(w,httpx.ErrValidation.WithMessage("name is required")); return }
	id,err:=h.repo.CreateRole(r.Context(),in.Name,false)
	if err!=nil { httpx.WriteError(w,httpx.ErrInternal); return }
	role,err:=h.repo.RoleByID(r.Context(),id)
	if err!=nil { httpx.WriteError(w,httpx.ErrInternal); return }
	httpx.WriteJSON(w,http.StatusCreated,roleDTO{ID:role.ID,Name:role.Name,IsSystem:role.IsSystem})
}

type setPermissionsInput struct { PermissionCodes []string `json:"permission_codes"` }

func (h *Handler) SetRolePermissions(w http.ResponseWriter,r *http.Request){
	if !requireTenant(w,r){return}
	roleID,err:=strconv.ParseInt(r.PathValue("id"),10,64)
	if err!=nil {httpx.WriteError(w,httpx.ErrNotFound);return}
	var in setPermissionsInput
	if !httpx.DecodeJSON(w,r,&in){return}
	if err:=h.repo.SetRolePermissions(r.Context(),roleID,in.PermissionCodes);err!=nil {httpx.WriteError(w,httpx.ErrNotFound);return}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"role_id": roleID,
		"permission_codes": in.PermissionCodes,
	})
}

type membershipDTO struct {
	MembershipID int64 `json:"membership_id"`
	UserID int64 `json:"user_id"`
	Email string `json:"email"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	RoleID int64 `json:"role_id"`
	RoleName string `json:"role_name"`
	Status string `json:"status"`
}

func (h *Handler) ListUsers(w http.ResponseWriter,r *http.Request){
	if !requireTenant(w,r){return}
	items,err:=h.repo.ListMemberships(r.Context())
	if err!=nil {httpx.WriteError(w,httpx.ErrInternal);return}
	out:=make([]membershipDTO,0,len(items))
	for _,m:=range items{out=append(out,membershipDTO{MembershipID:m.MembershipID,UserID:m.UserID,Email:m.Email,FirstName:m.FirstName,LastName:m.LastName,RoleID:m.RoleID,RoleName:m.RoleName,Status:string(m.Status)})}
	httpx.WriteJSON(w,http.StatusOK,out)
}

type updateMembershipInput struct {
	RoleID *int64 `json:"role_id"`
	Status *string `json:"status"`
}

func (h *Handler) UpdateUserMembership(w http.ResponseWriter,r *http.Request){
	if !requireTenant(w,r){return}
	membershipID,err:=strconv.ParseInt(r.PathValue("id"),10,64)
	if err!=nil {httpx.WriteError(w,httpx.ErrNotFound);return}
	var in updateMembershipInput
	if !httpx.DecodeJSON(w,r,&in){return}
	var status *MembershipStatus
	if in.Status!=nil {s:=MembershipStatus(*in.Status);status=&s}
	if err:=h.repo.UpdateMembership(r.Context(),membershipID,in.RoleID,status);err!=nil {httpx.WriteError(w,httpx.ErrNotFound);return}
	w.WriteHeader(http.StatusNoContent)
}
