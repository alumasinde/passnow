package visits

import (
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
)

type VisitTypeHandler struct { repo *VisitTypeRepository }
func NewVisitTypeHandler(repo *VisitTypeRepository) *VisitTypeHandler { return &VisitTypeHandler{repo: repo} }

func (h *VisitTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("all") != "true"
	items, err := h.repo.List(r.Context(), activeOnly)
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	dtos := make([]VisitTypeDTO, 0, len(items))
	for i := range items { dtos = append(dtos, VisitTypeToDTO(&items[i])) }
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

func (h *VisitTypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 { httpx.WriteError(w, httpx.ErrNotFound); return }
	t, err := h.repo.ByID(r.Context(), id)
	if err != nil { httpx.WriteError(w, httpx.ErrNotFound); return }
	httpx.WriteJSON(w, http.StatusOK, VisitTypeToDTO(t))
}

func (h *VisitTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in VisitTypeInput
	if !httpx.DecodeJSON(w, r, &in) { return }
	if in.Name == "" || in.Code == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("name and code are required")); return
	}
	id, err := h.repo.Create(r.Context(), in)
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	t, err := h.repo.ByID(r.Context(), id)
	if err != nil { httpx.WriteError(w,httpx.ErrInternal); return }
	httpx.WriteJSON(w, http.StatusCreated, VisitTypeToDTO(t))
}

func (h *VisitTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil { httpx.WriteError(w,httpx.ErrNotFound); return }
	var in VisitTypeInput
	if !httpx.DecodeJSON(w, r, &in) { return }
	if err := h.repo.Update(r.Context(), id, in); err != nil { httpx.WriteError(w,httpx.ErrNotFound); return }
	t, err := h.repo.ByID(r.Context(), id)
	if err != nil { httpx.WriteError(w,httpx.ErrInternal); return }
	httpx.WriteJSON(w, http.StatusOK, VisitTypeToDTO(t))
}
