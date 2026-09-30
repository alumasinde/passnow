package users

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

// User is a platform-global account. Tenant-specific role/access lives in
// tenant_memberships, not here — a user can belong to multiple tenants
// with a different role in each.
type User struct {
	ID               int64
	Email            string
	PasswordHash     string
	FirstName        string
	LastName         string
	DepartmentID     *int64
	Status           Status
	MustChangePassword bool
	FailedLoginCount int
	LockedUntil      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive && u.DeletedAt == nil
}

func (u *User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && u.LockedUntil.After(now)
}
