package platform

import (
    "errors"
    "net/http"

    "gatepass/internal/auth"
    "gatepass/internal/httpx"
    "gatepass/internal/reqctx"
    "gatepass/internal/users"
)

type AdminHandler struct {
    svc *AdminService
}

func NewAdminHandler(svc *AdminService) *AdminHandler { return &AdminHandler{svc: svc} }

type adminLoginRequest struct {
    Email string `json:"email"`
    Password string `json:"password"`
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req adminLoginRequest
    if !httpx.DecodeJSON(w, r, &req) { return }
    if req.Email == "" || req.Password == "" {
        httpx.WriteError(w, httpx.ErrValidation.WithMessage("email and password are required"))
        return
    }

    result, err := h.svc.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        httpx.WriteError(w, httpx.ErrInvalidCredentials)
        return
    }

    httpx.WriteJSON(w, http.StatusOK, map[string]any{
        "access_token": result.AccessToken,
        "expires_in": result.ExpiresIn,
        "role": result.Role,
        "user": users.ToDTO(result.User),
    })
}

func (h *AdminHandler) Me(w http.ResponseWriter, r *http.Request) {
    claims, ok := reqctx.ClaimsFromContext(r.Context())
    if !ok || claims.TenantID != 0 {
        httpx.WriteError(w, httpx.ErrAuthRequired)
        return
    }

    admin, err := h.svc.admins.ByUserID(r.Context(), claims.UserID)
    if err != nil {
        httpx.WriteError(w, httpx.ErrAuthRequired)
        return
    }

    httpx.WriteJSON(w, http.StatusOK, map[string]any{
        "user_id": claims.UserID,
        "role": admin.Role,
    })
}

var _ = errors.Is
var _ = auth.ErrTokenInvalid
