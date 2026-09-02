package visitors

import "time"

type Source string

const (
	SourceWalkIn        Source = "walk_in"
	SourcePreRegistered Source = "pre_registered"
)

type Status string

const (
	StatusActive      Status = "active"
	StatusBlacklisted Status = "blacklisted"
)

// Visitor belongs to the tenant database selected for the request.
type Visitor struct {
	ID int64
	FirstName string
	LastName string
	IDTypeID int64
	IDNumber *string
	CompanyID *int64
	Phone *string
	Email *string
	PhotoRef *string
	Notes *string
	Source Source
	Status Status
	BlacklistReason *string
	CreatedBy *int64
	UpdatedBy *int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (v *Visitor) FullName() string { return v.FirstName+" "+v.LastName }

type IDType struct {
	ID int64
	Name string
	Code string
	RequiresNumber bool
	Active bool
}

type Company struct {
	ID int64
	Name string
	Phone *string
	Email *string
	Address *string
	Active bool
}
