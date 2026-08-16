package visitors

import (
	"context"
	"database/sql"
	"errors"
)

var ErrIDTypeNotFound = errors.New("visitors: id type not found")

type IDTypeRepository struct {
	db *sql.DB
}

func NewIDTypeRepository(db *sql.DB) *IDTypeRepository {
	return &IDTypeRepository{db: db}
}

func (r *IDTypeRepository) List(ctx context.Context, tenantID int64, activeOnly bool) ([]IDType, error) {
	q := `SELECT id, tenant_id, name, code, requires_number, active FROM id_types
	      WHERE tenant_id = ? AND deleted_at IS NULL`
	args := []any{tenantID}
	if activeOnly {
		q += " AND active = 1"
	}
	q += " ORDER BY name"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []IDType
	for rows.Next() {
		var t IDType
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Code, &t.RequiresNumber, &t.Active); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ByID enforces tenant scope — an id_type belonging to another tenant is
// indistinguishable from a nonexistent one to the caller.
func (r *IDTypeRepository) ByID(ctx context.Context, tenantID, id int64) (*IDType, error) {
	var t IDType
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, code, requires_number, active FROM id_types
		WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL LIMIT 1`, id, tenantID,
	).Scan(&t.ID, &t.TenantID, &t.Name, &t.Code, &t.RequiresNumber, &t.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIDTypeNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *IDTypeRepository) Create(ctx context.Context, tenantID int64, name, code string, requiresNumber bool) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO id_types (tenant_id, name, code, requires_number, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, NOW(), NOW())`, tenantID, name, code, requiresNumber)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *IDTypeRepository) Update(ctx context.Context, tenantID, id int64, in IDTypeInput) error {
	t, err := r.ByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if in.Name != "" {
		t.Name = in.Name
	}
	if in.Code != "" {
		t.Code = in.Code
	}
	if in.RequiresNumber != nil {
		t.RequiresNumber = *in.RequiresNumber
	}
	if in.Active != nil {
		t.Active = *in.Active
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE id_types SET name = ?, code = ?, requires_number = ?, active = ?, updated_at = NOW()
		WHERE id = ? AND tenant_id = ?`,
		t.Name, t.Code, t.RequiresNumber, t.Active, id, tenantID)
	return err
}

func (r *IDTypeRepository) SoftDelete(ctx context.Context, tenantID, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE id_types SET deleted_at = NOW(), active = 0 WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL`,
		id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrIDTypeNotFound
	}
	return nil
}
