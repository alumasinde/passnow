package employees

import (
	"context"
	"errors"
	"gatepass/internal/rbac"
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

func requireTenant(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return false
	}
	return true
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w, r) {
		return
	}
	p := httpx.ParsePagination(r)
	var departmentID *int64
	if d, ok := rbac.DecisionFromContext(r.Context()); ok && d.Scope == rbac.ScopeDepartment {
		var err error
		departmentID, err = h.svc.UserDepartment(r.Context(), claims.UserID)
		if err != nil || departmentID == nil { httpx.WriteError(w,httpx.ErrForbidden); return }
	}
	items, total, err := h.svc.ListScoped(r.Context(), p, departmentID)
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
	if !requireTenant(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	e, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	if !h.canAccessEmployee(r.Context(), e) { httpx.WriteError(w,httpx.ErrForbidden); return }
	httpx.WriteJSON(w, http.StatusOK, h.svc.ToDTO(r.Context(), e))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w, r) {
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
	e, err := h.svc.Create(r.Context(), in)
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
	if !requireTenant(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	existing, err := h.svc.Get(r.Context(), id); if err != nil { httpx.WriteError(w,httpx.ErrNotFound); return }
	if !h.canAccessEmployee(r.Context(), existing) { httpx.WriteError(w,httpx.ErrForbidden); return }
	var body updateRequest
	if !httpx.DecodeJSON(w, r, &body) {
		return
	}
	in := UpdateInput{EmployeeNumber: body.EmployeeNumber, DepartmentID: body.DepartmentID}
	if body.Status != nil {
		st := Status(*body.Status)
		in.Status = &st
	}
	e, err := h.svc.Update(r.Context(), id, in)
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


func (h *Handler) canAccessEmployee(ctx context.Context, e *Employee) bool {
	d, ok := rbac.DecisionFromContext(ctx); if !ok || d.Scope == rbac.ScopeNone || d.Scope == rbac.ScopeAll { return true }
	if d.Scope == rbac.ScopeOwn { return false }
	if d.Scope != rbac.ScopeDepartment || e.DepartmentID == nil { return false }
	claims, ok := reqctx.ClaimsFromContext(ctx); if !ok { return false }
	dept, err := h.svc.UserDepartment(ctx, claims.UserID)
	return err == nil && dept != nil && *dept == *e.DepartmentID
}
