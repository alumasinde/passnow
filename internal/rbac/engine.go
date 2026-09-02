package rbac

import (
    "context"
    "strings"

    "gatepass/internal/roles"
)

type Scope string

const (
    ScopeNone Scope = ""
    ScopeOwn Scope = "own"
    ScopeDepartment Scope = "department"
    ScopeAssigned Scope = "assigned"
    ScopeAll Scope = "all"
)

type Decision struct {
    Allowed bool
    Granted string
    Resource string
    Action string
    Scope Scope
}

type Engine struct { repo *roles.Repository }

func New(repo *roles.Repository) *Engine { return &Engine{repo: repo} }

func Parse(code string) (resource, action string, scope Scope) {
    parts := strings.Split(strings.TrimSpace(code), ".")
    if len(parts) >= 2 { resource, action = parts[0], parts[1] }
    if len(parts) >= 3 { scope = Scope(parts[2]) }
    return
}

func (e *Engine) PermissionsForUser(ctx context.Context, userID, claimedRoleID int64) (map[string]bool, error) {
    membership, err := e.repo.MembershipFor(ctx, userID)
    if err != nil || !membership.IsActive() || membership.RoleID != claimedRoleID { return nil, roles.ErrNotFound }
    return e.repo.PermissionCodesForRole(ctx, membership.RoleID)
}

func (e *Engine) Authorize(ctx context.Context, userID, claimedRoleID int64, requested string) (Decision, error) {
    perms, err := e.PermissionsForUser(ctx, userID, claimedRoleID)
    if err != nil { return Decision{}, err }
    if perms[requested] {
        r,a,s:=Parse(requested); return Decision{Allowed:true,Granted:requested,Resource:r,Action:a,Scope:s},nil
    }
    r,a,s:=Parse(requested)
    if s!=ScopeNone {
        // all is intentionally NOT treated as a wildcard for department/own.
        // Scope must be explicitly enforced by Phase 3.
        if perms[r+"."+a+".all"] { return Decision{Allowed:true,Granted:r+"."+a+".all",Resource:r,Action:a,Scope:ScopeAll},nil }
    }
    return Decision{Allowed:false,Resource:r,Action:a,Scope:s},nil
}

func (e *Engine) AuthorizeAny(ctx context.Context, userID, roleID int64, requested ...string) (Decision,error) {
    for _, code := range requested {
        d,err:=e.Authorize(ctx,userID,roleID,code)
        if err!=nil { return Decision{},err }
        if d.Allowed { return d,nil }
    }
    return Decision{},nil
}

func (e *Engine) AuthorizeAll(ctx context.Context, userID, roleID int64, requested ...string) (Decision,error) {
    for _, code := range requested {
        d,err:=e.Authorize(ctx,userID,roleID,code)
        if err!=nil || !d.Allowed { return Decision{},err }
    }
    return Decision{Allowed:true},nil
}
