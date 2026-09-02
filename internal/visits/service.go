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

func (s *Service) Create(ctx context.Context, tenantID int64, in CreateInput, actorUserID int64) (*Visit, error) {
	visitor, err := s.visitorRepo.ByID(ctx, in.VisitorID)
	if err != nil {
		return nil, ErrVisitorNotFound
	}
	if visitor.Status == visitors.StatusBlacklisted {
		return nil, ErrVisitorBlacklisted
	}

	if in.VisitTypeID != nil {
		vt, err := s.visitTypes.ByID(ctx, *in.VisitTypeID)
		if err != nil || !vt.Active {
			return nil, ErrInvalidVisitType
		}
	}
	if in.DepartmentID != nil {
		d, err := s.deptRepo.ByID(ctx, *in.DepartmentID)
		if err != nil || !d.Active {
			return nil, ErrInvalidDepartment
		}
	}

	v := &Visit{
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

	return s.repo.ByID(ctx, id)
}

func (s *Service) CheckIn(ctx context.Context, tenantID, id, actorUserID int64) (*Visit, error) {
	v, err := s.repo.CheckIn(ctx, id, actorUserID)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionVisitCheckedIn, id, map[string]any{"badge_number": v.BadgeNumber})
	return v, nil
}

func (s *Service) CheckOut(ctx context.Context, tenantID, id, actorUserID int64) (*Visit, error) {
	v, err := s.repo.CheckOut(ctx, id, actorUserID)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionVisitCheckedOut, id, nil)
	return v, nil
}

func (s *Service) Cancel(ctx context.Context, tenantID, id, actorUserID int64, reason string) (*Visit, error) {
	v, err := s.repo.Cancel(ctx, id, actorUserID, reason)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, actorUserID, ActionVisitCancelled, id, map[string]any{"reason": reason})
	return v, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Visit, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) List(ctx context.Context, tenantID int64, f ListFilter, p httpx.Pagination) ([]Visit, int, error) {
	return s.repo.List(ctx, f, p)
}

func (s *Service) BadgeByToken(ctx context.Context, tenantID int64, token string) (*Visit, *visitors.Visitor, error) {
	v, err := s.repo.ByBadgeToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	visitor, err := s.visitorRepo.ByID(ctx, v.VisitorID)
	if err != nil {
		return nil, nil, err
	}
	return v, visitor, nil
}

func (s *Service) audit(ctx context.Context, tenantID, actorUserID int64, action string, entityID int64, metadata map[string]any) {
	_ = s.auditRepo.Record(ctx, s.auditRepo.DB(), audit.Entry{
		ActorUserID: &actorUserID, Action: action,
		EntityType: "visit", EntityID: &entityID, Metadata: metadata,
	})
}
