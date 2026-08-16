package dashboard

// Summary is a snapshot of tenant-scoped operational metrics. Every field
// is computed from a query filtered by tenant_id — never a global/
// unscoped count — matching the spec requirement that dashboard numbers
// must never leak across tenants.
type Summary struct {
	// Visits
	VisitorsToday        int `json:"visitors_today"`
	CurrentlyOnPremises  int `json:"currently_on_premises"`
	ExpectedToday        int `json:"expected_today"`
	CompletedVisitsToday int `json:"completed_visits_today"`

	// Gatepasses
	ActiveGatepasses        int `json:"active_gatepasses"`
	PendingApprovals        int `json:"pending_approvals"`
	RejectedGatepassesToday int `json:"rejected_gatepasses_today"`
	EmployeeGatepassesToday int `json:"employee_gatepasses_today"`
	VisitorGatepassesToday  int `json:"visitor_gatepasses_today"`

	// Items (approximated by the gatepass's creation date — an item
	// doesn't have its own "occurred at" timestamp distinct from its
	// parent gatepass)
	ItemsEnteringToday int `json:"items_entering_today"`
	ItemsLeavingToday  int `json:"items_leaving_today"`

	// Security
	// OverdueGatepasses: returnable passes checked out past their
	// expected_return_at with no check-in yet — the one genuinely
	// actionable "security alert" the current data model supports.
	OverdueGatepasses int `json:"overdue_gatepasses"`

	RecentActivity []ActivityEntry `json:"recent_activity"`
}

type ActivityEntry struct {
	Action     string `json:"action"`
	EntityType string `json:"entity_type"`
	EntityID   *int64 `json:"entity_id"`
	CreatedAt  string `json:"created_at"`
}
