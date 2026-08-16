package roles

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("roles: not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// MembershipFor returns the caller's membership (role + status) within a
// specific tenant. Every login and every permission check goes through
// this — tenant_id always comes from the resolved tenant context, never
// from client input.
func (r *Repository) MembershipFor(ctx context.Context, tenantID, userID int64) (*Membership, error) {
	var m Membership
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, role_id, status, created_at, updated_at
		FROM tenant_memberships
		WHERE tenant_id = ? AND user_id = ? LIMIT 1`, tenantID, userID,
	).Scan(&m.ID, &m.TenantID, &m.UserID, &m.RoleID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *Repository) RoleByID(ctx context.Context, tenantID, roleID int64) (*Role, error) {
	var role Role
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, is_system FROM roles
		WHERE id = ? AND tenant_id = ? LIMIT 1`, roleID, tenantID,
	).Scan(&role.ID, &role.TenantID, &role.Name, &role.IsSystem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

// PermissionCodesForRole returns every permission code granted to a role.
// Always filtered so a role from tenant A can never be queried using
// tenant B's ID by mistake (join guards this even though role_id is
// already unique).
func (r *Repository) PermissionCodesForRole(ctx context.Context, tenantID, roleID int64) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.code
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		JOIN roles r ON r.id = rp.role_id
		WHERE r.id = ? AND r.tenant_id = ?`, roleID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	codes := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes[code] = true
	}
	return codes, rows.Err()
}

func (r *Repository) CreateRole(ctx context.Context, tenantID int64, name string, isSystem bool) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO roles (tenant_id, name, is_system, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, tenantID, name, isSystem)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// --- Role & permission management (Settings) ---------------------------

func (r *Repository) CreateMembership(ctx context.Context, tenantID, userID, roleID int64, status MembershipStatus) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_memberships (tenant_id, user_id, role_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())`, tenantID, userID, roleID, status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CreateMembershipTx is the same as CreateMembership but runs inside an
// existing transaction — used by the platform bootstrap flow, which
// creates the tenant, role, user, and membership atomically.
func (r *Repository) CreateMembershipTx(ctx context.Context, tx *sql.Tx, tenantID, userID, roleID int64, status MembershipStatus) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO tenant_memberships (tenant_id, user_id, role_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())`, tenantID, userID, roleID, status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CreateRoleTx is CreateRole run inside an existing transaction (bootstrap use).
func (r *Repository) CreateRoleTx(ctx context.Context, tx *sql.Tx, tenantID int64, name string, isSystem bool) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO roles (tenant_id, name, is_system, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, tenantID, name, isSystem)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type Permission struct {
	Code  string
	Label string
}

// AllPermissions returns the full global permission taxonomy (not
// tenant-scoped — permissions are a fixed catalog defined by migrations,
// same set available to every tenant to assign to their own roles).
func (r *Repository) AllPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code, label FROM permissions ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.Code, &p.Label); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) ListRoles(ctx context.Context, tenantID int64) ([]Role, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, is_system FROM roles WHERE tenant_id = ? ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.IsSystem); err != nil {
			return nil, err
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

// SetRolePermissions replaces a role's entire permission set atomically
// (delete then insert in one transaction — never a partial update where a
// crash mid-way leaves stale grants). Unknown codes are silently ignored
// rather than erroring, so a client sending a slightly-stale permission
// list doesn't fail the whole request — but this means callers should
// treat the return value's applied count, not just err==nil, as
// confirmation if that matters.
func (r *Repository) SetRolePermissions(ctx context.Context, tenantID, roleID int64, codes []string) error {
	if _, err := r.RoleByID(ctx, tenantID, roleID); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, roleID); err != nil {
		return err
	}

	for _, code := range codes {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT ?, id FROM permissions WHERE code = ?`, roleID, code); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GrantAllPermissions is used once, when bootstrapping a tenant's first
// "Tenant Admin" role — every current permission in the catalog.
func (r *Repository) GrantAllPermissions(ctx context.Context, tx *sql.Tx, roleID int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id) SELECT ?, id FROM permissions`, roleID)
	return err
}

// MembershipView is a tenant membership joined with the user's identity —
// what a Settings "Users" page actually needs to display.
type MembershipView struct {
	MembershipID int64
	UserID       int64
	Email        string
	FirstName    string
	LastName     string
	RoleID       int64
	RoleName     string
	Status       MembershipStatus
}

func (r *Repository) ListMemberships(ctx context.Context, tenantID int64) ([]MembershipView, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tm.id, u.id, u.email, u.first_name, u.last_name, r.id, r.name, tm.status
		FROM tenant_memberships tm
		JOIN users u ON u.id = tm.user_id
		JOIN roles r ON r.id = tm.role_id
		WHERE tm.tenant_id = ?
		ORDER BY u.last_name, u.first_name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MembershipView
	for rows.Next() {
		var m MembershipView
		if err := rows.Scan(&m.MembershipID, &m.UserID, &m.Email, &m.FirstName, &m.LastName, &m.RoleID, &m.RoleName, &m.Status); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateMembership(ctx context.Context, tenantID, membershipID int64, roleID *int64, status *MembershipStatus) error {
	m, err := r.membershipByID(ctx, tenantID, membershipID)
	if err != nil {
		return err
	}
	if roleID != nil {
		m.RoleID = *roleID
	}
	if status != nil {
		m.Status = *status
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE tenant_memberships SET role_id = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND tenant_id = ?`, m.RoleID, m.Status, membershipID, tenantID)
	return err
}

func (r *Repository) membershipByID(ctx context.Context, tenantID, membershipID int64) (*Membership, error) {
	var m Membership
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, role_id, status, created_at, updated_at
		FROM tenant_memberships WHERE id = ? AND tenant_id = ? LIMIT 1`, membershipID, tenantID,
	).Scan(&m.ID, &m.TenantID, &m.UserID, &m.RoleID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}
