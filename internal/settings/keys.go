package settings

// Well-known setting keys. Centralizing key strings avoids configuration
// drift across modules.
const (
	KeyVisitorsAllowPreRegistration = "visitors.allow_pre_registration"
	KeyGatepassNumberPrefix         = "gatepass.number_prefix"
	KeyGatepassNumberUseYear        = "gatepass.number_use_year"

	// KeyTheme stores tenant branding and visual configuration as one JSON
	// document. The tenant database itself provides isolation.
	KeyTheme = "tenant.theme"
)