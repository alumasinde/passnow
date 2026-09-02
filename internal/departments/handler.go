package departments

import (
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("all") != "true"
	items, err := h.repo.List(r.Context(), activeOnly)
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
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" || in.Code == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("name and code are required"))
		return
	}
	id, err := h.repo.Create(r.Context(), in.Name, in.Code)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	d, err := h.repo.ByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, ToDTO(d))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.repo.Update(r.Context(), id, in); err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	d, err := h.repo.ByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToDTO(d))
}
