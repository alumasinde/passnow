package navigation

// DefaultRegistry defines the application's safe navigation surface.
// A definition is visible when the caller has at least one AnyPermissions code.
// Empty permission requirements mean the item is available to every authenticated
// tenant member, although current entries intentionally remain permission-aware.
func DefaultRegistry() []Definition {
	return []Definition{
		{
			Item: Item{
				Code: "dashboard", Label: "Dashboard", Icon: "fa-gauge",
				Href: "dashboard", MatchPrefixes: []string{"/dashboard"},
				Placement: "main", Order: 10,
			},
			AnyPermissions: []string{"dashboard.view"},
		},
		{
			Item: Item{
				Code: "visitors", Label: "Visitors", Icon: "fa-id-card",
				Href: "visitors", MatchPrefixes: []string{"/visitors", "/visitor"},
				Placement: "main", Order: 20,
			},
			AnyPermissions: []string{"visitors.view", "visitors.create", "visitors.update"},
		},
		{
			Item: Item{
				Code: "visits", Label: "Visits", Icon: "fa-calendar-check",
				Href: "visits", MatchPrefixes: []string{"/visits", "/visit"},
				Placement: "main", Order: 30,
			},
			AnyPermissions: []string{"visits.view", "visits.create", "visits.checkin", "visits.checkout", "visits.cancel"},
		},
		{
			Item: Item{
				Code: "gatepasses", Label: "Gate Passes", Icon: "fa-right-left",
				Href: "gatepasses", MatchPrefixes: []string{"/gatepasses", "/gatepass", "/gate-operations"},
				Placement: "main", Order: 40,
			},
			AnyPermissions: []string{"gatepasses.view", "gatepasses.create", "gatepasses.update", "gatepasses.approve", "gatepasses.reject", "gatepasses.issue", "gatepasses.verify", "gatepasses.cancel"},
		},
		{
			Item: Item{
				Code: "employees", Label: "Employees", Icon: "fa-users",
				Href: "employees", MatchPrefixes: []string{"/employees", "/employee"},
				Placement: "main", Order: 50,
			},
			AnyPermissions: []string{"employees.view", "employees.create", "employees.update"},
		},
		{
			Item: Item{
				Code: "approvals", Label: "Approvals", Icon: "fa-user-check",
				Href: "approvals", MatchPrefixes: []string{"/approvals", "/approval"},
				Placement: "main", Order: 60,
			},
			AnyPermissions: []string{"gatepasses.approve", "gatepasses.reject"},
		},
		{
			Item: Item{
				Code: "settings", Label: "Settings", Icon: "fa-gear",
				Href: "settings", MatchPrefixes: []string{
					"/settings", "/admin/users", "/users", "/user", "/admin/roles",
					"/roles", "/role-permissions", "/id-types", "/invitations",
					"/invite-user", "/visitor-settings", "/visitor-companies",
					"/visit-types", "/departments", "/gatepass-settings", "/gatepass-types",
					"/approval-workflows",
				},
				Placement: "bottom", Order: 10,
			},
			AnyPermissions: []string{
				"settings.users", "settings.roles", "settings.permissions",
				"settings.visitors", "settings.visits", "settings.gatepass",
				"settings.approvals",
			},
		},
	}
}
