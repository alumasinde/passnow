package visits

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"gatepass/internal/httpx"
	"gatepass/internal/numbering"
)

var (
	ErrNotFound          = errors.New("visits: not found")
	ErrInvalidTransition = errors.New("visits: this action is not valid for the visit's current status")
)

const badgeSequenceScope = "visit_badge"
const badgePrefix = "VB"

type Repository struct { db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const selectCols = `
	id, visitor_id, visit_type_id, department_id, host_name,
	purpose, expected_time, status, badge_number, badge_token,
	checked_in_at, checked_in_by, checked_out_at, checked_out_by,
	cancelled_at, cancelled_by, cancel_reason,
	created_by, created_at, updated_at, deleted_at
`

func (r *Repository) scan(row interface{ Scan(dest ...any) error }) (*Visit, error) {
	var v Visit
	if err := row.Scan(
		&v.ID, &v.VisitorID, &v.VisitTypeID, &v.DepartmentID, &v.HostName,
		&v.Purpose, &v.ExpectedTime, &v.Status, &v.BadgeNumber, &v.BadgeToken,
		&v.CheckedInAt, &v.CheckedInBy, &v.CheckedOutAt, &v.CheckedOutBy,
		&v.CancelledAt, &v.CancelledBy, &v.CancelReason,
		&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }
		return nil, err
	}
	return &v, nil
}

func (r *Repository) ByID(ctx context.Context, id int64) (*Visit, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+selectCols+" FROM visits WHERE id = ? AND deleted_at IS NULL LIMIT 1", id)
	return r.scan(row)
}

func (r *Repository) ByBadgeToken(ctx context.Context, token string) (*Visit, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+selectCols+" FROM visits WHERE badge_token = ? AND deleted_at IS NULL LIMIT 1", token)
	return r.scan(row)
}

type ListFilter struct { Status *Status; VisitorID *int64; Search string; Date string; CreatedBy *int64; DepartmentID *int64 }

func (r *Repository) List(ctx context.Context, f ListFilter, p httpx.Pagination) ([]Visit, int, error) {
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	if f.Status != nil { where += " AND status = ?"; args = append(args, *f.Status) }
	if f.VisitorID != nil { where += " AND visitor_id = ?"; args = append(args, *f.VisitorID) }
	if f.CreatedBy != nil { where += " AND created_by = ?"; args = append(args, *f.CreatedBy) }
	if f.DepartmentID != nil { where += " AND department_id = ?"; args = append(args, *f.DepartmentID) }
	if f.Date != "" { where += " AND DATE(COALESCE(expected_time, created_at)) = ?"; args = append(args, f.Date) }
	if f.Search != "" { where += " AND (badge_number LIKE ? OR host_name LIKE ? OR purpose LIKE ? OR visitor_id IN (SELECT id FROM visitors WHERE first_name LIKE ? OR last_name LIKE ?))"; like := "%"+f.Search+"%"; args=append(args,like,like,like,like,like) }

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM visits "+where, args...).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.db.QueryContext(ctx,
		"SELECT "+selectCols+" FROM visits "+where+" ORDER BY COALESCE(expected_time, created_at) DESC, created_at DESC LIMIT ? OFFSET ?",
		append(args, p.Limit, p.Offset)...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var out []Visit
	for rows.Next() {
		v, err := r.scan(rows); if err != nil { return nil, 0, err }
		out = append(out, *v)
	}
	return out, total, rows.Err()
}

func (r *Repository) Create(ctx context.Context, v *Visit) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO visits
			(visitor_id, visit_type_id, department_id, host_name,
			 purpose, expected_time, status, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		v.VisitorID, v.VisitTypeID, v.DepartmentID, v.HostName,
		v.Purpose, v.ExpectedTime, v.Status, v.CreatedBy)
	if err != nil { return 0, err }
	return res.LastInsertId()
}

func (r *Repository) CheckIn(ctx context.Context, id, actorUserID int64) (*Visit, error) {
	tx, err := r.db.BeginTx(ctx, nil); if err != nil { return nil, err }
	defer tx.Rollback()
	var status Status
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM visits WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, id).Scan(&status)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }; return nil, err }
	if status != StatusScheduled && status != StatusExpected { return nil, ErrInvalidTransition }
	period := time.Now().UTC().Format("2006")
	seq, err := numbering.Next(ctx, tx, badgeSequenceScope, period); if err != nil { return nil, err }
	badgeNumber := numbering.Format(badgePrefix, period, seq)
	badgeToken, err := randomToken(); if err != nil { return nil, err }
	if _, err := tx.ExecContext(ctx, `
		UPDATE visits SET status = 'checked_in', badge_number = ?, badge_token = ?,
			checked_in_at = NOW(), checked_in_by = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`, badgeNumber, badgeToken, actorUserID, id); err != nil { return nil, err }
	if err := tx.Commit(); err != nil { return nil, err }
	return r.ByID(ctx, id)
}

func (r *Repository) CheckOut(ctx context.Context, id, actorUserID int64) (*Visit, error) {
	tx, err := r.db.BeginTx(ctx, nil); if err != nil { return nil, err }
	defer tx.Rollback()
	var status Status
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM visits WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, id).Scan(&status)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }; return nil, err }
	if status != StatusCheckedIn { return nil, ErrInvalidTransition }
	if _, err := tx.ExecContext(ctx, `
		UPDATE visits SET status = 'checked_out', checked_out_at = NOW(), checked_out_by = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`, actorUserID, id); err != nil { return nil, err }
	if err := tx.Commit(); err != nil { return nil, err }
	return r.ByID(ctx, id)
}

func (r *Repository) Cancel(ctx context.Context, id, actorUserID int64, reason string) (*Visit, error) {
	tx, err := r.db.BeginTx(ctx, nil); if err != nil { return nil, err }
	defer tx.Rollback()
	var status Status
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM visits WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, id).Scan(&status)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }; return nil, err }
	if status != StatusScheduled && status != StatusExpected { return nil, ErrInvalidTransition }
	if _, err := tx.ExecContext(ctx, `
		UPDATE visits SET status = 'cancelled', cancelled_at = NOW(), cancelled_by = ?, cancel_reason = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`, actorUserID, reason, id); err != nil { return nil, err }
	if err := tx.Commit(); err != nil { return nil, err }
	return r.ByID(ctx, id)
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil { return "", err }
	return hex.EncodeToString(b), nil
}


func (r *Repository) UserDepartment(ctx context.Context, userID int64) (*int64, error) {
	var departmentID *int64
	err := r.db.QueryRowContext(ctx, "SELECT department_id FROM users WHERE id = ? LIMIT 1", userID).Scan(&departmentID)
	if err != nil { return nil, err }
	return departmentID, nil
}
