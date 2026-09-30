package navigation

import (
	"net/http"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	result, err := h.service.Build(r.Context(), claims.UserID, claims.RoleID)
	if err != nil {
		// A role changed after the current token was issued is an authorization
		// state change, not a server failure. Return 403 instead of a misleading 500.
		httpx.WriteError(w, httpx.ErrForbidden)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}
