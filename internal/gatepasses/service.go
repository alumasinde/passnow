package gatepasses

import (
	"context"
	"errors"
	"time"

	"gatepass/internal/approvals"
	"gatepass/internal/audit"
	"gatepass/internal/departments"
	"gatepass/internal/httpx"
	"gatepass/internal/roles"
	"gatepass/internal/settings"
	"gatepass/internal/users"
	"gatepass/internal/visitors"
	"gatepass/internal/visits"
)

var (
	ErrInvalidType = errors.New("gatepasses: gatepass_type_id not found or inactive")
	ErrInvalidDepartment = errors.New("gatepasses: department_id not found or inactive")
	ErrInvalidVisitor = errors.New("gatepasses: requester_visitor_id not found")
	ErrVisitorBlacklisted = errors.New("gatepasses: visitor is blacklisted")
	ErrInvalidVisit = errors.New("gatepasses: visit_id not found")
	ErrInvalidRequesterType = errors.New("gatepasses: requester_type must be 'employee' or 'visitor'")
	ErrVisitorIDRequired = errors.New("gatepasses: requester_visitor_id is required when requester_type is 'visitor'")
	ErrApprovalMisconfigured = errors.New("gatepasses: this gatepass type requires approval but has no workflow configured")
	ErrItemsRequired = errors.New("gatepasses: this gatepass type requires at least one item")
	ErrReturnDateRequired = errors.New("gatepasses: expected_return_at is required when is_returnable is set")
	ErrReturnabilityPolicy = errors.New("gatepasses: requested returnability conflicts with the gatepass type policy")
	ErrInvalidItem = errors.New("gatepasses: item name, quantity, and direction are invalid")
	ErrItemDirectionMismatch = errors.New("gatepasses: item direction conflicts with the gatepass type direction")
	ErrNotEligibleApprover = errors.New("gatepasses: you are not the eligible approver for this step")
)

const (
	ActionGatepassCreated = "GATEPASS_CREATED"
	ActionGatepassApproved = "GATEPASS_APPROVED"
	ActionGatepassRejected = "GATEPASS_REJECTED"
	ActionGatepassCancelled = "GATEPASS_CANCELLED"
	ActionGatepassIssued = "GATEPASS_ISSUED"
	ActionGatepassVerified = "GATEPASS_VERIFIED"
	ActionGatepassExpired = "GATEPASS_EXPIRED"
	ActionGatepassOverdue = "GATEPASS_RETURN_OVERDUE"
	numberScope = "gatepass"
)

type Service struct {
	repo *Repository
	movements *MovementRepository
	types *TypeRepository
	deptRepo *departments.Repository
	visitorRepo *visitors.Repository
	visitRepo *visits.Repository
	workflowRepo *approvals.Repository
	roleRepo *roles.Repository
	settingsRepo *settings.Repository
	auditRepo *audit.Repository
	userRepo *users.Repository
}

func NewService(repo *Repository, types *TypeRepository, deptRepo *departments.Repository, visitorRepo *visitors.Repository, visitRepo *visits.Repository, workflowRepo *approvals.Repository, roleRepo *roles.Repository, settingsRepo *settings.Repository, auditRepo *audit.Repository, userRepo *users.Repository) *Service {
	return &Service{repo: repo, movements: NewMovementRepository(repo.db), types: types, deptRepo: deptRepo, visitorRepo: visitorRepo, visitRepo: visitRepo, workflowRepo: workflowRepo, roleRepo: roleRepo, settingsRepo: settingsRepo, auditRepo: auditRepo, userRepo: userRepo}
}

