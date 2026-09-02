package dashboard

import (
	"net/http"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct {
	repo *Repository
	service *Service
}

func NewHandler(repo *Repository, service *Service) *Handler {
	return &Handler{repo: repo, service: service}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	_, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	summary, err := h.repo.Summary(r.Context())
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, summary)
}


func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	result, err := h.service.BuildDashboard(r.Context(), claims.UserID, claims.RoleID)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}
