package settings

// Well-known setting keys. Centralizing the key STRINGS here (even though
// the values live in a generic table) avoids typos scattered across
// packages — "visitors.allow_pre_registration" typed in two different
// files that drift apart is exactly the kind of bug this avoids.
const (
	// KeyVisitorsAllowPreRegistration gates whether staff can register a
	// visitor with source=pre_registered. Off by default — Platform Admin
	// opts a tenant in.
	KeyVisitorsAllowPreRegistration = "visitors.allow_pre_registration"

	// KeyGatepassNumberPrefix configures the prefix used for auto-generated
	// pass numbers, e.g. "GP" -> "GP-2026-000001", or "ACME-GP" ->
	// "ACME-GP-2026-000001". Platform Admin sets this per tenant.
	KeyGatepassNumberPrefix = "gatepass.number_prefix"
	// KeyGatepassNumberUseYear controls whether the sequence resets each
	// calendar year (true, default) or runs continuously (false).
	KeyGatepassNumberUseYear = "gatepass.number_use_year"
)
