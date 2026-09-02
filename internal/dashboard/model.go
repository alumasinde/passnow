package dashboard

import "time"

// Summary is the internal tenant-scoped operational snapshot used by the
// dashboard widget engine. It is intentionally not the public dashboard
// contract; the public response only contains widgets the caller may see.
type Summary struct {
	VisitorsToday        int
	CurrentlyOnPremises  int
	ExpectedToday        int
	CompletedVisitsToday int

	ActiveGatepasses        int
	PendingApprovals        int
	RejectedGatepassesToday int
	EmployeeGatepassesToday int
	VisitorGatepassesToday  int

	ItemsEnteringToday int
	ItemsLeavingToday  int
	OverdueGatepasses  int

	RecentActivity []ActivityEntry
}

type ActivityEntry struct {
	Action     string `json:"action"`
	EntityType string `json:"entity_type"`
	EntityID   *int64 `json:"entity_id"`
	CreatedAt  string `json:"created_at"`
}

// Dashboard is the API-driven dashboard contract consumed by the frontend.
// The frontend renders these definitions and does not decide access itself.
type Dashboard struct {
	Widgets     []Widget     `json:"widgets"`
	GeneratedAt time.Time    `json:"generated_at"`
}

// Widget is a safe, backend-defined dashboard component.
type Widget struct {
	Code        string `json:"code"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Icon        string `json:"icon"`
	Accent      string `json:"accent"`
	Size        string `json:"size"`
	Order       int    `json:"order"`
	Value       any    `json:"value,omitempty"`
	Data        any    `json:"data,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// WidgetDefinition belongs to the server-side registry. SQL or executable
// widget logic is never stored in tenant configuration.
type WidgetDefinition struct {
	Code        string
	Type        string
	Title       string
	Icon        string
	Accent      string
	Size        string
	Order       int
	Permissions []string
	Build       func(*Summary) (value any, data any)
}
