package visitors

import (
	"context"
	"errors"

	"gatepass/internal/audit"
	"gatepass/internal/httpx"
	"gatepass/internal/settings"
)

var (
	ErrPreRegistrationDisabled = errors.New("visitors: pre-registration is not enabled for this tenant")
	ErrIDNumberRequired        = errors.New("visitors: id_number is required for this id type")
	ErrInvalidIDType           = errors.New("visitors: id type not found or inactive")
	ErrInvalidCompany          = errors.New("visitors: company not found or inactive")
)

type Service struct {
	repo         *Repository
	idTypes      *IDTypeRepository
	companies    *CompanyRepository
	settingsRepo *settings.Repository
	auditRepo    *audit.Repository
}

func NewService(repo *Repository, idTypes *IDTypeRepository, companies *CompanyRepository, settingsRepo *settings.Repository, auditRepo *audit.Repository) *Service {
	return &Service{repo: repo, idTypes: idTypes, companies: companies, settingsRepo: settingsRepo, auditRepo: auditRepo}
}

// Create registers a visitor. Whether the result is "walk_in" or
// "pre_registered" is decided HERE, not trusted from the client: the
// client can only ASK for pre-registration (WantsPreRegistration); the
// tenant's Platform-Admin-controlled setting decides whether that request
// is honored. If pre-registration is off, the visitor is still created —
// just as a normal walk_in — rather than the whole request failing, unless
// the caller specifically needs pre-registration to succeed (handler
// decides that UX call; the service exposes the distinction via the
// returned error).
func (s *Service) Create(ctx context.Context, tenantID int64, in CreateInput, actorUserID int64) (*Visitor, error) {
	idType, err := s.idTypes.ByID(ctx, tenantID, in.IDTypeID)
	if err != nil {
		return nil, ErrInvalidIDType
	}
	if !idType.Active {
		return nil, ErrInvalidIDType
	}
	if idType.RequiresNumber && (in.IDNumber == nil || *in.IDNumber == "") {
		return nil, ErrIDNumberRequired
	}

	if in.CompanyID != nil {
		c, err := s.companies.ByID(ctx, tenantID, *in.CompanyID)
		if err != nil || !c.Active {
			return nil, ErrInvalidCompany
		}
	}

	source := SourceWalkIn
	if in.WantsPreRegistration {
		allowed := s.settingsRepo.GetBool(ctx, tenantID, settings.KeyVisitorsAllowPreRegistration, false)
		if !allowed {
			return nil, ErrPreRegistrationDisabled
		}
		source = SourcePreRegistered
	}

	v := &Visitor{
		TenantID:  tenantID,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		IDTypeID:  in.IDTypeID,
		IDNumber:  in.IDNumber,
		CompanyID: in.CompanyID,
		Phone:     in.Phone,
		Email:     in.Email,
		Notes:     in.Notes,
		Source:    source,
		CreatedBy: &actorUserID,
	}

	id, err := s.repo.Create(ctx, v)
	if err != nil {
		return nil, err
	}
	v.ID = id

	_ = s.auditRepo.Record(ctx, s.auditRepo.DB(), audit.Entry{
		TenantID:    &tenantID,
		ActorUserID: &actorUserID,
		Action:      audit.ActionVisitorCreated,
		EntityType:  "visitor",
		EntityID:    &id,
		Metadata:    map[string]any{"source": string(source)},
	})

	return v, nil
}

func (s *Service) Update(ctx context.Context, tenantID, id int64, in UpdateInput, actorUserID int64) (*Visitor, error) {
	if in.IDTypeID != nil {
		idType, err := s.idTypes.ByID(ctx, tenantID, *in.IDTypeID)
		if err != nil || !idType.Active {
			return nil, ErrInvalidIDType
		}
	}
	if in.CompanyID != nil {
		c, err := s.companies.ByID(ctx, tenantID, *in.CompanyID)
		if err != nil || !c.Active {
			return nil, ErrInvalidCompany
		}
	}

	v, err := s.repo.Update(ctx, tenantID, id, in, actorUserID)
	if err != nil {
		return nil, err
	}

	_ = s.auditRepo.Record(ctx, s.auditRepo.DB(), audit.Entry{
		TenantID:    &tenantID,
		ActorUserID: &actorUserID,
		Action:      audit.ActionVisitorUpdated,
		EntityType:  "visitor",
		EntityID:    &id,
	})

	return v, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Visitor, error) {
	return s.repo.ByID(ctx, tenantID, id)
}

func (s *Service) List(ctx context.Context, tenantID int64, f ListFilter, p httpx.Pagination) ([]Visitor, int, error) {
	return s.repo.List(ctx, tenantID, f, p)
}

func (s *Service) SetBlacklist(ctx context.Context, tenantID, id int64, in BlacklistInput, actorUserID int64) error {
	return s.repo.SetBlacklist(ctx, tenantID, id, in.Blacklisted, in.Reason, actorUserID)
}
