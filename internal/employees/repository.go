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

const cols = `id, employee_number, first_name, last_name, department_id, user_id, status`

func (r *Repository) scan(row interface{ Scan(dest ...any) error }) (*Employee, error) {
	var e Employee
	if err := row.Scan(&e.ID, &e.EmployeeNumber, &e.FirstName, &e.LastName,
		&e.DepartmentID, &e.UserID, &e.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *Repository) ByID(ctx context.Context, id int64) (*Employee, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+cols+" FROM employees WHERE id = ? AND deleted_at IS NULL LIMIT 1", id)
	return r.scan(row)
}

func (r *Repository) List(ctx context.Context, p httpx.Pagination) ([]Employee, int, error) {
	return r.ListScoped(ctx, p, nil)
}

func (r *Repository) ListScoped(ctx context.Context, p httpx.Pagination, departmentID *int64) ([]Employee, int, error) {
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	if departmentID != nil {
		where += " AND department_id = ?"
		args = append(args, *departmentID)
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM employees "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, p.Limit, p.Offset)
	rows, err := r.db.QueryContext(ctx, "SELECT "+cols+" FROM employees "+where+" ORDER BY employee_number LIMIT ? OFFSET ?", queryArgs...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var out []Employee
	for rows.Next() {
		e, err := r.scan(rows)
		if err != nil { return nil, 0, err }
		out = append(out, *e)
	}
	return out, total, rows.Err()
}

func (r *Repository) Create(ctx context.Context, e *Employee) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO employees (employee_number, first_name, last_name, department_id, user_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'active', NOW(), NOW())`,
		e.EmployeeNumber, e.FirstName, e.LastName, e.DepartmentID, e.UserID)
	if err != nil {
		if database.IsDuplicateKeyErr(err) {
			return 0, ErrDuplicateNumber
		}
		return 0, err
	}
	return res.LastInsertId()
}

type UpdateInput struct {
	EmployeeNumber *string
	DepartmentID   *int64
	Status         *Status
}

func (r *Repository) Update(ctx context.Context, id int64, in UpdateInput) error {
	e, err := r.ByID(ctx, id)
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
		WHERE id = ?`, e.EmployeeNumber, e.DepartmentID, e.Status, id)
	if err != nil && database.IsDuplicateKeyErr(err) {
		return ErrDuplicateNumber
	}
	return err
}
