package employees

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// Employee is deliberately slim — this is a gatepass system, not an HR
// system. It exists so hosts/requesters/department rosters can reference
// "a person who works here" without requiring that person to have a
// login.
//
// Name rule: exactly one of (UserID set) or (FirstName+LastName set) —
// never both, never neither. Enforced by a DB CHECK, not just app code.
// If UserID is set, the display name comes from the linked user (see
// Service/DTO layer), keeping identity in one place instead of two
// records that can drift apart.
type Employee struct {
	ID             int64
	TenantID       int64
	EmployeeNumber string
	FirstName      *string // NULL if UserID is set
	LastName       *string // NULL if UserID is set
	DepartmentID   *int64
	UserID         *int64 // set only for employees who also log in
	Status         Status
}
