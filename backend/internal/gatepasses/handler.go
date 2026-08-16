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

// --- Gatepass types (Platform Admin config) ------------------------------

func (h *Handler) ListTypes(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	activeOnly := r.URL.Query().Get("all") != "true"
	items, err := h.types.List(r.Context(), tenant.ID, activeOnly)
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
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
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
	id, err := h.types.Create(r.Context(), tenant.ID, in)
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	t, _ := h.types.ByID(r.Context(), tenant.ID, id)
	httpx.WriteJSON(w, http.StatusCreated, TypeToDTO(t))
}

func (h *Handler) UpdateType(w http.ResponseWriter, r *http.Request) {
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
	var in TypeInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.types.Update(r.Context(), tenant.ID, id, in); err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	t, _ := h.types.ByID(r.Context(), tenant.ID, id)
	httpx.WriteJSON(w, http.StatusOK, TypeToDTO(t))
}

// --- Gatepasses -----------------------------------------------------------

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
		dtos = append(dtos, ToDTO(&items[i]))
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

// --- QR ------------------------------------------------------------------

func (h *Handler) QRLookup(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	token := r.PathValue("token")
	dto, err := h.svc.QRLookup(r.Context(), tenant.ID, token)
	if err != nil {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, dto)
}

// QRImage renders the gatepass's QR code as a PNG, encoding a verification
// URL/token payload — this is what gets printed on the physical pass.
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
	w.Header().Set("Cache-Control", "no-store") // token is a bearer-style credential, don't let it linger in caches
	_, _ = w.Write(png)
}

// --- Approval work queue --------------------------------------------------

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

// MyPendingApprovals is the approver's work queue: every gatepass whose
// current step is theirs to act on right now. This is what makes approval
// usable day to day — without it, an approver has no way to discover what
// needs them short of being told a gatepass ID directly.
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

// --- helpers --------------------------------------------------------------

func parseIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func withDetails(r *http.Request, svc *Service, tenantID int64, g *Gatepass) DTO {
	dto := ToDTO(g)
	if items, err := svc.Items(r.Context(), tenantID, g.ID); err == nil {
		dto.Items = items
	}
	if steps, err := svc.ApprovalSteps(r.Context(), tenantID, g.ID); err == nil {
		for _, s := range steps {
			dto.Approvals = append(dto.Approvals, ApprovalStepDTO{
				StepOrder: s.StepOrder, Label: s.Label, Status: s.Status, ActedBy: s.ActedBy, Comments: s.Comments,
			})
		}
	}
	if movements, err := svc.Movements(r.Context(), tenantID, g.ID); err == nil {
		dto.Movements = movements
	}
	return dto
}

func writeMovementError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrMovementInvalid):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(err.Error()))
	case errors.Is(err, ErrMovementNotAllowed):
		httpx.WriteError(w, httpx.AppError{Code: "movement_not_allowed", Message: "this gatepass cannot be moved in its current state", Status: http.StatusConflict})
	case errors.Is(err, ErrReturnQuantityExceeded):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("return quantity exceeds the outstanding quantity"))
	case errors.Is(err, ErrReturnItemInvalid):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("return item does not belong to this gatepass"))
	case errors.Is(err, ErrFullReturnWithItems):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("full_return cannot be combined with item quantities"))
	case errors.Is(err, ErrInvalidType):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("gatepass_type_id is invalid or inactive"))
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, httpx.ErrNotFound)
	default:
		writeServiceError(w, err)
	}
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, httpx.ErrNotFound)
	case errors.Is(err, ErrInvalidType):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("gatepass_type_id is invalid or inactive"))
	case errors.Is(err, ErrInvalidDepartment):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("department_id is invalid or inactive"))
	case errors.Is(err, ErrInvalidVisitor):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("requester_visitor_id is invalid"))
	case errors.Is(err, ErrVisitorBlacklisted):
		httpx.WriteError(w, httpx.AppError{Code: "visitor_blacklisted", Message: "this visitor is blacklisted", Status: http.StatusForbidden})
	case errors.Is(err, ErrInvalidVisit):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("visit_id is invalid"))
	case errors.Is(err, ErrInvalidRequesterType):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("requester_type must be 'employee' or 'visitor'"))
	case errors.Is(err, ErrVisitorIDRequired):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("requester_visitor_id is required when requester_type is 'visitor'"))
	case errors.Is(err, ErrApprovalMisconfigured):
		httpx.WriteError(w, httpx.AppError{Code: "approval_misconfigured", Message: "this gatepass type requires approval but has no workflow configured — contact an admin", Status: http.StatusConflict})
	case errors.Is(err, ErrItemsRequired):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("this gatepass type requires at least one item"))
	case errors.Is(err, ErrReturnabilityPolicy):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("is_returnable conflicts with the gatepass type returnability policy"))
	case errors.Is(err, ErrInvalidItem):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("each item requires a name, quantity greater than zero, and a valid direction"))
	case errors.Is(err, ErrItemDirectionMismatch):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("item direction conflicts with the gatepass type direction"))
	case errors.Is(err, ErrReturnDateRequired):
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("expected_return_at is required when is_returnable is set"))
	case errors.Is(err, ErrNotEligibleApprover):
		httpx.WriteError(w, httpx.AppError{Code: "not_eligible_approver", Message: "you are not the eligible approver for this step", Status: http.StatusForbidden})
	case errors.Is(err, ErrInvalidTransition):
		httpx.WriteError(w, httpx.AppError{Code: "invalid_transition", Message: "this action is not valid for the gatepass's current status", Status: http.StatusConflict})
	case errors.Is(err, ErrDuplicatePassNumber):
		httpx.WriteError(w, httpx.ErrInternal)
	default:
		httpx.WriteError(w, httpx.ErrInternal)
	}
}
