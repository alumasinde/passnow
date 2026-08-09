package employees

// DTO always has a resolved, non-null display name — callers never need
// to know whether that name came from this employee's own record or from
// a linked user account.
type DTO struct {
	ID             int64  `json:"id"`
	EmployeeNumber string `json:"employee_number"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DepartmentID   *int64 `json:"department_id"`
	UserID         *int64 `json:"user_id"`
	Status         string `json:"status"`
}

// CreateInput: exactly one of UserID or (FirstName+LastName) must be set —
// validated in the service. Linking UserID means this employee's name is
// always read live from their user account (they log in, so it's already
// the source of truth); otherwise the name lives here.
type CreateInput struct {
	EmployeeNumber string  `json:"employee_number"`
	DepartmentID   *int64  `json:"department_id"`
	UserID         *int64  `json:"user_id"`
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
}
