package gatepasses

import (
	"errors"
	"net/http"
	"strconv"

	qrcode "github.com/skip2/go-qrcode"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct {
	svc   *Service
	types *TypeRepository
}

func NewHandler(svc *Service, types *TypeRepository) *Handler {
	return &Handler{svc: svc, types: types}
}

// --- Gatepass types (tenant-local configuration) -------------------------

func (h *Handler) ListTypes(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	activeOnly := r.URL.Query().Get("all") != "true"
	items, err := h.types.List(r.Context(), activeOnly)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	dtos := make([]TypeDTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, TypeToDTO(&items[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

func (h *Handler) CreateType(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in TypeInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" || in.Code == "" || (in.Direction != "in" && in.Direction != "out" && in.Direction != "both") {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("name, code, and a valid direction (in|out|both) are required"))
		return
	}
	if in.ReturnabilityPolicy != nil && *in.ReturnabilityPolicy != "optional" && *in.ReturnabilityPolicy != "required" && *in.ReturnabilityPolicy != "not_allowed" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("returnability_policy must be optional, required, or not_allowed"))
		return
	}
	if in.RequiresApproval != nil && *in.RequiresApproval && in.WorkflowID == nil {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("workflow_id is required when requires_approval is true"))
		return
	}
	id, err := h.types.Create(r.Context(), in)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	t, err := h.types.ByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, TypeToDTO(t))
}

func (h *Handler) UpdateType(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	var in TypeInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.types.Update(r.Context(), id, in); err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	t, err := h.types.ByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, TypeToDTO(t))
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
	if in.GatepassTypeID == 0 || in.RequesterType == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("gatepass_type_id and requester_type are required"))
		return
	}
	g, err := h.svc.Create(r.Context(), tenant.ID, in, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, withDetails(r, h.svc, tenant.ID, g))
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
	g, err := h.svc.Get(r.Context(), tenant.ID, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, withDetails(r, h.svc, tenant.ID, g))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var f ListFilter
	if s := r.URL.Query().Get("status"); s != "" {
		st := Status(s)
		f.Status = &st
	}
	p := httpx.ParsePagination(r)
	items, total, err := h.svc.List(r.Context(), tenant.ID, f, p)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	dtos := make([]DTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, h.svc.Details(r.Context(), tenant.ID, &items[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListEnvelope[DTO]{Items: dtos, Limit: p.Limit, Offset: p.Offset, Total: total})
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) { h.act(w, r, true) }
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request)  { h.act(w, r, false) }

func (h *Handler) act(w http.ResponseWriter, r *http.Request, approve bool) {
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
	gatepassID, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	stepID, err := strconv.ParseInt(r.PathValue("stepId"), 10, 64)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	var in ApprovalActionInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	g, err := h.svc.Act(r.Context(), tenant.ID, gatepassID, stepID, claims.UserID, approve, in.Comments)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, withDetails(r, h.svc, tenant.ID, g))
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
	var in MovementInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	g, err := h.svc.CheckOutMovement(r.Context(), tenant.ID, id, claims.UserID, in)
	if err != nil {
		writeMovementError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, withDetails(r, h.svc, tenant.ID, g))
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
	var in MovementInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	g, err := h.svc.CheckInMovement(r.Context(), tenant.ID, id, claims.UserID, in)
	if err != nil {
		writeMovementError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, withDetails(r, h.svc, tenant.ID, g))
}

func (h *Handler) Movements(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.svc.Movements(r.Context(), tenant.ID, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
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
	g, err := h.svc.Cancel(r.Context(), tenant.ID, id, claims.UserID, in.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, withDetails(r, h.svc, tenant.ID, g))
}

func (h *Handler) QRLookup(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	token := r.PathValue("token")
	dto, err := h.svc.QRLookupDetails(r.Context(), tenant.ID, token)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, dto)
}


// QRTokenImage renders a QR image from the opaque token capability. The token
// itself is random and validated against the tenant database before rendering.
func (h *Handler) QRTokenImage(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok { httpx.WriteError(w, httpx.ErrAuthRequired); return }
	token := r.PathValue("token")
	if _, err := h.svc.QRLookup(r.Context(), tenant.ID, token); err != nil {
		httpx.WriteError(w, httpx.ErrNotFound); return
	}
	png, err := qrcode.Encode(token, qrcode.Medium, 320)
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(png)
}

func (h *Handler) QRImage(w http.ResponseWriter, r *http.Request) {
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
	token, err := h.svc.QRToken(r.Context(), tenant.ID, id)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	png, err := qrcode.Encode(token, qrcode.Medium, 320)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(png)
}

type pendingApprovalDTO struct {
	GatepassID    int64   `json:"gatepass_id"`
	PassNumber    string  `json:"pass_number"`
	StepID        int64   `json:"step_id"`
	StepOrder     int     `json:"step_order"`
	StepLabel     string  `json:"step_label"`
	RequesterType string  `json:"requester_type"`
	Purpose       *string `json:"purpose"`
	CreatedAt     string  `json:"created_at"`
}

func (h *Handler) MyPendingApprovals(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.svc.PendingForApprover(r.Context(), tenant.ID, claims.UserID)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	out := make([]pendingApprovalDTO, 0, len(items))
	for _, it := range items {
		out = append(out, pendingApprovalDTO{
			GatepassID: it.GatepassID, PassNumber: it.PassNumber, StepID: it.StepID,
			StepOrder: it.StepOrder, StepLabel: it.StepLabel, RequesterType: it.RequesterType,
			Purpose: it.Purpose, CreatedAt: it.CreatedAt,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func parseIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidType), errors.Is(err, ErrInvalidDepartment), errors.Is(err, ErrInvalidVisitor),
		errors.Is(err, ErrInvalidVisit), errors.Is(err, ErrInvalidRequesterType), errors.Is(err, ErrVisitorIDRequired),
		errors.Is(err, ErrApprovalMisconfigured), errors.Is(err, ErrItemsRequired), errors.Is(err, ErrReturnDateRequired),
		errors.Is(err, ErrReturnabilityPolicy), errors.Is(err, ErrInvalidItem), errors.Is(err, ErrItemDirectionMismatch):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(err.Error()))
	case errors.Is(err, ErrNotEligibleApprover):
		httpx.WriteError(w, httpx.ErrForbidden)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, httpx.ErrNotFound)
	case errors.Is(err, ErrInvalidTransition):
		httpx.WriteError(w, httpx.ErrConflict.WithMessage(err.Error()))
	default:
		httpx.WriteError(w, httpx.ErrInternal)
	}
}

func writeMovementError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrMovementInvalid), errors.Is(err, ErrMovementNotAllowed), errors.Is(err, ErrReturnItemInvalid), errors.Is(err, ErrReturnQuantityExceeded):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(err.Error()))
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, httpx.ErrNotFound)
	default:
		writeServiceError(w, err)
	}
}

func withDetails(r *http.Request, svc *Service, tenantID int64, g *Gatepass) DTO {
	return svc.Details(r.Context(), tenantID, g)
}
