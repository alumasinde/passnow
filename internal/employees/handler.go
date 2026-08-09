package employees

import (
	"errors"
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	p := httpx.ParsePagination(r)
	items, total, err := h.svc.List(r.Context(), tenant.ID, p)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	dtos := make([]DTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, h.svc.ToDTO(r.Context(), &items[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListEnvelope[DTO]{Items: dtos, Limit: p.Limit, Offset: p.Offset, Total: total})
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
	e, err := h.svc.Get(r.Context(), tenant.ID, id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.ToDTO(r.Context(), e))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in CreateInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.EmployeeNumber == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("employee_number is required"))
		return
	}
	e, err := h.svc.Create(r.Context(), tenant.ID, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, h.svc.ToDTO(r.Context(), e))
}

type updateRequest struct {
	EmployeeNumber *string `json:"employee_number"`
	DepartmentID   *int64  `json:"department_id"`
	Status         *string `json:"status"`
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
	var body updateRequest
	if !httpx.DecodeJSON(w, r, &body) {
		return
	}
	in := UpdateInput{EmployeeNumber: body.EmployeeNumber, DepartmentID: body.DepartmentID}
	if body.Status != nil {
		st := Status(*body.Status)
		in.Status = &st
	}
	e, err := h.svc.Update(r.Context(), tenant.ID, id, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.ToDTO(r.Context(), e))
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, httpx.ErrNotFound)
	case errors.Is(err, ErrDuplicateNumber):
		httpx.WriteError(w, httpx.AppError{Code: "duplicate_employee_number", Message: "employee_number is already in use", Status: http.StatusConflict})
	case errors.Is(err, ErrNameSourceConflict):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(err.Error()))
	case errors.Is(err, ErrNameRequired):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(err.Error()))
	case errors.Is(err, ErrInvalidUser):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(err.Error()))
	default:
		httpx.WriteError(w, httpx.ErrInternal)
	}
}
