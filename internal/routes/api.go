package routes

import (
	"net/http"
	"time"

	"gatepass/internal/approvals"
	"gatepass/internal/auth"
	"gatepass/internal/dashboard"
	"gatepass/internal/departments"
	"gatepass/internal/employees"
	"gatepass/internal/gatepasses"
	"gatepass/internal/invite"
	"gatepass/internal/media"
	"gatepass/internal/middleware"
	"gatepass/internal/navigation"
	"gatepass/internal/platform"
	"gatepass/internal/roles"
	"gatepass/internal/settings"
	"gatepass/internal/visitors"
	"gatepass/internal/visits"
)

// API contains the HTTP handlers and shared route dependencies used by the
// tenant-scoped API. Keeping this wiring here keeps cmd/api/main.go focused on
// process startup and shutdown.
type API struct {
	JWTSecret []byte
	RoleRepo  *roles.Repository

	AuthHandler             *auth.Handler
	VisitorHandler          *visitors.Handler
	VisitorSettingsHandler  *settings.VisitorSettingsHandler
	GatepassSettingsHandler *settings.GatepassSettingsHandler
	ThemeHandler            *settings.ThemeHandler
	MediaHandler            *media.Handler
	VisitTypeHandler        *visits.VisitTypeHandler
	DepartmentHandler       *departments.Handler
	VisitHandler            *visits.Handler
	WorkflowHandler         *approvals.Handler
	GatepassHandler         *gatepasses.Handler
	EmployeeHandler         *employees.Handler
	RoleHandler             *roles.Handler
	InviteHandler           *invite.Handler
	BootstrapHandler        *platform.Handler
	DashboardHandler        *dashboard.Handler
	NavigationHandler       *navigation.Handler

	LoginLimiter   *middleware.RateLimiter
	RefreshLimiter *middleware.RateLimiter
}

func NewAPI(jwtSecret []byte, roleRepo *roles.Repository) *API {
	return &API{
		JWTSecret:     jwtSecret,
		RoleRepo:      roleRepo,
		LoginLimiter:  middleware.NewRateLimiter(10, time.Minute),
		RefreshLimiter: middleware.NewRateLimiter(30, time.Minute),
	}
}

