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
	ErrInvalidType           = errors.New("gatepasses: gatepass_type_id not found or inactive")
	ErrInvalidDepartment     = errors.New("gatepasses: department_id not found or inactive")
	ErrInvalidVisitor        = errors.New("gatepasses: requester_visitor_id not found")
	ErrVisitorBlacklisted    = errors.New("gatepasses: visitor is blacklisted")
	ErrInvalidVisit          = errors.New("gatepasses: visit_id not found")
	ErrInvalidRequesterType  = errors.New("gatepasses: requester_type must be 'employee' or 'visitor'")
	ErrVisitorIDRequired     = errors.New("gatepasses: requester_visitor_id is required when requester_type is 'visitor'")
	ErrApprovalMisconfigured = errors.New("gatepasses: this gatepass type requires approval but has no workflow configured")
	ErrItemsRequired         = errors.New("gatepasses: this gatepass type requires at least one item")
	ErrReturnDateRequired    = errors.New("gatepasses: expected_return_at is required when is_returnable is set")
	ErrReturnabilityPolicy   = errors.New("gatepasses: requested returnability conflicts with the gatepass type policy")
	ErrInvalidItem           = errors.New("gatepasses: item name, quantity, and direction are invalid")
	ErrItemDirectionMismatch = errors.New("gatepasses: item direction conflicts with the gatepass type direction")
	ErrNotEligibleApprover   = errors.New("gatepasses: you are not the eligible approver for this step")
)

const (
	ActionGatepassCreated   = "GATEPASS_CREATED"
	ActionGatepassApproved  = "GATEPASS_APPROVED"
	ActionGatepassRejected  = "GATEPASS_REJECTED"
	ActionGatepassCancelled = "GATEPASS_CANCELLED"
	ActionGatepassIssued    = "GATEPASS_ISSUED"   // checked out at gate
	ActionGatepassVerified  = "GATEPASS_VERIFIED" // checked in at gate
	ActionGatepassExpired   = "GATEPASS_EXPIRED"
	ActionGatepassOverdue   = "GATEPASS_RETURN_OVERDUE"

	numberScope = "gatepass"
)

type Service struct {
	repo         *Repository
	movements    *MovementRepository
	types        *TypeRepository
	deptRepo     *departments.Repository
	visitorRepo  *visitors.Repository
	visitRepo    *visits.Repository
	workflowRepo *approvals.Repository
	roleRepo     *roles.Repository
	settingsRepo *settings.Repository
	auditRepo    *audit.Repository
	userRepo     *users.Repository
}

func NewService(
	repo *Repository, types *TypeRepository, deptRepo *departments.Repository,
	visitorRepo *visitors.Repository, visitRepo *visits.Repository,
	workflowRepo *approvals.Repository, roleRepo *roles.Repository,
	settingsRepo *settings.Repository, auditRepo *audit.Repository, userRepo *users.Repository,
) *Service {
	return &Service{
		repo: repo, movements: NewMovementRepository(repo.db), types: types, deptRepo: deptRepo, visitorRepo: visitorRepo, visitRepo: visitRepo,
		workflowRepo: workflowRepo, roleRepo: roleRepo, settingsRepo: settingsRepo, auditRepo: auditRepo,
		userRepo: userRepo,
	}
}

