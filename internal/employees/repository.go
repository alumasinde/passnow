package employees

import (
	"context"
	"database/sql"
	"errors"

	"gatepass/internal/database"
	"gatepass/internal/httpx"
)

var (
	ErrNotFound        = errors.New("employees: not found")
	ErrDuplicateNumber = errors.New("employees: employee_number already in use")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const cols = `id, tenant_id, employee_number, first_name, last_name, department_id, user_id, status`

func (r *Repository) scan(row interface{ Scan(dest ...any) error }) (*Employee, error) {
	var e Employee
	if err := row.Scan(&e.ID, &e.TenantID, &e.EmployeeNumber, &e.FirstName, &e.LastName,
		&e.DepartmentID, &e.UserID, &e.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *Repository) ByID(ctx context.Context, tenantID, id int64) (*Employee, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+cols+" FROM employees WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL LIMIT 1", id, tenantID)
	return r.scan(row)
}

func (r *Repository) List(ctx context.Context, tenantID int64, p httpx.Pagination) ([]Employee, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM employees WHERE tenant_id = ? AND deleted_at IS NULL", tenantID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT "+cols+" FROM employees WHERE tenant_id = ? AND deleted_at IS NULL ORDER BY employee_number LIMIT ? OFFSET ?",
		tenantID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Employee
	for rows.Next() {
		e, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *e)
	}
	return out, total, rows.Err()
}

// Create inserts an employee row. The exclusivity rule (exactly one of
// user_id or first_name+last_name) is validated by the SERVICE before
// this is called, and backstopped by a DB CHECK constraint — this method
// just persists whatever it's given.
func (r *Repository) Create(ctx context.Context, e *Employee) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO employees (tenant_id, employee_number, first_name, last_name, department_id, user_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())`,
		e.TenantID, e.EmployeeNumber, e.FirstName, e.LastName, e.DepartmentID, e.UserID)
	if err != nil {
		if database.IsDuplicateKeyErr(err) {
			return 0, ErrDuplicateNumber
		}
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateInput deliberately does NOT allow switching between a linked-user
// employee and a standalone-name employee after creation — that identity
// choice is made once, at creation, to avoid a whole class of "which name
// wins" edge cases. Only employee_number, department, and status change here.
type UpdateInput struct {
	EmployeeNumber *string
	DepartmentID   *int64
	Status         *Status
}

func (r *Repository) Update(ctx context.Context, tenantID, id int64, in UpdateInput) error {
	e, err := r.ByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if in.EmployeeNumber != nil {
		e.EmployeeNumber = *in.EmployeeNumber
	}
	if in.DepartmentID != nil {
		e.DepartmentID = in.DepartmentID
	}
	if in.Status != nil {
		e.Status = *in.Status
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE employees SET employee_number = ?, department_id = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND tenant_id = ?`, e.EmployeeNumber, e.DepartmentID, e.Status, id, tenantID)
	if err != nil && database.IsDuplicateKeyErr(err) {
		return ErrDuplicateNumber
	}
	return err
}
