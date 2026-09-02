package auth

import (
	"encoding/json"
	"errors"
	"log"
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
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("========== LOGIN DEBUG START ==========")

	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		log.Printf("LOGIN FAILED: tenant missing from request context path=%q", r.URL.Path)
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	log.Printf("LOGIN tenant resolved: id=%d slug=%q status=%q", tenant.ID, tenant.Slug, tenant.Status)

	var req loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		log.Printf("LOGIN FAILED: request body decode error: %v", err)
		httpx.WriteError(w, httpx.ErrBadRequestBody)
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	log.Printf("LOGIN email received/normalized: %q", req.Email)

	if req.Email == "" || req.Password == "" {
		log.Printf("LOGIN FAILED: missing email or password")
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("email and password are required"))
		return
	}

	pair, u, err := h.svc.Login(r.Context(), tenant.ID, req.Email, req.Password)
	if err != nil {
		log.Printf("LOGIN SERVICE FAILED: tenant=%q tenant_id=%d email=%q error=%v", tenant.Slug, tenant.ID, req.Email, err)
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

	log.Printf("LOGIN SUCCESS: tenant=%q tenant_id=%d user_id=%d email=%q role_response_ready=true", tenant.Slug, tenant.ID, u.ID, u.Email)
	log.Println("========== LOGIN DEBUG END ==========")

	httpx.WriteJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         users.ToDTO(u),
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
