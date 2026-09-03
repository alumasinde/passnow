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
	ErrInvalidEntrySource = errors.New("visits: invalid entry source")
	ErrInvalidExpectedTime = errors.New("visits: expected departure must be after expected arrival")
)

const (
	ActionVisitCreated    = "VISIT_CREATED"
	ActionVisitCheckedIn  = "VISIT_CHECKED_IN"
	ActionVisitCheckedOut = "VISIT_CHECKED_OUT"
	ActionVisitCancelled  = "VISIT_CANCELLED"
	ActionVisitQRIssued = "VISIT_QR_ISSUED"
	ActionVisitQRInvalidated = "VISIT_QR_INVALIDATED"
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
	if in.EntrySource == "" { in.EntrySource = EntrySourcePreRegistered }
	if in.EntrySource != EntrySourceWalkIn && in.EntrySource != EntrySourcePreRegistered { return nil, ErrInvalidEntrySource }
	if in.ExpectedTime != nil && in.ExpectedDepartureAt != nil && !in.ExpectedDepartureAt.After(*in.ExpectedTime) { return nil, ErrInvalidExpectedTime }
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
		EntrySource: in.EntrySource,
		VisitTypeID:  in.VisitTypeID,
		DepartmentID: in.DepartmentID,
		HostName:     in.HostName,
		Purpose:      in.Purpose,
		ExpectedTime: in.ExpectedTime,
		ExpectedDepartureAt: in.ExpectedDepartureAt,
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
	if in.EntrySource == EntrySourcePreRegistered {
		if _, err := s.IssueQR(ctx, tenantID, id, actorUserID); err != nil { return nil, err }
	}
	return s.repo.ByID(ctx, id)
}

func (s *Service) CheckIn(ctx context.Context, tenantID, id, actorUserID int64) (*Visit, error) {
	current, err := s.repo.ByID(ctx, id)
	if err != nil { return nil, err }
	visitor, err := s.visitorRepo.ByID(ctx, current.VisitorID)
	if err != nil { return nil, ErrVisitorNotFound }
	if visitor.Status == visitors.StatusBlacklisted { return nil, ErrVisitorBlacklisted }
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


// ToDTO enriches a visit with tenant-local display names. Missing optional
// lookup records do not make an otherwise valid visit unreadable.
func (s *Service) ToDTO(ctx context.Context, v *Visit) DTO {
	d := ToDTO(v)
	if visitor, err := s.visitorRepo.ByID(ctx, v.VisitorID); err == nil {
		d.VisitorName = visitor.FullName()
	}
	if v.VisitTypeID != nil {
		if visitType, err := s.visitTypes.ByID(ctx, *v.VisitTypeID); err == nil {
			d.VisitTypeName = visitType.Name
		}
	}
	if v.DepartmentID != nil {
		if department, err := s.deptRepo.ByID(ctx, *v.DepartmentID); err == nil {
			d.DepartmentName = department.Name
		}
	}
	return d
}

func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Visit, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) List(ctx context.Context, tenantID int64, f ListFilter, p httpx.Pagination) ([]Visit, int, error) {
	return s.repo.List(ctx, f, p)
}

func (s *Service) IssueQR(ctx context.Context,tenantID,id,actorUserID int64)(*Visit,error){v,err:=s.repo.ByID(ctx,id);if err!=nil{return nil,err};if v.Status==StatusCancelled||v.Status==StatusCheckedOut||v.Status==StatusExpired{return nil,ErrInvalidEntrySource};token,err:=randomToken();if err!=nil{return nil,err};if err=s.repo.IssueQR(ctx,id,token);err!=nil{return nil,err};s.audit(ctx,tenantID,actorUserID,ActionVisitQRIssued,id,nil);return s.repo.ByID(ctx,id)}
func (s *Service) InvalidateQR(ctx context.Context,tenantID,id,actorUserID int64) error { if _,err:=s.repo.ByID(ctx,id);err!=nil{return err};if err:=s.repo.InvalidateQR(ctx,id);err!=nil{return err};s.audit(ctx,tenantID,actorUserID,ActionVisitQRInvalidated,id,nil);return nil }
func (s *Service) QRLookup(ctx context.Context,tenantID int64,token string)(*Visit,*visitors.Visitor,error){v,err:=s.repo.ByQRToken(ctx,token);if err!=nil{return nil,nil,err};visitor,err:=s.visitorRepo.ByID(ctx,v.VisitorID);if err!=nil{return nil,nil,err};return v,visitor,nil}

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


func (s *Service) UserDepartment(ctx context.Context, userID int64) (*int64, error) {
	return s.repo.UserDepartment(ctx, userID)
}
