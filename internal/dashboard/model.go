package dashboard

import "time"

// Summary is the internal tenant-scoped operational snapshot used by the
// dashboard widget engine. It is intentionally not the public dashboard
// contract; the public response only contains widgets the caller may see.
type Summary struct {
	VisitorsToday        int `json:"visitors_today"`
	CurrentlyOnPremises  int `json:"currently_on_premises"`
	CheckedInToday       int `json:"checked_in_today"`
	ExpectedToday        int `json:"expected_today"`
	CompletedVisitsToday int `json:"completed_visits_today"`

	ActiveGatepasses        int `json:"active_gatepasses"`
	PendingApprovals        int `json:"pending_approvals"`
	RejectedGatepassesToday int `json:"rejected_gatepasses_today"`
	EmployeeGatepassesToday int `json:"employee_gatepasses_today"`
	VisitorGatepassesToday  int `json:"visitor_gatepasses_today"`

	ItemsEnteringToday int `json:"items_entering_today"`
	ItemsLeavingToday  int `json:"items_leaving_today"`
	OverdueGatepasses  int `json:"overdue_gatepasses"`

	RecentActivity []ActivityEntry `json:"recent_activity"`
	GatepassTypeBreakdown []BreakdownEntry `json:"gatepass_type_breakdown"`
}

type BreakdownEntry struct {
	Key string `json:"key"`
	Label string `json:"label"`
	Value int `json:"value"`
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
