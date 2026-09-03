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

type EntrySource string

const (
	EntrySourceWalkIn EntrySource = "walk_in"
	EntrySourcePreRegistered EntrySource = "pre_registered"
)

type Visit struct {
	ID           int64
	VisitorID    int64
	EntrySource EntrySource
	VisitTypeID  *int64
	DepartmentID *int64
	HostName     *string

	Purpose      *string
	ExpectedTime *time.Time
	ExpectedDepartureAt *time.Time
	ArrivedAt *time.Time

	Status Status

	BadgeNumber *string
	BadgeToken  *string
	QRToken *string
	QRIssuedAt *time.Time
	QRInvalidatedAt *time.Time

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

func (v *Visit) CanCheckIn() bool {
	return v.Status == StatusScheduled || v.Status == StatusExpected
}

func (v *Visit) CanCheckOut() bool {
	return v.Status == StatusCheckedIn
}

func (v *Visit) CanCancel() bool {
	return v.Status == StatusScheduled || v.Status == StatusExpected
}
