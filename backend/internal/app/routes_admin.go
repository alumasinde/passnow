package app

import (
	"net/http"

	"gatepass/internal/middleware"
)

func registerEmployeeRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/employees", middleware.Protected(c.JWTSecret, c.RoleRepo, "employees.view", c.EmployeeHandler.List))
	mux.Handle("GET /api/v1/employees/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "employees.view", c.EmployeeHandler.Get))
	mux.Handle("POST /api/v1/employees", middleware.Protected(c.JWTSecret, c.RoleRepo, "employees.create", c.EmployeeHandler.Create))
	mux.Handle("PATCH /api/v1/employees/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "employees.update", c.EmployeeHandler.Update))
}

func registerSettingsRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/settings/visitors", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/visitors", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorSettingsHandler.Update))
	mux.Handle("GET /api/v1/settings/gatepass", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.gatepass", c.GatepassSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/gatepass", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.gatepass", c.GatepassSettingsHandler.Update))
}

func registerUserAndRoleRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/permissions", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.roles", c.RoleHandler.ListPermissions))
	mux.Handle("GET /api/v1/roles", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.roles", c.RoleHandler.ListRoles))
	mux.Handle("POST /api/v1/roles", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.roles", c.RoleHandler.CreateRole))
	mux.Handle("PUT /api/v1/roles/{id}/permissions", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.permissions", c.RoleHandler.SetRolePermissions))

	mux.Handle("GET /api/v1/users", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.users", c.RoleHandler.ListUsers))
	mux.Handle("POST /api/v1/users/invite", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.users", c.InviteHandler.Invite))
	mux.Handle("PATCH /api/v1/users/memberships/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.users", c.RoleHandler.UpdateUserMembership))
}

func registerDashboardRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/dashboard/summary", middleware.Protected(c.JWTSecret, c.RoleRepo, "dashboard.view", c.DashboardHandler.Summary))
}
