package app

import (
	"context"
	"net/http"
	"time"

	"gatepass/internal/middleware"
)

// RegisterRoutes builds the complete HTTP router from the application
// container. Keeping route registration here makes the bootstrap lifecycle
// independent from endpoint definitions while preserving the existing API.
func RegisterRoutes(c *Container) http.Handler {
	rootMux := http.NewServeMux()
	tenantMux := http.NewServeMux()

	rootMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	rootMux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := c.DB.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	rootMux.HandleFunc("POST /api/v1/platform/bootstrap-tenant", c.BootstrapHandler.BootstrapTenant)

	registerAuthRoutes(tenantMux, c)
	registerVisitorRoutes(tenantMux, c)
	registerVisitRoutes(tenantMux, c)
	registerApprovalRoutes(tenantMux, c)
	registerGatepassRoutes(tenantMux, c)
	registerEmployeeRoutes(tenantMux, c)
	registerSettingsRoutes(tenantMux, c)
	registerUserAndRoleRoutes(tenantMux, c)
	registerDashboardRoutes(tenantMux, c)

	rootMux.Handle("/", middleware.ResolveTenant(c.TenantRepo, c.Config.BaseDomain)(tenantMux))
	return rootMux
}

func registerAuthRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("POST /api/v1/auth/login", c.LoginLimiter.Middleware("login")(http.HandlerFunc(c.AuthHandler.Login)))
	mux.Handle("POST /api/v1/auth/refresh", c.RefreshLimiter.Middleware("refresh")(http.HandlerFunc(c.AuthHandler.Refresh)))
	mux.Handle("POST /api/v1/auth/logout", middleware.Authenticated(c.JWTSecret, c.AuthHandler.Logout))
}

func registerVisitorRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("POST /api/v1/visitors", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.create", c.VisitorHandler.Create))
	mux.Handle("GET /api/v1/visitors", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.List))
	mux.Handle("GET /api/v1/visitors/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.Get))
	mux.Handle("PATCH /api/v1/visitors/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.update", c.VisitorHandler.Update))
	mux.Handle("POST /api/v1/visitors/{id}/blacklist", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.update", c.VisitorHandler.SetBlacklist))

	mux.Handle("GET /api/v1/id-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.ListIDTypes))
	mux.Handle("POST /api/v1/id-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.CreateIDType))
	mux.Handle("PATCH /api/v1/id-types/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.UpdateIDType))

	mux.Handle("GET /api/v1/visitor-companies", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.ListCompanies))
	mux.Handle("POST /api/v1/visitor-companies", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.CreateCompany))
	mux.Handle("PATCH /api/v1/visitor-companies/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.UpdateCompany))
}

func registerVisitRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/visit-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitTypeHandler.List))
	mux.Handle("POST /api/v1/visit-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitTypeHandler.Create))
	mux.Handle("PATCH /api/v1/visit-types/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitTypeHandler.Update))

	mux.Handle("GET /api/v1/departments", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.DepartmentHandler.List))
	mux.Handle("POST /api/v1/departments", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visits", c.DepartmentHandler.Create))
	mux.Handle("PATCH /api/v1/departments/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visits", c.DepartmentHandler.Update))

	mux.Handle("POST /api/v1/visits", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.create", c.VisitHandler.Create))
	mux.Handle("GET /api/v1/visits", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.VisitHandler.List))
	mux.Handle("GET /api/v1/visits/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.VisitHandler.Get))
	mux.Handle("POST /api/v1/visits/{id}/check-in", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.checkin", c.VisitHandler.CheckIn))
	mux.Handle("POST /api/v1/visits/{id}/check-out", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.checkout", c.VisitHandler.CheckOut))
	mux.Handle("POST /api/v1/visits/{id}/cancel", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.cancel", c.VisitHandler.Cancel))
	mux.Handle("GET /api/v1/visits/badge/{token}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.VisitHandler.BadgeLookup))
}

func registerApprovalRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/approval-workflows", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.approvals", c.WorkflowHandler.List))
	mux.Handle("GET /api/v1/approval-workflows/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.approvals", c.WorkflowHandler.Get))
	mux.Handle("POST /api/v1/approval-workflows", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.approvals", c.WorkflowHandler.Create))
}

func registerGatepassRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/gatepass-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.ListTypes))
	mux.Handle("POST /api/v1/gatepass-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.gatepass", c.GatepassHandler.CreateType))
	mux.Handle("PATCH /api/v1/gatepass-types/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.gatepass", c.GatepassHandler.UpdateType))

	mux.Handle("POST /api/v1/gatepasses", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.create", c.GatepassHandler.Create))
	mux.Handle("GET /api/v1/gatepasses", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.List))
	mux.Handle("GET /api/v1/gatepasses/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.Get))
	mux.Handle("POST /api/v1/gatepasses/{id}/cancel", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.cancel", c.GatepassHandler.Cancel))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/approve", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.approve", c.GatepassHandler.Approve))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/reject", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.reject", c.GatepassHandler.Reject))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-out", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.issue", c.GatepassHandler.CheckOut))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-in", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.verify", c.GatepassHandler.CheckIn))
	mux.Handle("GET /api/v1/gatepasses/{id}/movements", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.Movements))
	mux.Handle("GET /api/v1/gatepasses/{id}/qr.png", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.QRImage))
	mux.Handle("GET /api/v1/gatepasses/qr/{token}", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.QRLookup))

	mux.Handle("GET /api/v1/approvals/pending", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.approve", c.GatepassHandler.MyPendingApprovals))
}

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
