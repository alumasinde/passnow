package employees

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Employee struct {
	ID             int64
	EmployeeNumber string
	FirstName      *string
	LastName       *string
	DepartmentID   *int64
	UserID         *int64
	Status         Status
}
