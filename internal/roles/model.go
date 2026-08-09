package roles

import "time"

type Role struct {
	ID       int64
	TenantID int64
	Name     string
	IsSystem bool
}

type MembershipStatus string

const (
	MembershipActive   MembershipStatus = "active"
	MembershipInvited  MembershipStatus = "invited"
	MembershipDisabled MembershipStatus = "disabled"
)

// Membership is the row that determines what a user can do IN a specific
// tenant. This — not the JWT claims alone — is the source of truth for
// authorization; the JWT just avoids a DB round trip for cheap checks
// within a short TTL.
type Membership struct {
	ID        int64
	TenantID  int64
	UserID    int64
	RoleID    int64
	Status    MembershipStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Membership) IsActive() bool {
	return m.Status == MembershipActive
}
