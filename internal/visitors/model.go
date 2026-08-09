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

// Visitor is the domain/DB entity. Never expose this directly — use DTO.
type Visitor struct {
	ID       int64
	TenantID int64

	FirstName string
	LastName  string

	IDTypeID int64
	IDNumber *string

	CompanyID *int64

	Phone    *string
	Email    *string
	PhotoRef *string
	Notes    *string

	Source          Source
	Status          Status
	BlacklistReason *string

	CreatedBy *int64
	UpdatedBy *int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (v *Visitor) FullName() string {
	return v.FirstName + " " + v.LastName
}

// IDType is a tenant-configurable ID document type (National ID, Passport...).
type IDType struct {
	ID             int64
	TenantID       int64
	Name           string
	Code           string
	RequiresNumber bool
	Active         bool
}

// Company is a tenant-configurable visitor organization.
type Company struct {
	ID       int64
	TenantID int64
	Name     string
	Phone    *string
	Email    *string
	Address  *string
	Active   bool
}
