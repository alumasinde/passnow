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
			AnyPermissions: []string{"report.read.own", "report.read.department", "report.read.all"},
		},
		{
			Item: Item{
				Code: "visitors", Label: "Visitors", Icon: "fa-id-card",
				Href: "visitors", MatchPrefixes: []string{"/visitors", "/visitors.php", "/visitor", "/visitor.php", "/visitor-create.php", "/visitor-blacklist.php", "/visitor-companies.php", "/visitor-companies-edit.php"},
				Placement: "main", Order: 20,
			},
			AnyPermissions: []string{"visitor.read.own", "visitor.read.department", "visitor.read.all", "visitor.create", "visitor.update.own", "visitor.update.department", "visitor.update.all"},
		},
		{
			Item: Item{
				Code: "visits", Label: "Visits", Icon: "fa-calendar-check",
				Href: "visits", MatchPrefixes: []string{"/visits", "/visits.php", "/visit", "/visit.php", "/visit-create.php", "/visit-operation.php", "/visit-types.php", "/visit-types-edit.php"},
				Placement: "main", Order: 30,
			},
			AnyPermissions: []string{"visit.read.own", "visit.read.department", "visit.read.all", "visit.create", "visit.check_in", "visit.check_out", "visit.cancel.own", "visit.cancel.department", "visit.cancel.all"},
		},
		{
			Item: Item{
				Code: "gatepasses", Label: "Gate Passes", Icon: "fa-right-left",
				Href: "gatepasses", MatchPrefixes: []string{"/gatepasses", "/gatepasses.php", "/gatepass", "/gatepass.php", "/gatepass-create.php", "/gatepass-operation.php", "/gate-operations.php", "/gatepass-qr.php", "/gatepass-settings.php", "/gatepass-types.php", "/gatepass-types-edit.php"},
				Placement: "main", Order: 40,
			},
			AnyPermissions: []string{"gatepass.read.own", "gatepass.read.department", "gatepass.read.all", "gatepass.create", "gatepass.update.own", "gatepass.update.department", "gatepass.update.all", "gatepass.check_out", "gatepass.check_in", "gatepass.verify", "gatepass.cancel.own", "gatepass.cancel.department", "gatepass.cancel.all"},
		},
		{
			Item: Item{
				Code: "employees", Label: "Employees", Icon: "fa-users",
				Href: "employees", MatchPrefixes: []string{"/employees", "/employees.php", "/employee", "/employee.php", "/employee-create.php"},
				Placement: "main", Order: 50,
			},
			AnyPermissions: []string{"employee.read.department", "employee.read.all", "employee.create", "employee.update.department", "employee.update.all"},
		},
		{
			Item: Item{
				Code: "approvals", Label: "Approvals", Icon: "fa-user-check",
				Href: "approvals", MatchPrefixes: []string{"/approvals", "/approvals.php", "/approval", "/approval.php", "/approval-decision.php", "/approval-workflows.php", "/approval-workflow-edit.php"},
				Placement: "main", Order: 60,
			},
			AnyPermissions: []string{"approval.read.assigned", "approval.read.department", "approval.read.all", "approval.approve", "approval.reject"},
		},
		{
			Item: Item{
				Code: "settings", Label: "Settings", Icon: "fa-gear",
				Href: "settings", MatchPrefixes: []string{
					"/settings", "/admin/users", "/users", "/user", "/admin/roles",
					"/roles", "/role-permissions", "/id-types", "/invitations",
					"/invite-user", "/visitor-settings", "/visitor-companies",
					"/visit-types", "/departments", "/gatepass-settings", "/gatepass-types",
					"/approval-workflows", "/gates", "/gates.php", "/gates/edit", "/gates-edit.php",
				},
				Placement: "bottom", Order: 10,
			},
			AnyPermissions: []string{
				"user.read.all", "user.create", "user.update.all",
				"role.read", "role.create", "role.update", "permission.read", "permission.assign",
				"department.read", "department.create", "department.update",
				"workflow.read", "workflow.create", "workflow.update", "workflow.activate",
				"gate.read", "gate.create", "gate.update",
			},
		},
	}
}