// Create validates every reference (type, department, visitor/visit),
// resolves the requester (employee path auto-fills from the authenticated
// caller — never from client input, so nobody can create a pass "as"
// someone else), resolves whether approval is actually required (type
// mandate wins over the client's opt-in checkbox — see CreateInput docs),
// and snapshots the workflow steps if so.
func (s *Service) Create(ctx context.Context, tenantID int64, in CreateInput, actorUserID int64) (*Gatepass, error) {
	gtype, err := s.types.ByID(ctx, tenantID, in.GatepassTypeID)
	if err != nil || !gtype.Active {
		return nil, ErrInvalidType
	}

	if in.DepartmentID != nil {
		d, err := s.deptRepo.ByID(ctx, *in.DepartmentID)
		if err != nil || !d.Active {
			return nil, ErrInvalidDepartment
		}
	}

	var requesterUserID, requesterVisitorID *int64
	switch RequesterType(in.RequesterType) {
	case RequesterEmployee:
		// Autopilot: the requester IS the authenticated caller. Never take
		// a requester_user_id from the client — that would let one user
		// create a gatepass in another user's name.
		requesterUserID = &actorUserID
	case RequesterVisitor:
		if in.RequesterVisitorID == nil {
			return nil, ErrVisitorIDRequired
		}
		visitor, err := s.visitorRepo.ByID(ctx, *in.RequesterVisitorID)
		if err != nil {
			return nil, ErrInvalidVisitor
		}
		if visitor.Status == visitors.StatusBlacklisted {
			return nil, ErrVisitorBlacklisted
		}
		requesterVisitorID = in.RequesterVisitorID
	default:
		return nil, ErrInvalidRequesterType
	}

	if in.VisitID != nil {
		if _, err := s.visitRepo.ByID(ctx, tenantID, *in.VisitID); err != nil {
			return nil, ErrInvalidVisit
		}
	}

	policy := gtype.ReturnabilityPolicy
	if policy == "" {
		policy = ReturnabilityOptional // backward-compatible with existing tenant data
	}
	isReturnable, err := ResolveReturnability(policy, in.IsReturnable, gtype.IsReturnableDefault)
	if err != nil {
		return nil, ErrReturnabilityPolicy
	}
	if isReturnable && in.ExpectedReturnAt == nil {
		return nil, ErrReturnDateRequired
	}
	if !isReturnable && in.ExpectedReturnAt != nil {
		return nil, ErrReturnabilityPolicy
	}

	if gtype.RequiresItems && len(in.Items) == 0 {
		return nil, ErrItemsRequired
	}
	for _, item := range in.Items {
		if item.Name == "" || item.Quantity <= 0 ||
			(item.Direction != "entering" && item.Direction != "leaving" && item.Direction != "returning") {
			return nil, ErrInvalidItem
		}
		switch gtype.Direction {
		case DirectionOut:
			if item.Direction == "entering" {
				return nil, ErrItemDirectionMismatch
			}
		case DirectionIn:
			if item.Direction != "entering" {
				return nil, ErrItemDirectionMismatch
			}
		case DirectionBoth:
			// both is intentionally permissive: each item can be entering or leaving.
		default:
			return nil, ErrInvalidType
		}
	}

	// --- Resolve approval requirement: type mandate beats client opt-in ---
	requiresApproval := gtype.RequiresApproval
	if !requiresApproval && in.RequestsApproval && gtype.WorkflowID != nil {
		requiresApproval = true // voluntary opt-in, only honored if a workflow actually exists
	}

	var steps []WorkflowStepSnapshot
	if requiresApproval {
		if gtype.WorkflowID == nil {
			return nil, ErrApprovalMisconfigured
		}
		_, tmplSteps, err := s.workflowRepo.ByID(ctx, tenantID, *gtype.WorkflowID)
		if err != nil {
			return nil, ErrApprovalMisconfigured
		}
		for _, ts := range tmplSteps {
			steps = append(steps, WorkflowStepSnapshot{
				StepOrder: ts.StepOrder, Label: ts.Label, ApproverType: string(ts.ApproverType),
				RoleID: ts.RoleID, UserID: ts.UserID, Required: ts.Required,
			})
		}
	}

	g := &Gatepass{
		TenantID: tenantID, GatepassTypeID: in.GatepassTypeID, DepartmentID: in.DepartmentID,
		RequesterType: RequesterType(in.RequesterType), RequesterUserID: requesterUserID, RequesterVisitorID: requesterVisitorID,
		VisitID: in.VisitID, Purpose: in.Purpose, IsReturnable: isReturnable, ExpectedReturnAt: in.ExpectedReturnAt,
		RequiresApproval: requiresApproval, WorkflowID: gtype.WorkflowID, CreatedBy: &actorUserID,
	}

	prefix := s.settingsRepo.GetString(ctx, settings.KeyGatepassNumberPrefix, "GP")
	useYear := s.settingsRepo.GetBool(ctx, settings.KeyGatepassNumberUseYear, true)
	period := ""
	if useYear {
		period = time.Now().UTC().Format("2006")
	}

	id, passNumber, err := s.repo.Create(ctx, CreateInputResolved{
		Gatepass: g, Items: in.Items, WorkflowSteps: steps,
		NumberScope: numberScope, NumberPrefix: prefix, NumberPeriod: period,
	})
	if err != nil {
		return nil, err
	}
	_ = passNumber

	s.audit(ctx, tenantID, actorUserID, ActionGatepassCreated, id, nil)
	return s.repo.ByID(ctx, tenantID, id)
}