func (s *Service) Create(ctx context.Context, tenantID int64, in CreateInput, actorUserID int64) (*Gatepass, error) {
	gtype, err := s.types.ByID(ctx, in.GatepassTypeID)
	if err != nil || !gtype.Active { return nil, ErrInvalidType }
	if in.DepartmentID != nil { d, err := s.deptRepo.ByID(ctx, *in.DepartmentID); if err != nil || !d.Active { return nil, ErrInvalidDepartment } }
	var requesterUserID, requesterVisitorID *int64
	switch RequesterType(in.RequesterType) {
	case RequesterEmployee: requesterUserID = &actorUserID
	case RequesterVisitor:
		if in.RequesterVisitorID == nil { return nil, ErrVisitorIDRequired }
		visitor, err := s.visitorRepo.ByID(ctx, *in.RequesterVisitorID); if err != nil { return nil, ErrInvalidVisitor }
		if visitor.Status == visitors.StatusBlacklisted { return nil, ErrVisitorBlacklisted }
		requesterVisitorID = in.RequesterVisitorID
	default: return nil, ErrInvalidRequesterType
	}
	if in.VisitID != nil { if _, err := s.visitRepo.ByID(ctx, *in.VisitID); err != nil { return nil, ErrInvalidVisit } }
	policy := gtype.ReturnabilityPolicy; if policy == "" { policy = ReturnabilityOptional }
	isReturnable, err := ResolveReturnability(policy, in.IsReturnable, gtype.IsReturnableDefault); if err != nil { return nil, ErrReturnabilityPolicy }
	if isReturnable && in.ExpectedReturnAt == nil { return nil, ErrReturnDateRequired }
	if !isReturnable && in.ExpectedReturnAt != nil { return nil, ErrReturnabilityPolicy }
	if gtype.RequiresItems && len(in.Items) == 0 { return nil, ErrItemsRequired }
	for _, item := range in.Items {
		if item.Name == "" || item.Quantity <= 0 || (item.Direction != "entering" && item.Direction != "leaving" && item.Direction != "returning") { return nil, ErrInvalidItem }
		switch gtype.Direction { case DirectionOut: if item.Direction == "entering" { return nil, ErrItemDirectionMismatch }; case DirectionIn: if item.Direction != "entering" { return nil, ErrItemDirectionMismatch }; case DirectionBoth: default: return nil, ErrInvalidType }
	}
	requiresApproval := gtype.RequiresApproval; if !requiresApproval && in.RequestsApproval && gtype.WorkflowID != nil { requiresApproval = true }
	var steps []WorkflowStepSnapshot
	if requiresApproval {
		if gtype.WorkflowID == nil { return nil, ErrApprovalMisconfigured }
		_, tmplSteps, err := s.workflowRepo.ByID(ctx, tenantID, *gtype.WorkflowID); if err != nil { return nil, ErrApprovalMisconfigured }
		for _, ts := range tmplSteps { steps = append(steps, WorkflowStepSnapshot{StepOrder: ts.StepOrder, Label: ts.Label, ApproverType: string(ts.ApproverType), RoleID: ts.RoleID, UserID: ts.UserID, Required: ts.Required}) }
	}
	g := &Gatepass{GatepassTypeID: in.GatepassTypeID, DepartmentID: in.DepartmentID, RequesterType: RequesterType(in.RequesterType), RequesterUserID: requesterUserID, RequesterVisitorID: requesterVisitorID, VisitID: in.VisitID, Purpose: in.Purpose, IsReturnable: isReturnable, ExpectedReturnAt: in.ExpectedReturnAt, RequiresApproval: requiresApproval, WorkflowID: gtype.WorkflowID, CreatedBy: &actorUserID}
	prefix := s.settingsRepo.GetString(ctx, settings.KeyGatepassNumberPrefix, "GP")
	useYear := s.settingsRepo.GetBool(ctx, settings.KeyGatepassNumberUseYear, true)
	period := ""; if useYear { period = time.Now().UTC().Format("2006") }
	id, _, err := s.repo.Create(ctx, CreateInputResolved{Gatepass: g, Items: in.Items, WorkflowSteps: steps, NumberScope: numberScope, NumberPrefix: prefix, NumberPeriod: period}); if err != nil { return nil, err }
	s.audit(ctx, tenantID, actorUserID, ActionGatepassCreated, id, nil)
	return s.repo.ByID(ctx, id)
}

