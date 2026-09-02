package departments

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("departments: not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, activeOnly bool) ([]Department, error) {
	q := `SELECT id, name, code, active FROM departments WHERE deleted_at IS NULL`
	if activeOnly {
		q += " AND active = 1"
	}
	q += " ORDER BY name"

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Department
	for rows.Next() {
		var d Department
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.Active); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) ByID(ctx context.Context, id int64) (*Department, error) {
	var d Department
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, code, active FROM departments
		WHERE id = ? AND deleted_at IS NULL LIMIT 1`, id,
	).Scan(&d.ID, &d.Name, &d.Code, &d.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *Repository) Create(ctx context.Context, name, code string, active bool) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO departments (name, code, active, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, name, code, active)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, id int64, in Input) error {
	d, err := r.ByID(ctx, id)
	if err != nil {
		return err
	}
	if in.Name != "" {
		d.Name = in.Name
	}
	if in.Code != "" {
		d.Code = in.Code
	}
	if in.Active != nil {
		d.Active = *in.Active
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE departments SET name = ?, code = ?, active = ?, updated_at = NOW()
		WHERE id = ?`, d.Name, d.Code, d.Active, id)
	return err
}
