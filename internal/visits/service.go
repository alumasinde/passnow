package visits

import (
	"context"
	"errors"

	"gatepass/internal/audit"
	"gatepass/internal/departments"
	"gatepass/internal/httpx"
	"gatepass/internal/visitors"
)

var (
	ErrVisitorNotFound    = errors.New("visits: visitor not found")
	ErrVisitorBlacklisted = errors.New("visits: visitor is blacklisted and cannot be scheduled")
	ErrInvalidVisitType   = errors.New("visits: visit type not found or inactive")
	ErrInvalidDepartment  = errors.New("visits: department not found or inactive")
)

// Audit action codes for this module.
const (
	ActionVisitCreated    = "VISIT_CREATED"
	ActionVisitCheckedIn  = "VISIT_CHECKED_IN"
	ActionVisitCheckedOut = "VISIT_CHECKED_OUT"
	ActionVisitCancelled  = "VISIT_CANCELLED"
)

type Service struct {
	repo        *Repository
	visitorRepo *visitors.Repository
	visitTypes  *VisitTypeRepository
	deptRepo    *departments.Repository
	auditRepo   *audit.Repository
}

func NewService(repo *Repository, visitorRepo *visitors.Repository, visitTypes *VisitTypeRepository, deptRepo *departments.Repository, auditRepo *audit.Repository) *Service {
	return &Service{repo: repo, visitorRepo: visitorRepo, visitTypes: visitTypes, deptRepo: deptRepo, auditRepo: auditRepo}
}

// Create validates the visitor (must exist, must not be blacklisted — a
// security boundary, not just a UX nicety) plus any referenced visit
// type/department, then schedules the visit. If CheckInNow is set, it
// immediately performs the check-in transition (badge generation
// included) as one logical operation from the caller's point of view.
func (s *Service) Create(ctx context.Context, tenantID int64, in CreateInput, actorUserID int64) (*Visit, error) {
	visitor, err := s.visitorRepo.ByID(ctx, tenantID, in.VisitorID)
	if err != nil {
		return nil, ErrVisitorNotFound
	}
	if visitor.Status == visitors.StatusBlacklisted {
		return nil, ErrVisitorBlacklisted
	}

	if in.VisitTypeID != nil {
		vt, err := s.visitTypes.ByID(ctx, tenantID, *in.VisitTypeID)
		if err != nil || !vt.Active {
			return nil, ErrInvalidVisitType
		}
	}
	if in.DepartmentID != nil {
		d, err := s.deptRepo.ByID(ctx, tenantID, *in.DepartmentID)
		if err != nil || !d.Active {
			return nil, ErrInvalidDepartment
		}
	}

	v := &Visit{
		TenantID:     tenantID,
		VisitorID:    in.VisitorID,
		VisitTypeID:  in.VisitTypeID,
		DepartmentID: in.DepartmentID,
		HostName:     in.HostName,
		Purpose:      in.Purpose,
		ExpectedTime: in.ExpectedTime,
		Status:       StatusScheduled,
		CreatedBy:    &actorUserID,
	}

	id, err := s.repo.Create(ctx, v)
	if err != nil {
		return nil, err
	}
	v.ID = id

	s.audit(ctx, tenantID, actorUserID, ActionVisitCreated, id, nil)

	if in.CheckInNow {
		return s.CheckIn(ctx, tenantID, id, actorUserID)
	}

	return s.repo.ByID(ctx, tenantID, id)
}

func (s *Service) CheckIn(ctx context.Context, tenantID, id, actorUserID int64) (*Visit, error) {
	v, err := s.repo.CheckIn(ctx, tenantID, id, actorUserID)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionVisitCheckedIn, id, map[string]any{"badge_number": v.BadgeNumber})
	return v, nil
}

func (s *Service) CheckOut(ctx context.Context, tenantID, id, actorUserID int64) (*Visit, error) {
	v, err := s.repo.CheckOut(ctx, tenantID, id, actorUserID)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionVisitCheckedOut, id, nil)
	return v, nil
}

func (s *Service) Cancel(ctx context.Context, tenantID, id, actorUserID int64, reason string) (*Visit, error) {
	v, err := s.repo.Cancel(ctx, tenantID, id, actorUserID, reason)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionVisitCancelled, id, map[string]any{"reason": reason})
	return v, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Visit, error) {
	return s.repo.ByID(ctx, tenantID, id)
}

func (s *Service) List(ctx context.Context, tenantID int64, f ListFilter, p httpx.Pagination) ([]Visit, int, error) {
	return s.repo.List(ctx, tenantID, f, p)
}

// BadgeByToken resolves a scanned badge token to its visit, but ONLY if
// the visit belongs to the tenant resolved for the current request — a
// badge token is unique platform-wide, but a scan at tenant A's gate must
// never succeed against a badge issued by tenant B.
func (s *Service) BadgeByToken(ctx context.Context, tenantID int64, token string) (*Visit, *visitors.Visitor, error) {
	v, err := s.repo.ByBadgeToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	if v.TenantID != tenantID {
		return nil, nil, ErrNotFound
	}
	visitor, err := s.visitorRepo.ByID(ctx, tenantID, v.VisitorID)
	if err != nil {
		return nil, nil, err
	}
	return v, visitor, nil
}

func (s *Service) audit(ctx context.Context, tenantID, actorUserID int64, action string, entityID int64, metadata map[string]any) {
	_ = s.auditRepo.Record(ctx, s.auditRepo.DB(), audit.Entry{
		TenantID: &tenantID, ActorUserID: &actorUserID, Action: action,
		EntityType: "visit", EntityID: &entityID, Metadata: metadata,
	})
}