func (s *Service) Details(ctx context.Context, tenantID int64, g *Gatepass) DTO {
	d := ToDTO(g)
	if t, err := s.types.ByID(ctx, g.GatepassTypeID); err == nil {
		d.GatepassTypeName = t.Name
		d.Direction = string(t.Direction)
	}
	if g.DepartmentID != nil {
		if dept, err := s.deptRepo.ByID(ctx, *g.DepartmentID); err == nil { d.DepartmentName = dept.Name }
	}
	if g.RequesterType == RequesterVisitor && g.RequesterVisitorID != nil {
		if v, err := s.visitorRepo.ByID(ctx, *g.RequesterVisitorID); err == nil {
			d.RequesterName = v.FullName(); d.SubjectName = d.RequesterName
		}
	}
	if g.RequesterType == RequesterEmployee && g.RequesterUserID != nil {
		if u, err := s.userRepo.ByID(ctx, *g.RequesterUserID); err == nil {
			d.RequesterName = u.FullName(); d.SubjectName = d.RequesterName
		}
	}
	if d.SubjectName == "" { d.SubjectName = d.RequesterName }
	if items, err := s.Items(ctx, tenantID, g.ID); err == nil { d.Items = items }
	if steps, err := s.ApprovalSteps(ctx, tenantID, g.ID); err == nil {
		d.Approvals = make([]ApprovalStepDTO, 0, len(steps))
		for _, step := range steps {
			d.Approvals = append(d.Approvals, ApprovalStepDTO{StepOrder: step.StepOrder, Label: step.Label, Status: step.Status, ActedBy: step.ActedBy, Comments: step.Comments})
		}
	}
	return d
}

