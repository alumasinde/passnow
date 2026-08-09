package departments

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	activeOnly := r.URL.Query().Get("all") != "true"
	items, err := h.repo.List(r.Context(), tenant.ID, activeOnly)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	dtos := make([]DTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, ToDTO(&items[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" || in.Code == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("name and code are required"))
		return
	}
	id, err := h.repo.Create(r.Context(), tenant.ID, in.Name, in.Code)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	d, _ := h.repo.ByID(r.Context(), tenant.ID, id)
	httpx.WriteJSON(w, http.StatusCreated, ToDTO(d))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.repo.Update(r.Context(), tenant.ID, id, in); err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	d, _ := h.repo.ByID(r.Context(), tenant.ID, id)
	httpx.WriteJSON(w, http.StatusOK, ToDTO(d))
}
