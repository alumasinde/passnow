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
	ErrInvalidUser        = errors.New("employees: user_id must belong to a user with an active membership")
)

type Service struct {
	repo     *Repository
	userRepo *users.Repository
	roleRepo *roles.Repository
}

func NewService(repo *Repository, userRepo *users.Repository, roleRepo *roles.Repository) *Service {
	return &Service{repo: repo, userRepo: userRepo, roleRepo: roleRepo}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Employee, error) {
	hasUser := in.UserID != nil
	hasName := in.FirstName != nil && *in.FirstName != "" && in.LastName != nil && *in.LastName != ""

	if hasUser && hasName {
		return nil, ErrNameSourceConflict
	}
	if !hasUser && !hasName {
		return nil, ErrNameRequired
	}
	if hasUser {
		m, err := s.roleRepo.MembershipFor(ctx, *in.UserID)
		if err != nil || !m.IsActive() {
			return nil, ErrInvalidUser
		}
	}

	e := &Employee{
		EmployeeNumber: in.EmployeeNumber, DepartmentID: in.DepartmentID,
		UserID: in.UserID, FirstName: in.FirstName, LastName: in.LastName,
	}
	id, err := s.repo.Create(ctx, e)
	if err != nil {
		return nil, err
	}
	e.ID = id
	return e, nil
}

func (s *Service) Get(ctx context.Context, id int64) (*Employee, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) List(ctx context.Context, p httpx.Pagination) ([]Employee, int, error) {
	return s.repo.List(ctx, p)
}

func (s *Service) ListScoped(ctx context.Context, p httpx.Pagination, departmentID *int64) ([]Employee, int, error) {
	return s.repo.ListScoped(ctx, p, departmentID)
}

func (s *Service) UserDepartment(ctx context.Context, userID int64) (*int64, error) {
	return s.repo.UserDepartment(ctx, userID)
}

func (s *Service) Update(ctx context.Context, id int64, in UpdateInput) (*Employee, error) {
	if err := s.repo.Update(ctx, id, in); err != nil {
		return nil, err
	}
	return s.repo.ByID(ctx, id)
}

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