func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Gatepass, error) {
	return s.repo.ByID(ctx, tenantID, id)
}

func (s *Service) List(ctx context.Context, tenantID int64, f ListFilter, p httpx.Pagination) ([]Gatepass, int, error) {
	return s.repo.List(ctx, tenantID, f, p)
}

func (s *Service) Items(ctx context.Context, tenantID, gatepassID int64) ([]ItemDTO, error) {
	return s.repo.items.ListForGatepass(ctx, tenantID, gatepassID)
}

func (s *Service) Movements(ctx context.Context, tenantID, gatepassID int64) ([]MovementDTO, error) {
	return s.movements.List(ctx, tenantID, gatepassID)
}

func (s *Service) CheckOutMovement(ctx context.Context, tenantID, id, actorUserID int64, in MovementInput) (*Gatepass, error) {
	if err := validateMovementInput(in, false); err != nil {
		return nil, err
	}
	g, err := s.repo.ByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	gtype, err := s.types.ByID(ctx, tenantID, g.GatepassTypeID)
	if err != nil || !gtype.Active {
		return nil, ErrInvalidType
	}
	result, err := s.movements.Checkout(ctx, tenantID, id, actorUserID, string(gtype.Direction), in)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionGatepassIssued, id, map[string]any{"movement": "checkout", "gate": in.GateName})
	return result, nil
}

func (s *Service) CheckInMovement(ctx context.Context, tenantID, id, actorUserID int64, in MovementInput) (*Gatepass, error) {
	if err := validateMovementInput(in, true); err != nil {
		return nil, err
	}
	g, err := s.repo.ByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	gtype, err := s.types.ByID(ctx, tenantID, g.GatepassTypeID)
	if err != nil || !gtype.Active {
		return nil, ErrInvalidType
	}
	result, err := s.movements.Checkin(ctx, tenantID, id, actorUserID, string(gtype.Direction), in)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionGatepassVerified, id, map[string]any{"movement": "checkin", "gate": in.GateName, "full_return": in.FullReturn})
	return result, nil
}

func (s *Service) ApprovalSteps(ctx context.Context, tenantID, gatepassID int64) ([]ApprovalStep, error) {
	return s.repo.ApprovalSteps(ctx, tenantID, gatepassID)
}

// Act processes an approve/reject on one step. Eligibility (does the
// caller actually hold the required role, or are they the specific named
// approver) is checked HERE, in addition to the generic "gatepasses.approve"
// permission already required to reach this handler — a Tenant Admin with
// gatepasses.approve should not be able to approve a step reserved for a
// different named user or a role they don't hold.
func (s *Service) Act(ctx context.Context, tenantID, gatepassID, stepID, actorUserID int64, approve bool, comments string) (*Gatepass, error) {
	steps, err := s.repo.ApprovalSteps(ctx, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}
	var target *ApprovalStep
	for i := range steps {
		if steps[i].ID == stepID {
			target = &steps[i]
			break
		}
	}
	if target == nil {
		return nil, ErrNotFound
	}

	eligible := false
	if target.ApproverType == string(approvals.ApproverSpecificUser) {
		if target.UserID != nil && *target.UserID == actorUserID {
			membership, err := s.roleRepo.MembershipFor(ctx, actorUserID)
			eligible = err == nil && membership.IsActive()
		}
	} else if target.RoleID != nil {
		membership, err := s.roleRepo.MembershipFor(ctx, actorUserID)
		eligible = err == nil && membership.IsActive() && membership.RoleID == *target.RoleID
	}
	if !eligible {
		return nil, ErrNotEligibleApprover
	}

	g, err := s.repo.ActOnApprovalStep(ctx, tenantID, gatepassID, stepID, actorUserID, approve, comments)
	if err != nil {
		return nil, err
	}

	action := ActionGatepassApproved
	if !approve {
		action = ActionGatepassRejected
	}
	s.audit(ctx, tenantID, actorUserID, action, gatepassID, map[string]any{"step_id": stepID, "comments": comments})
	return g, nil
}

