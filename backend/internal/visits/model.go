package visits

import "time"

type Status string

const (
	StatusScheduled  Status = "scheduled"
	StatusExpected   Status = "expected"
	StatusCheckedIn  Status = "checked_in"
	StatusCheckedOut Status = "checked_out"
	StatusCancelled  Status = "cancelled"
	StatusNoShow     Status = "no_show"
	StatusExpired    Status = "expired"
)

// Visit is the domain/DB entity. Never expose this directly — use DTO.
type Visit struct {
	ID           int64
	TenantID     int64
	VisitorID    int64
	VisitTypeID  *int64
	DepartmentID *int64
	HostName     *string

	Purpose      *string
	ExpectedTime *time.Time

	Status Status

	BadgeNumber *string
	BadgeToken  *string

	CheckedInAt  *time.Time
	CheckedInBy  *int64
	CheckedOutAt *time.Time
	CheckedOutBy *int64

	CancelledAt  *time.Time
	CancelledBy  *int64
	CancelReason *string

	CreatedBy *int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// CanCheckIn reports whether this visit's current status permits check-in.
// Centralizing the state machine here (not scattered across handler/repo
// conditionals) keeps "what transitions are legal" defined in exactly one
// place.
func (v *Visit) CanCheckIn() bool {
	return v.Status == StatusScheduled || v.Status == StatusExpected
}

func (v *Visit) CanCheckOut() bool {
	return v.Status == StatusCheckedIn
}

func (v *Visit) CanCancel() bool {
	return v.Status == StatusScheduled || v.Status == StatusExpected
}