func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Gatepass, error) { return s.repo.ByID(ctx, id) }
func (s *Service) List(ctx context.Context, tenantID int64, f ListFilter, p httpx.Pagination) ([]Gatepass, int, error) { return s.repo.List(ctx, f, p) }
func (s *Service) Items(ctx context.Context, tenantID, gatepassID int64) ([]ItemDTO, error) { return s.repo.items.ListForGatepass(ctx, gatepassID) }
func (s *Service) Movements(ctx context.Context, tenantID, gatepassID int64) ([]MovementDTO, error) { return s.movements.List(ctx, gatepassID) }
func (s *Service) CheckOutMovement(ctx context.Context, tenantID, id, actorUserID int64, in MovementInput) (*Gatepass, error) {
	if err := validateMovementInput(in, false); err != nil { return nil, err }
	g, err := s.repo.ByID(ctx, id); if err != nil { return nil, err }
	gtype, err := s.types.ByID(ctx, g.GatepassTypeID); if err != nil || !gtype.Active { return nil, ErrInvalidType }
	result, err := s.movements.Checkout(ctx, id, actorUserID, string(gtype.Direction), in); if err != nil { return nil, err }
	s.audit(ctx, tenantID, actorUserID, ActionGatepassIssued, id, map[string]any{"movement":"checkout", "gate":in.GateName}); return result, nil
}
func (s *Service) CheckInMovement(ctx context.Context, tenantID, id, actorUserID int64, in MovementInput) (*Gatepass, error) {
	if err := validateMovementInput(in, true); err != nil { return nil, err }
	g, err := s.repo.ByID(ctx, id); if err != nil { return nil, err }
	gtype, err := s.types.ByID(ctx, g.GatepassTypeID); if err != nil || !gtype.Active { return nil, ErrInvalidType }
	result, err := s.movements.Checkin(ctx, id, actorUserID, string(gtype.Direction), in); if err != nil { return nil, err }
	s.audit(ctx, tenantID, actorUserID, ActionGatepassVerified, id, map[string]any{"movement":"checkin", "gate":in.GateName, "full_return":in.FullReturn}); return result, nil
}
func (s *Service) ApprovalSteps(ctx context.Context, tenantID, gatepassID int64) ([]ApprovalStep, error) { return s.repo.ApprovalSteps(ctx, gatepassID) }
func (s *Service) Act(ctx context.Context, tenantID, gatepassID, stepID, actorUserID int64, approve bool, comments string) (*Gatepass, error) {
	steps, err := s.repo.ApprovalSteps(ctx, gatepassID); if err != nil { return nil, err }
	var target *ApprovalStep; for i := range steps { if steps[i].ID == stepID { target = &steps[i]; break } }
	if target == nil { return nil, ErrNotFound }
	eligible := false
	if target.ApproverType == string(approvals.ApproverSpecificUser) { if target.UserID != nil && *target.UserID == actorUserID { membership, e := s.roleRepo.MembershipFor(ctx, actorUserID); eligible = e == nil && membership.IsActive() } } else if target.RoleID != nil { membership, e := s.roleRepo.MembershipFor(ctx, actorUserID); eligible = e == nil && membership.IsActive() && membership.RoleID == *target.RoleID }
	if !eligible { return nil, ErrNotEligibleApprover }
	g, err := s.repo.ActOnApprovalStep(ctx, gatepassID, stepID, actorUserID, approve, comments); if err != nil { return nil, err }
	action := ActionGatepassApproved; if !approve { action = ActionGatepassRejected }; s.audit(ctx, tenantID, actorUserID, action, gatepassID, map[string]any{"step_id":stepID, "comments":comments}); return g, nil
}
func (s *Service) CheckOut(ctx context.Context, tenantID, id, actorUserID int64) (*Gatepass, error) { g, err := s.repo.ByID(ctx, id); if err != nil { return nil, err }; gtype, err := s.types.ByID(ctx, g.GatepassTypeID); if err != nil { return nil, ErrInvalidType }; result, err := s.repo.CheckOut(ctx, id, actorUserID, string(gtype.Direction)); if err != nil { return nil, err }; s.audit(ctx, tenantID, actorUserID, ActionGatepassIssued, id, nil); return result, nil }
func (s *Service) CheckIn(ctx context.Context, tenantID, id, actorUserID int64) (*Gatepass, error) { g, err := s.repo.ByID(ctx, id); if err != nil { return nil, err }; gtype, err := s.types.ByID(ctx, g.GatepassTypeID); if err != nil { return nil, ErrInvalidType }; result, err := s.repo.CheckIn(ctx, id, actorUserID, string(gtype.Direction)); if err != nil { return nil, err }; s.audit(ctx, tenantID, actorUserID, ActionGatepassVerified, id, nil); return result, nil }
func (s *Service) Cancel(ctx context.Context, tenantID, id, actorUserID int64, reason string) (*Gatepass, error) { g, err := s.repo.Cancel(ctx, id, actorUserID, reason); if err != nil { return nil, err }; s.audit(ctx, tenantID, actorUserID, ActionGatepassCancelled, id, map[string]any{"reason":reason}); return g, nil }
func (s *Service) QRLookup(ctx context.Context, tenantID int64, token string) (*QRDTO, error) { g, err := s.repo.ByQRToken(ctx, token); if err != nil { return nil, err }; requesterName := ""; if g.RequesterUserID != nil { if u, e := s.userRepo.ByID(ctx, *g.RequesterUserID); e == nil { requesterName = u.FullName() } }; if g.RequesterVisitorID != nil { if v, e := s.visitorRepo.ByID(ctx, *g.RequesterVisitorID); e == nil { requesterName = v.FullName() } }; return &QRDTO{GatepassID:g.ID, PassNumber:g.PassNumber, Status:string(g.Status), Requester:requesterName, Purpose:g.Purpose, IsReturnable:g.IsReturnable}, nil }
func (s *Service) QRLookupDetails(ctx context.Context, tenantID int64, token string) (DTO, error) {
	g, err := s.repo.ByQRToken(ctx, token)
	if err != nil { return DTO{}, err }
	return s.Details(ctx, tenantID, g), nil
}
func (s *Service) QRToken(ctx context.Context, tenantID, id int64) (string, error) { g, err := s.repo.ByID(ctx, id); if err != nil { return "", err }; return g.QRToken, nil }
func (s *Service) PendingForApprover(ctx context.Context, tenantID, actorUserID int64) ([]PendingApprovalItem, error) { membership, err := s.roleRepo.MembershipFor(ctx, actorUserID); if err != nil || !membership.IsActive() { return nil, nil }; return s.repo.PendingForApprover(ctx, actorUserID, membership.RoleID) }
func (s *Service) audit(ctx context.Context, tenantID, actorUserID int64, action string, entityID int64, metadata map[string]any) { _ = s.auditRepo.Record(ctx, s.auditRepo.DB(), audit.Entry{ ActorUserID:&actorUserID, Action:action, EntityType:"gatepass", EntityID:&entityID, Metadata:metadata}) }


func (s *Service) UserDepartment(ctx context.Context, userID int64) (*int64, error) {
	return s.repo.UserDepartment(ctx, userID)
}
