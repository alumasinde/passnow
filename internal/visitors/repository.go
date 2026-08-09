package visitors

import (
	"context"
	"database/sql"
	"errors"

	"gatepass/internal/database"
	"gatepass/internal/httpx"
)

var (
	ErrNotFound          = errors.New("visitors: not found")
	ErrDuplicateIDNumber = errors.New("visitors: this ID document is already registered")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const selectCols = `
	id, tenant_id, first_name, last_name, id_type_id, id_number, company_id,
	phone, email, photo_ref, notes, source, status, blacklist_reason,
	created_by, updated_by, created_at, updated_at, deleted_at
`

func (r *Repository) scan(row interface{ Scan(dest ...any) error }) (*Visitor, error) {
	var v Visitor
	if err := row.Scan(
		&v.ID, &v.TenantID, &v.FirstName, &v.LastName, &v.IDTypeID, &v.IDNumber, &v.CompanyID,
		&v.Phone, &v.Email, &v.PhotoRef, &v.Notes, &v.Source, &v.Status, &v.BlacklistReason,
		&v.CreatedBy, &v.UpdatedBy, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

// ByID enforces tenant scope in the WHERE clause — this is the one place
// every other visitor lookup funnels through, so tenant isolation for
// visitors lives in exactly one query.
func (r *Repository) ByID(ctx context.Context, tenantID, id int64) (*Visitor, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+selectCols+" FROM visitors WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL LIMIT 1",
		id, tenantID)
	return r.scan(row)
}

type ListFilter struct {
	Status    *Status
	CompanyID *int64
	Search    string // matches exact id_number OR first/last name prefix
}

func (r *Repository) List(ctx context.Context, tenantID int64, f ListFilter, p httpx.Pagination) ([]Visitor, int, error) {
	where := "WHERE tenant_id = ? AND deleted_at IS NULL"
	whereArgs := []any{tenantID}

	if f.Status != nil {
		where += " AND status = ?"
		whereArgs = append(whereArgs, *f.Status)
	}
	if f.CompanyID != nil {
		where += " AND company_id = ?"
		whereArgs = append(whereArgs, *f.CompanyID)
	}
	if f.Search != "" {
		// Single search box covers both cases: an exact id_number match
		// (the fast path for a returning visitor) OR a first/last name
		// prefix match. No separate "search mode" for the UI to expose.
		where += " AND (id_number = ? OR first_name LIKE ? OR last_name LIKE ?)"
		like := f.Search + "%"
		whereArgs = append(whereArgs, f.Search, like, like)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM visitors "+where, whereArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := "ORDER BY created_at DESC"
	selectArgs := append([]any{}, whereArgs...)
	if f.Search != "" {
		// Exact id_number hits sort first, ahead of name matches.
		orderBy = "ORDER BY (id_number = ?) DESC, created_at DESC"
		selectArgs = append(selectArgs, f.Search)
	}
	selectArgs = append(selectArgs, p.Limit, p.Offset)

	rows, err := r.db.QueryContext(ctx,
		"SELECT "+selectCols+" FROM visitors "+where+" "+orderBy+" LIMIT ? OFFSET ?",
		selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Visitor
	for rows.Next() {
		v, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *v)
	}
	return out, total, rows.Err()
}

func (r *Repository) Create(ctx context.Context, v *Visitor) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO visitors
			(tenant_id, first_name, last_name, id_type_id, id_number, company_id,
			 phone, email, notes, source, status, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, ?, NOW(), NOW())`,
		v.TenantID, v.FirstName, v.LastName, v.IDTypeID, v.IDNumber, v.CompanyID,
		v.Phone, v.Email, v.Notes, v.Source, v.CreatedBy, v.CreatedBy,
	)
	if err != nil {
		if database.IsDuplicateKeyErr(err) {
			return 0, ErrDuplicateIDNumber
		}
		return 0, err
	}
	return res.LastInsertId()
}

// Update applies a partial update. Loads the current row first (tenant-
// scoped) so unset fields in the input are preserved rather than nulled.
func (r *Repository) Update(ctx context.Context, tenantID, id int64, in UpdateInput, updatedBy int64) (*Visitor, error) {
	v, err := r.ByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	if in.FirstName != nil {
		v.FirstName = *in.FirstName
	}
	if in.LastName != nil {
		v.LastName = *in.LastName
	}
	if in.IDTypeID != nil {
		v.IDTypeID = *in.IDTypeID
	}
	if in.IDNumber != nil {
		v.IDNumber = in.IDNumber
	}
	if in.CompanyID != nil {
		v.CompanyID = in.CompanyID
	}
	if in.Phone != nil {
		v.Phone = in.Phone
	}
	if in.Email != nil {
		v.Email = in.Email
	}
	if in.Notes != nil {
		v.Notes = in.Notes
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE visitors SET
			first_name = ?, last_name = ?, id_type_id = ?, id_number = ?, company_id = ?,
			phone = ?, email = ?, notes = ?, updated_by = ?, updated_at = NOW()
		WHERE id = ? AND tenant_id = ?`,
		v.FirstName, v.LastName, v.IDTypeID, v.IDNumber, v.CompanyID,
		v.Phone, v.Email, v.Notes, updatedBy, id, tenantID,
	)
	if err != nil {
		if database.IsDuplicateKeyErr(err) {
			return nil, ErrDuplicateIDNumber
		}
		return nil, err
	}
	v.UpdatedBy = &updatedBy
	return v, nil
}

func (r *Repository) SetBlacklist(ctx context.Context, tenantID, id int64, blacklisted bool, reason *string, updatedBy int64) error {
	status := StatusActive
	if blacklisted {
		status = StatusBlacklisted
	} else {
		reason = nil
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE visitors SET status = ?, blacklist_reason = ?, updated_by = ?, updated_at = NOW()
		WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL`,
		status, reason, updatedBy, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
