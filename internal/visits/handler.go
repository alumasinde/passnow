package visits

import (
	"errors"
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
	"gatepass/internal/visitors"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	var in CreateInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.VisitorID == 0 {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("visitor_id is required"))
		return
	}

	v, err := h.svc.Create(r.Context(), tenant.ID, in, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, h.svc.ToDTO(r.Context(), v))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	v, err := h.svc.Get(r.Context(), tenant.ID, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.ToDTO(r.Context(), v))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var f ListFilter
	q := r.URL.Query()
	if s := q.Get("status"); s != "" {
		st := Status(s)
		f.Status = &st
	}
	if vid := q.Get("visitor_id"); vid != "" {
		if n, err := strconv.ParseInt(vid, 10, 64); err == nil {
			f.VisitorID = &n
		}
	}
	if q.Get("q") != "" { f.Search = q.Get("q") }
	if q.Get("date") != "" { f.Date = q.Get("date") }

	p := httpx.ParsePagination(r)
	items, total, err := h.svc.List(r.Context(), tenant.ID, f, p)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	dtos := make([]DTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, h.svc.ToDTO(r.Context(), &items[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items": dtos, "limit": p.Limit, "offset": p.Offset,
		"meta": map[string]any{"total": total, "statuses": visitStatusOptions()},
	})
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	v, err := h.svc.CheckIn(r.Context(), tenant.ID, id, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToDTO(v))
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	v, err := h.svc.CheckOut(r.Context(), tenant.ID, id, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToDTO(v))
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	var in CancelInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Reason == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("reason is required"))
		return
	}
	v, err := h.svc.Cancel(r.Context(), tenant.ID, id, claims.UserID, in.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToDTO(v))
}

// BadgeLookup is what a security guard's scanner calls with the token read
// from the badge's QR code — fast verification without exposing the
// visitor's ID number or other sensitive fields.
func (h *Handler) BadgeLookup(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	token := r.PathValue("token")
	if token == "" {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	v, visitor, err := h.svc.BadgeByToken(r.Context(), tenant.ID, token)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, BadgeDTO{
		VisitID: v.ID, VisitorName: visitor.FullName(), DepartmentID: v.DepartmentID,
		HostName: v.HostName, Status: string(v.Status), CheckedInAt: v.CheckedInAt, BadgeNumber: v.BadgeNumber,
	})
}

func visitStatusOptions() []map[string]string {
	return []map[string]string{
		{"value":string(StatusScheduled),"label":"Scheduled"},
		{"value":string(StatusExpected),"label":"Expected"},
		{"value":string(StatusCheckedIn),"label":"Checked in"},
		{"value":string(StatusCheckedOut),"label":"Checked out"},
		{"value":string(StatusCancelled),"label":"Cancelled"},
		{"value":string(StatusNoShow),"label":"No show"},
		{"value":string(StatusExpired),"label":"Expired"},
	}
}

func parseIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, httpx.ErrNotFound)
	case errors.Is(err, ErrVisitorNotFound):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("visitor_id is invalid"))
	case errors.Is(err, ErrVisitorBlacklisted):
		httpx.WriteError(w, httpx.AppError{Code: "visitor_blacklisted", Message: "this visitor is blacklisted and cannot be scheduled", Status: http.StatusForbidden})
	case errors.Is(err, ErrInvalidVisitType):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("visit_type_id is invalid or inactive"))
	case errors.Is(err, ErrInvalidDepartment):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("department_id is invalid or inactive"))
	case errors.Is(err, ErrInvalidTransition):
		httpx.WriteError(w, httpx.AppError{Code: "invalid_transition", Message: "this action is not valid for the visit's current status", Status: http.StatusConflict})
	case errors.Is(err, visitors.ErrNotFound):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("visitor_id is invalid"))
	default:
		httpx.WriteError(w, httpx.ErrInternal)
	}
}
