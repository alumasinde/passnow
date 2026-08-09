package employees

import (
	"context"
	"errors"

	"gatepass/internal/httpx"
	"gatepass/internal/roles"
	"gatepass/internal/users"
)

var (
	ErrNameSourceConflict = errors.New("employees: provide either user_id, or first_name+last_name, not both")
	ErrNameRequired       = errors.New("employees: first_name and last_name are required when user_id is not set")
	ErrInvalidUser        = errors.New("employees: user_id must belong to a user with an active membership in this tenant")
)

type Service struct {
	repo     *Repository
	userRepo *users.Repository
	roleRepo *roles.Repository
}

func NewService(repo *Repository, userRepo *users.Repository, roleRepo *roles.Repository) *Service {
	return &Service{repo: repo, userRepo: userRepo, roleRepo: roleRepo}
}

func (s *Service) Create(ctx context.Context, tenantID int64, in CreateInput) (*Employee, error) {
	hasUser := in.UserID != nil
	hasName := in.FirstName != nil && *in.FirstName != "" && in.LastName != nil && *in.LastName != ""

	if hasUser && hasName {
		return nil, ErrNameSourceConflict
	}
	if !hasUser && !hasName {
		return nil, ErrNameRequired
	}
	if hasUser {
		// A linked employee must be an actual member of this tenant — an
		// employee record pointing at a user who has no membership here
		// would be a dangling, meaningless reference.
		m, err := s.roleRepo.MembershipFor(ctx, tenantID, *in.UserID)
		if err != nil || !m.IsActive() {
			return nil, ErrInvalidUser
		}
	}

	e := &Employee{
		TenantID: tenantID, EmployeeNumber: in.EmployeeNumber, DepartmentID: in.DepartmentID,
		UserID: in.UserID, FirstName: in.FirstName, LastName: in.LastName,
	}
	id, err := s.repo.Create(ctx, e)
	if err != nil {
		return nil, err
	}
	e.ID = id
	return e, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Employee, error) {
	return s.repo.ByID(ctx, tenantID, id)
}

func (s *Service) List(ctx context.Context, tenantID int64, p httpx.Pagination) ([]Employee, int, error) {
	return s.repo.List(ctx, tenantID, p)
}

func (s *Service) Update(ctx context.Context, tenantID, id int64, in UpdateInput) (*Employee, error) {
	if err := s.repo.Update(ctx, tenantID, id, in); err != nil {
		return nil, err
	}
	return s.repo.ByID(ctx, tenantID, id)
}

// ToDTO resolves the display name: from the employee record if standalone,
// or by looking up the linked user if UserID is set. Falls back to empty
// strings (never errors) if a linked user somehow can't be found, so a
// listing endpoint degrades gracefully instead of failing entirely over
// one bad row.
func (s *Service) ToDTO(ctx context.Context, e *Employee) DTO {
	dto := DTO{
		ID: e.ID, EmployeeNumber: e.EmployeeNumber, DepartmentID: e.DepartmentID,
		UserID: e.UserID, Status: string(e.Status),
	}
	if e.UserID != nil {
		if u, err := s.userRepo.ByID(ctx, *e.UserID); err == nil {
			dto.FirstName = u.FirstName
			dto.LastName = u.LastName
		}
	} else {
		if e.FirstName != nil {
			dto.FirstName = *e.FirstName
		}
		if e.LastName != nil {
			dto.LastName = *e.LastName
		}
	}
	return dto
}
