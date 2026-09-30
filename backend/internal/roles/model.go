package roles

import "time"

type Role struct {
	ID       int64
	Name     string
	IsSystem bool
}

type MembershipStatus string

const (
	MembershipActive   MembershipStatus = "active"
	MembershipInvited  MembershipStatus = "invited"
	MembershipDisabled MembershipStatus = "disabled"
)

// Membership is tenant-local. The database connection determines the tenant,
// so no tenant identifier is stored on the membership row.
type Membership struct {
	ID        int64
	UserID    int64
	RoleID    int64
	Status    MembershipStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Membership) IsActive() bool {
	return m.Status == MembershipActive
}
