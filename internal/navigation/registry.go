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
				Href: "visitors", MatchPrefixes: []string{"/visitors", "/visitors.php", "/visitor", "/visitor.php", "/visitor-create.php", "/visitor-blacklist.php", "/visitor-companies.php", "/visitor-companies-edit.php"},
				Placement: "main", Order: 20,
			},
			AnyPermissions: []string{"visitors.view", "visitors.create", "visitors.update"},
		},
		{
			Item: Item{
				Code: "visits", Label: "Visits", Icon: "fa-calendar-check",
				Href: "visits", MatchPrefixes: []string{"/visits", "/visits.php", "/visit", "/visit.php", "/visit-create.php", "/visit-operation.php", "/visit-types.php", "/visit-types-edit.php"},
				Placement: "main", Order: 30,
			},
			AnyPermissions: []string{"visits.view", "visits.create", "visits.checkin", "visits.checkout", "visits.cancel"},
		},
		{
			Item: Item{
				Code: "gatepasses", Label: "Gate Passes", Icon: "fa-right-left",
				Href: "gatepasses", MatchPrefixes: []string{"/gatepasses", "/gatepasses.php", "/gatepass", "/gatepass.php", "/gatepass-create.php", "/gatepass-operation.php", "/gate-operations.php", "/gatepass-qr.php", "/gatepass-settings.php", "/gatepass-types.php", "/gatepass-types-edit.php"},
				Placement: "main", Order: 40,
			},
			AnyPermissions: []string{"gatepasses.view", "gatepasses.create", "gatepasses.update", "gatepasses.approve", "gatepasses.reject", "gatepasses.issue", "gatepasses.verify", "gatepasses.cancel"},
		},
		{
			Item: Item{
				Code: "employees", Label: "Employees", Icon: "fa-users",
				Href: "employees", MatchPrefixes: []string{"/employees", "/employees.php", "/employee", "/employee.php", "/employee-create.php"},
				Placement: "main", Order: 50,
			},
			AnyPermissions: []string{"employees.view", "employees.create", "employees.update"},
		},
		{
			Item: Item{
				Code: "approvals", Label: "Approvals", Icon: "fa-user-check",
				Href: "approvals", MatchPrefixes: []string{"/approvals", "/approvals.php", "/approval", "/approval.php", "/approval-decision.php", "/approval-workflows.php", "/approval-workflow-edit.php"},
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
