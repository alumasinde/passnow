package settings

import (
	"net/http"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type VisitorSettingsHandler struct {
	repo *Repository
}

func NewVisitorSettingsHandler(repo *Repository) *VisitorSettingsHandler {
	return &VisitorSettingsHandler{repo: repo}
}

type visitorSettingsDTO struct {
	AllowPreRegistration bool `json:"allow_pre_registration"`
}

func (h *VisitorSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	allowed := h.repo.GetBool(r.Context(), KeyVisitorsAllowPreRegistration, false)
	httpx.WriteJSON(w, http.StatusOK, visitorSettingsDTO{AllowPreRegistration: allowed})
}

func (h *VisitorSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	var in visitorSettingsDTO
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}

	if err := h.repo.Set(r.Context(), KeyVisitorsAllowPreRegistration, in.AllowPreRegistration, claims.UserID); err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, in)
}
