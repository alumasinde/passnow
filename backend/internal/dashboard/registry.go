package dashboard

// DefaultRegistry returns the safe, server-defined dashboard widgets.
// Permissions are checked dynamically for the current caller on every request.
func DefaultRegistry() []WidgetDefinition {
	return []WidgetDefinition{
		{
			Code: "visitors_today", Type: "stat", Title: "Today's Visitors",
			Icon: "users", Accent: "primary", Size: "sm", Order: 10,
			Permissions: []string{"visitors.view"},
			Build: func(s *Summary) (any, any) { return s.VisitorsToday, nil },
		},
		{
			Code: "active_visits", Type: "stat", Title: "Active Visits",
			Icon: "door-open", Accent: "success", Size: "sm", Order: 20,
			Permissions: []string{"visits.view"},
			Build: func(s *Summary) (any, any) { return s.CurrentlyOnPremises, nil },
		},
		{
			Code: "gatepasses_issued_today", Type: "stat", Title: "Gate Pass Issued",
			Icon: "file-text", Accent: "info", Size: "sm", Order: 30,
			Permissions: []string{"gatepasses.view"},
			Build: func(s *Summary) (any, any) {
				return s.EmployeeGatepassesToday + s.VisitorGatepassesToday, nil
			},
		},
		{
			Code: "pending_approvals", Type: "stat", Title: "Pending Approvals",
			Icon: "hourglass", Accent: "warning", Size: "sm", Order: 40,
			Permissions: []string{"gatepasses.approve"},
			Build: func(s *Summary) (any, any) { return s.PendingApprovals, nil },
		},
		{
			Code: "checked_in_today", Type: "stat", Title: "Checked In",
			Icon: "log-in", Accent: "success", Size: "sm", Order: 50,
			Permissions: []string{"visits.view"},
			Build: func(s *Summary) (any, any) { return s.CheckedInToday, nil },
		},
		{
			Code: "checked_out_today", Type: "stat", Title: "Checked Out",
			Icon: "log-out", Accent: "danger", Size: "sm", Order: 60,
			Permissions: []string{"visits.view"},
			Build: func(s *Summary) (any, any) { return s.CompletedVisitsToday, nil },
		},
		{
			Code: "overstayed_visits", Type: "stat", Title: "Overstayed Visits",
			Icon: "clock", Accent: "warning", Size: "sm", Order: 70,
			Permissions: []string{"visits.view"},
			Build: func(s *Summary) (any, any) { return s.OverstayedVisits, nil },
		},
		{
			Code: "gatepass_overview", Type: "chart", Title: "Gate Pass Overview",
			Icon: "chart-line", Accent: "primary", Size: "lg", Order: 100,
			Permissions: []string{"gatepasses.view"},
			Build: func(s *Summary) (any, any) {
				return nil, map[string]any{
					"series": []map[string]any{
						{"key": "issued", "label": "Issued", "value": s.EmployeeGatepassesToday + s.VisitorGatepassesToday},
						{"key": "active", "label": "Active", "value": s.ActiveGatepasses},
						{"key": "pending", "label": "Pending", "value": s.PendingApprovals},
					},
				}
			},
		},
		{
			Code: "gatepass_by_type", Type: "breakdown", Title: "Gate Pass by Type",
			Icon: "pie-chart", Accent: "primary", Size: "lg", Order: 110,
			Permissions: []string{"gatepasses.view"},
			Build: func(s *Summary) (any, any) {
				return s.EmployeeGatepassesToday + s.VisitorGatepassesToday, map[string]any{
					"series": s.GatepassTypeBreakdown,
				}
			},
		},
		{
			Code: "recent_activity", Type: "activity", Title: "Recent Activity",
			Icon: "activity", Accent: "primary", Size: "lg", Order: 120,
			Permissions: []string{"dashboard.view"},
			Build: func(s *Summary) (any, any) { return nil, s.RecentActivity },
		},

		{
			Code: "quick_actions", Type: "actions", Title: "Quick Actions",
			Icon: "bolt", Accent: "primary", Size: "lg", Order: 130,
			Permissions: []string{"dashboard.view"},
			Build: func(s *Summary) (any, any) {
				return nil, []map[string]any{
					{"label": "Manage Gate Passes", "icon": "right-left", "route": "gatepasses.php"},
					{"label": "Review Approvals", "icon": "user-check", "route": "approvals.php"},
					{"label": "Manage Visitors", "icon": "user-plus", "route": "visitors.php"},
				}
			},
		},

		{
			Code: "overdue_gatepasses", Type: "stat", Title: "Overdue Gatepasses",
			Icon: "alert-triangle", Accent: "danger", Size: "sm", Order: 140,
			Permissions: []string{"gatepasses.view"},
			Build: func(s *Summary) (any, any) { return s.OverdueGatepasses, nil },
		},
	}
}