// CheckOut/CheckIn need the gatepass type's direction, so the service
// loads the type before delegating to the repository's row-locked
// transition — the repository itself stays free of any type-lookup
// concerns.
func (s *Service) CheckOut(ctx context.Context, tenantID, id, actorUserID int64) (*Gatepass, error) {
	g, err := s.repo.ByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	gtype, err := s.types.ByID(ctx, tenantID, g.GatepassTypeID)
	if err != nil {
		return nil, ErrInvalidType
	}
	result, err := s.repo.CheckOut(ctx, tenantID, id, actorUserID, string(gtype.Direction))
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionGatepassIssued, id, nil)
	return result, nil
}

func (s *Service) CheckIn(ctx context.Context, tenantID, id, actorUserID int64) (*Gatepass, error) {
	g, err := s.repo.ByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	gtype, err := s.types.ByID(ctx, tenantID, g.GatepassTypeID)
	if err != nil {
		return nil, ErrInvalidType
	}
	result, err := s.repo.CheckIn(ctx, tenantID, id, actorUserID, string(gtype.Direction))
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionGatepassVerified, id, nil)
	return result, nil
}

func (s *Service) Cancel(ctx context.Context, tenantID, id, actorUserID int64, reason string) (*Gatepass, error) {
	g, err := s.repo.Cancel(ctx, tenantID, id, actorUserID, reason)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionGatepassCancelled, id, map[string]any{"reason": reason})
	return g, nil
}

// QRLookup resolves a scanned token to the minimal safe display info,
// scoped strictly to the requesting tenant (a token is globally unique,
// but a scan at tenant A's gate must never resolve tenant B's gatepass).
func (s *Service) QRLookup(ctx context.Context, tenantID int64, token string) (*QRDTO, error) {
	g, err := s.repo.ByQRToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if g.TenantID != tenantID {
		return nil, ErrNotFound
	}

	requesterName := ""
	if g.RequesterUserID != nil {
		if u, err := s.userRepo.ByID(ctx, *g.RequesterUserID); err == nil {
			requesterName = u.FullName()
		}
	}
	if g.RequesterVisitorID != nil {
		if v, err := s.visitorRepo.ByID(ctx, *g.RequesterVisitorID); err == nil {
			requesterName = v.FullName()
		}
	}

	return &QRDTO{
		GatepassID: g.ID, PassNumber: g.PassNumber, Status: string(g.Status),
		Requester: requesterName, Purpose: g.Purpose, IsReturnable: g.IsReturnable,
	}, nil
}

func (s *Service) QRToken(ctx context.Context, tenantID, id int64) (string, error) {
	g, err := s.repo.ByID(ctx, tenantID, id)
	if err != nil {
		return "", err
	}
	return g.QRToken, nil
}

// PendingForApprover resolves the caller's current role membership, then
// returns their personal approval queue — see Repository.PendingForApprover.
func (s *Service) PendingForApprover(ctx context.Context, tenantID, actorUserID int64) ([]PendingApprovalItem, error) {
	membership, err := s.roleRepo.MembershipFor(ctx, actorUserID)
	if err != nil || !membership.IsActive() {
		return nil, nil // not an active member here -> empty queue, not an error
	}
	return s.repo.PendingForApprover(ctx, tenantID, actorUserID, membership.RoleID)
}

func (s *Service) audit(ctx context.Context, tenantID, actorUserID int64, action string, entityID int64, metadata map[string]any) {
	_ = s.auditRepo.Record(ctx, s.auditRepo.DB(), audit.Entry{
		TenantID: &tenantID, ActorUserID: &actorUserID, Action: action,
		EntityType: "gatepass", EntityID: &entityID, Metadata: metadata,
	})
}
