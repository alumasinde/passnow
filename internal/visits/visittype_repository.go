package visits

import (
	"context"
	"database/sql"
	"errors"
)

var ErrVisitTypeNotFound = errors.New("visits: visit type not found")

type VisitTypeRepository struct { db *sql.DB }
func NewVisitTypeRepository(db *sql.DB) *VisitTypeRepository { return &VisitTypeRepository{db: db} }

const visitTypeCols = `id, name, code, description, active`

func (r *VisitTypeRepository) List(ctx context.Context, activeOnly bool) ([]VisitType, error) {
	q := "SELECT " + visitTypeCols + " FROM visit_types WHERE deleted_at IS NULL"
	if activeOnly { q += " AND active = 1" }
	q += " ORDER BY name"
	rows, err := r.db.QueryContext(ctx, q); if err != nil { return nil, err }
	defer rows.Close()
	var out []VisitType
	for rows.Next() {
		var t VisitType
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.Description, &t.Active); err != nil { return nil, err }
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *VisitTypeRepository) ByID(ctx context.Context, id int64) (*VisitType, error) {
	var t VisitType
	err := r.db.QueryRowContext(ctx, "SELECT "+visitTypeCols+" FROM visit_types WHERE id = ? AND deleted_at IS NULL LIMIT 1", id).
		Scan(&t.ID, &t.Name, &t.Code, &t.Description, &t.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, ErrVisitTypeNotFound }
		return nil, err
	}
	return &t, nil
}

func (r *VisitTypeRepository) Create(ctx context.Context, in VisitTypeInput) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO visit_types (name, code, description, active, created_at, updated_at)
		VALUES (?, ?, ?, 1, NOW(), NOW())`, in.Name, in.Code, in.Description)
	if err != nil { return 0, err }
	return res.LastInsertId()
}

func (r *VisitTypeRepository) Update(ctx context.Context, id int64, in VisitTypeInput) error {
	t, err := r.ByID(ctx, id); if err != nil { return err }
	if in.Name != "" { t.Name = in.Name }
	if in.Code != "" { t.Code = in.Code }
	if in.Description != nil { t.Description = in.Description }
	if in.Active != nil { t.Active = *in.Active }
	_, err = r.db.ExecContext(ctx, `
		UPDATE visit_types SET name = ?, code = ?, description = ?, active = ?, updated_at = NOW()
		WHERE id = ?`, t.Name, t.Code, t.Description, t.Active, id)
	return err
}