// RegisterAPI registers all tenant-scoped API endpoints.
func RegisterAPI(mux *http.ServeMux, api *API) {
	protected := func(permission string, h http.HandlerFunc) http.Handler {
		return middleware.Protected(api.JWTSecret, api.RoleRepo, permission, h)
	}

	// --- tenant theme (public read for branded login/application shell) ---
	mux.Handle("GET /api/v1/theme", http.HandlerFunc(api.ThemeHandler.Get))
	mux.Handle("PUT /api/v1/theme", protected("settings.theme", api.ThemeHandler.Update))

	// --- tenant media library ---
	mux.Handle("GET /api/v1/media/public/{publicID}", http.HandlerFunc(api.MediaHandler.Public))
	mux.Handle("POST /api/v1/media", protected("settings.media", api.MediaHandler.Upload))
	mux.Handle("GET /api/v1/media", protected("settings.media", api.MediaHandler.List))
	mux.Handle("DELETE /api/v1/media/{id}", protected("settings.media", api.MediaHandler.Delete))

	// --- auth ---
	mux.Handle("POST /api/v1/auth/login", api.LoginLimiter.Middleware("login")(http.HandlerFunc(api.AuthHandler.Login)))
	mux.Handle("POST /api/v1/auth/refresh", api.RefreshLimiter.Middleware("refresh")(http.HandlerFunc(api.AuthHandler.Refresh)))
	mux.Handle("POST /api/v1/auth/logout", middleware.Authenticated(api.JWTSecret, api.AuthHandler.Logout))

	// --- visitors ---
	mux.Handle("POST /api/v1/visitors", protected("visitors.create", api.VisitorHandler.Create))
	mux.Handle("GET /api/v1/visitors", protected("visitors.view", api.VisitorHandler.List))
	mux.Handle("GET /api/v1/visitors/{id}", protected("visitors.view", api.VisitorHandler.Get))
	mux.Handle("PATCH /api/v1/visitors/{id}", protected("visitors.update", api.VisitorHandler.Update))
	mux.Handle("POST /api/v1/visitors/{id}/blacklist", protected("visitors.update", api.VisitorHandler.SetBlacklist))

	// --- visitor configuration ---
	mux.Handle("GET /api/v1/id-types", protected("visitors.view", api.VisitorHandler.ListIDTypes))
	mux.Handle("GET /api/v1/id-types/{id}", protected("visitors.view", api.VisitorHandler.GetIDType))
	mux.Handle("POST /api/v1/id-types", protected("settings.visitors", api.VisitorHandler.CreateIDType))
	mux.Handle("PATCH /api/v1/id-types/{id}", protected("settings.visitors", api.VisitorHandler.UpdateIDType))
	mux.Handle("GET /api/v1/visitor-companies", protected("visitors.view", api.VisitorHandler.ListCompanies))
	mux.Handle("POST /api/v1/visitor-companies", protected("settings.visitors", api.VisitorHandler.CreateCompany))
	mux.Handle("PATCH /api/v1/visitor-companies/{id}", protected("settings.visitors", api.VisitorHandler.UpdateCompany))
	mux.Handle("GET /api/v1/visit-types", protected("visitors.view", api.VisitTypeHandler.List))
	mux.Handle("POST /api/v1/visit-types", protected("settings.visitors", api.VisitTypeHandler.Create))
	mux.Handle("PATCH /api/v1/visit-types/{id}", protected("settings.visitors", api.VisitTypeHandler.Update))
	mux.Handle("GET /api/v1/settings/visitors", protected("settings.visitors", api.VisitorSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/visitors", protected("settings.visitors", api.VisitorSettingsHandler.Update))

	// --- departments and visits ---
	mux.Handle("GET /api/v1/departments", protected("visits.view", api.DepartmentHandler.List))
	mux.Handle("GET /api/v1/departments/{id}", protected("visits.view", api.DepartmentHandler.Get))
	mux.Handle("POST /api/v1/departments", protected("settings.visits", api.DepartmentHandler.Create))
	mux.Handle("PATCH /api/v1/departments/{id}", protected("settings.visits", api.DepartmentHandler.Update))
	mux.Handle("POST /api/v1/visits", protected("visits.create", api.VisitHandler.Create))
	mux.Handle("GET /api/v1/visits", protected("visits.view", api.VisitHandler.List))
	mux.Handle("GET /api/v1/visits/{id}", protected("visits.view", api.VisitHandler.Get))
	mux.Handle("POST /api/v1/visits/{id}/check-in", protected("visits.checkin", api.VisitHandler.CheckIn))
	mux.Handle("POST /api/v1/visits/{id}/check-out", protected("visits.checkout", api.VisitHandler.CheckOut))
	mux.Handle("POST /api/v1/visits/{id}/cancel", protected("visits.cancel", api.VisitHandler.Cancel))
	mux.Handle("GET /api/v1/visits/badge/{token}", protected("visits.view", api.VisitHandler.BadgeLookup))

	// --- approval workflows ---
	mux.Handle("GET /api/v1/approval-workflows", protected("settings.approvals", api.WorkflowHandler.List))
	mux.Handle("GET /api/v1/approval-workflows/{id}", protected("settings.approvals", api.WorkflowHandler.Get))
	mux.Handle("POST /api/v1/approval-workflows", protected("settings.approvals", api.WorkflowHandler.Create))
	mux.Handle("PATCH /api/v1/approval-workflows/{id}", protected("settings.approvals", api.WorkflowHandler.Update))

	// --- gatepass types and operations ---
	// Opaque QR token is a high-entropy capability generated per gatepass.
	mux.Handle("GET /api/v1/gatepasses/qr/image/{token}", http.HandlerFunc(api.GatepassHandler.QRTokenImage))
	mux.Handle("GET /api/v1/gatepass-types", protected("gatepasses.view", api.GatepassHandler.ListTypes))
	mux.Handle("POST /api/v1/gatepass-types", protected("settings.gatepass", api.GatepassHandler.CreateType))
	mux.Handle("PATCH /api/v1/gatepass-types/{id}", protected("settings.gatepass", api.GatepassHandler.UpdateType))
	mux.Handle("POST /api/v1/gatepasses", protected("gatepasses.create", api.GatepassHandler.Create))
	mux.Handle("GET /api/v1/gatepasses", protected("gatepasses.view", api.GatepassHandler.List))
	mux.Handle("GET /api/v1/gatepasses/{id}", protected("gatepasses.view", api.GatepassHandler.Get))
	mux.Handle("POST /api/v1/gatepasses/{id}/cancel", protected("gatepasses.cancel", api.GatepassHandler.Cancel))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/approve", protected("gatepasses.approve", api.GatepassHandler.Approve))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/reject", protected("gatepasses.reject", api.GatepassHandler.Reject))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-out", protected("gatepasses.issue", api.GatepassHandler.CheckOut))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-in", protected("gatepasses.verify", api.GatepassHandler.CheckIn))
	mux.Handle("GET /api/v1/gatepasses/{id}/movements", protected("gatepasses.view", api.GatepassHandler.Movements))
	mux.Handle("GET /api/v1/gatepasses/{id}/qr.png", protected("gatepasses.view", api.GatepassHandler.QRImage))
	mux.Handle("GET /api/v1/gatepasses/qr/token/{token}", protected("gatepasses.view", api.GatepassHandler.QRLookup))
	mux.Handle("GET /api/v1/settings/gatepass", protected("settings.gatepass", api.GatepassSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/gatepass", protected("settings.gatepass", api.GatepassSettingsHandler.Update))

	// --- employees ---
	mux.Handle("GET /api/v1/employees", protected("employees.view", api.EmployeeHandler.List))
	mux.Handle("GET /api/v1/employees/{id}", protected("employees.view", api.EmployeeHandler.Get))
	mux.Handle("POST /api/v1/employees", protected("employees.create", api.EmployeeHandler.Create))
	mux.Handle("PATCH /api/v1/employees/{id}", protected("employees.update", api.EmployeeHandler.Update))

	// --- roles, users and invitations ---
	mux.Handle("GET /api/v1/permissions", protected("settings.roles", api.RoleHandler.ListPermissions))
	mux.Handle("GET /api/v1/roles", protected("settings.roles", api.RoleHandler.ListRoles))
	mux.Handle("POST /api/v1/roles", protected("settings.roles", api.RoleHandler.CreateRole))
	mux.Handle("PUT /api/v1/roles/{id}/permissions", protected("settings.permissions", api.RoleHandler.SetRolePermissions))
	mux.Handle("GET /api/v1/users", protected("settings.users", api.RoleHandler.ListUsers))
	mux.Handle("POST /api/v1/users/invite", protected("settings.users", api.InviteHandler.Invite))
	mux.Handle("PATCH /api/v1/users/memberships/{id}", protected("settings.users", api.RoleHandler.UpdateUserMembership))

	// --- dynamic navigation ---
	mux.Handle("GET /api/v1/navigation", middleware.Authenticated(api.JWTSecret, api.NavigationHandler.List))

	// --- dashboard and personal approval queue ---
	mux.Handle("GET /api/v1/dashboard", protected("dashboard.view", api.DashboardHandler.Dashboard))
	mux.Handle("GET /api/v1/dashboard/summary", protected("dashboard.view", api.DashboardHandler.Summary))
	mux.Handle("GET /api/v1/approvals/pending", protected("gatepasses.approve", api.GatepassHandler.MyPendingApprovals))
}
