package approvals

import (
	"errors"
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
	dtos := make([]WorkflowDTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, WorkflowDTO{ID: items[i].ID, Name: items[i].Name, Active: items[i].Active})
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
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
	wf, steps, err := h.repo.ByID(r.Context(), tenant.ID, id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	dto := WorkflowDTO{ID: wf.ID, Name: wf.Name, Active: wf.Active}
	for i := range steps {
		dto.Steps = append(dto.Steps, StepToDTO(&steps[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, dto)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in CreateWorkflowInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" || len(in.Steps) == 0 {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("name and at least one step are required"))
		return
	}
	id, err := h.repo.CreateWithSteps(r.Context(), tenant.ID, in)
	if err != nil {
		if errors.Is(err, ErrInvalidStepConfig) ||
			errors.Is(err, ErrApproverNotInTenant) ||
			errors.Is(err, ErrWorkflowNeedsRequiredStep) {
			httpx.WriteError(w, httpx.ErrValidation.WithMessage(err.Error()))
			return
		}
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	wf, steps, _ := h.repo.ByID(r.Context(), tenant.ID, id)
	dto := WorkflowDTO{ID: wf.ID, Name: wf.Name, Active: wf.Active}
	for i := range steps {
		dto.Steps = append(dto.Steps, StepToDTO(&steps[i]))
	}
	httpx.WriteJSON(w, http.StatusCreated, dto)
}
