package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
	"gatepass/internal/users"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	User         users.DTO `json:"user"`
	TenantSlug   string    `json:"tenant_slug"`
	MustChangePassword bool `json:"must_change_password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		httpx.WriteError(w, httpx.ErrBadRequestBody)
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Email == "" || req.Password == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("email and password are required"))
		return
	}

	pair, u, err := h.svc.Login(r.Context(), tenant.ID, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountLocked):
			httpx.WriteError(w, httpx.ErrAccountLocked)
		case errors.Is(err, ErrAccountDisabled):
			httpx.WriteError(w, httpx.ErrAccountDisabled)
		default:
			// Includes ErrInvalidCredentials and "no membership in tenant" —
			// always the same generic error (no user enumeration).
			httpx.WriteError(w, httpx.ErrInvalidCredentials)
		}
		return
	}


	httpx.WriteJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         users.ToDTO(u),
		TenantSlug:   tenant.Slug,
		MustChangePassword: u.MustChangePassword,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil || req.RefreshToken == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("refresh_token is required"))
		return
	}

	pair, err := h.svc.Refresh(r.Context(), tenant.ID, req.RefreshToken)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInvalidRefreshToken)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	if err := h.svc.Logout(r.Context(), claims.UserID); err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type changePasswordRequest struct { CurrentPassword string `json:"current_password"`; NewPassword string `json:"new_password"` }
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := reqctx.ClaimsFromContext(r.Context()); if !ok { httpx.WriteError(w,httpx.ErrAuthRequired); return }
	var req changePasswordRequest; if !httpx.DecodeJSON(w,r,&req){return}
	if req.CurrentPassword=="" || req.NewPassword=="" { httpx.WriteError(w,httpx.ErrValidation.WithMessage("current_password and new_password are required")); return }
	if err := h.svc.ChangePassword(r.Context(),claims.UserID,req.CurrentPassword,req.NewPassword); err != nil {
		if errors.Is(err,ErrInvalidCredentials){httpx.WriteError(w,httpx.ErrValidation.WithMessage("current password is incorrect"));return}
		if strings.Contains(err.Error(),"at least 8"){httpx.WriteError(w,httpx.ErrValidation.WithMessage("new password must be at least 8 characters"));return}
		httpx.WriteError(w,httpx.ErrInternal);return
	}
	httpx.WriteJSON(w,http.StatusOK,map[string]any{"message":"Password changed successfully"})
}



type updateProfileRequest struct { FirstName string `json:"first_name"`; LastName string `json:"last_name"` }
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := reqctx.ClaimsFromContext(r.Context()); if !ok { httpx.WriteError(w,httpx.ErrAuthRequired); return }
	u, err := h.svc.users.ByID(r.Context(),claims.UserID); if err != nil { httpx.WriteError(w,httpx.ErrAuthRequired); return }
	httpx.WriteJSON(w,http.StatusOK,users.ToDTO(u))
}
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := reqctx.ClaimsFromContext(r.Context()); if !ok { httpx.WriteError(w,httpx.ErrAuthRequired); return }
	var req updateProfileRequest; if !httpx.DecodeJSON(w,r,&req){return}
	u, err := h.svc.UpdateProfile(r.Context(),claims.UserID,req.FirstName,req.LastName); if err != nil {
		httpx.WriteError(w,httpx.ErrValidation.WithMessage("first_name and last_name are required")); return
	}
	httpx.WriteJSON(w,http.StatusOK,users.ToDTO(u))
}
