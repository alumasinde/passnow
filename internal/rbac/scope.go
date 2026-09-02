package rbac

import (
    "context"
    "errors"

    "gatepass/internal/roles"
)

var ErrScopeDenied = errors.New("rbac: scope denied")

type Subject struct {
    UserID int64
    DepartmentID *int64
}

func (e *Engine) Subject(ctx context.Context, userID int64) (Subject, error) {
    m, err := e.repo.MembershipViewByUserID(ctx, userID)
    if err != nil { return Subject{}, err }
    return Subject{UserID: userID, DepartmentID: m.DepartmentID}, nil
}

func AllowsCreatedBy(scope Scope, actorID, createdBy int64) bool {
    return scope == ScopeAll || (scope == ScopeOwn && actorID == createdBy)
}

func AllowsDepartment(scope Scope, actorID int64, actorDepartmentID *int64, resourceDepartmentID *int64, createdBy int64) bool {
    if scope == ScopeAll { return true }
    if scope == ScopeOwn { return actorID == createdBy }
    if scope != ScopeDepartment || actorDepartmentID == nil || resourceDepartmentID == nil { return false }
    return *actorDepartmentID == *resourceDepartmentID
}

func RequireScope(scope Scope, allowed bool) error {
    if scope == ScopeAll || allowed { return nil }
    return ErrScopeDenied
}
