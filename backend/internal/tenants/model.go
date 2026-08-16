package tenants

import "time"

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted" // soft delete
)

// Tenant is the domain/DB entity. Never expose this directly over the API —
// use tenants.DTO in the handler layer instead.
type Tenant struct {
	ID     int64
	Name   string
	Slug   string // used for path-prefix and subdomain resolution, e.g. "acme"
	Status Status

	// CustomDomain, if set, is a fully-qualified domain the tenant has
	// pointed at this platform (via CNAME/A record), e.g. "gate.acme.com".
	// Verified via CustomDomainVerifiedAt before it is trusted for TLS/
	// routing purposes.
	CustomDomain         *string
	CustomDomainVerified bool
	CustomDomainToken    string // DNS TXT verification token, e.g. gatepass-verify=xxxx

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive && t.DeletedAt == nil
}
