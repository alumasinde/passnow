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

func (r *Repository) MembershipFor(ctx context.Context, userID int64) (*Membership, error) {
	var m Membership
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, role_id, status, created_at, updated_at
		FROM tenant_memberships
		WHERE user_id = ? LIMIT 1`, userID,
	).Scan(&m.ID, &m.UserID, &m.RoleID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *Repository) RoleByID(ctx context.Context, roleID int64) (*Role, error) {
	var role Role
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, is_system FROM roles
		WHERE id = ? LIMIT 1`, roleID,
	).Scan(&role.ID, &role.Name, &role.IsSystem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *Repository) PermissionCodesForRole(ctx context.Context, roleID int64) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.code
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = ?`, roleID)
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

func (r *Repository) CreateRole(ctx context.Context, name string, isSystem bool) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO roles (name, is_system, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())`, name, isSystem)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CreateMembership(ctx context.Context, userID, roleID int64, status MembershipStatus) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_memberships (user_id, role_id, status, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, userID, roleID, status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CreateMembershipTx(ctx context.Context, tx *sql.Tx, userID, roleID int64, status MembershipStatus) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO tenant_memberships (user_id, role_id, status, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, userID, roleID, status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CreateRoleTx(ctx context.Context, tx *sql.Tx, name string, isSystem bool) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO roles (name, is_system, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())`, name, isSystem)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type Permission struct {
	Code  string
	Label string
}

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

func (r *Repository) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, is_system FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.IsSystem); err != nil {
			return nil, err
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

func (r *Repository) SetRolePermissions(ctx context.Context, roleID int64, codes []string) error {
	if _, err := r.RoleByID(ctx, roleID); err != nil {
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

func (r *Repository) GrantAllPermissions(ctx context.Context, tx *sql.Tx, roleID int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id) SELECT ?, id FROM permissions`, roleID)
	return err
}

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

func (r *Repository) ListMemberships(ctx context.Context) ([]MembershipView, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tm.id, u.id, u.email, u.first_name, u.last_name, r.id, r.name, tm.status
		FROM tenant_memberships tm
		JOIN users u ON u.id = tm.user_id
		JOIN roles r ON r.id = tm.role_id
		ORDER BY u.last_name, u.first_name`)
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

func (r *Repository) UpdateMembership(ctx context.Context, membershipID int64, roleID *int64, status *MembershipStatus) error {
	m, err := r.membershipByID(ctx, membershipID)
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
		WHERE id = ?`, m.RoleID, m.Status, membershipID)
	return err
}

func (r *Repository) membershipByID(ctx context.Context, membershipID int64) (*Membership, error) {
	var m Membership
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, role_id, status, created_at, updated_at
		FROM tenant_memberships WHERE id = ? LIMIT 1`, membershipID,
	).Scan(&m.ID, &m.UserID, &m.RoleID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}
