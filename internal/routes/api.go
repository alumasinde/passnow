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
	"gatepass/internal/gates"
	"gatepass/internal/gatedevices"
	"gatepass/internal/invite"
	"gatepass/internal/media"
	"gatepass/internal/middleware"
	"gatepass/internal/navigation"
	"gatepass/internal/platform"
	"gatepass/internal/roles"
	"gatepass/internal/rbac"
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
	RBAC      *rbac.Engine

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
	GateHandler             *gates.Handler
	GateDeviceHandler       *gatedevices.Handler
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
		RBAC:          rbac.New(roleRepo),
		LoginLimiter:  middleware.NewRateLimiter(10, time.Minute),
		RefreshLimiter: middleware.NewRateLimiter(30, time.Minute),
	}
}

// RegisterAPI registers all tenant-scoped API endpoints.
func RegisterAPI(mux *http.ServeMux, api *API) {
	protected := func(permission string, h http.HandlerFunc) http.Handler {
		return middleware.ProtectedRBAC(api.JWTSecret, api.RBAC, permission, h)
	}
	protectedAny := func(permissions []string, h http.HandlerFunc) http.Handler {
		return middleware.ProtectedRBACAny(api.JWTSecret, api.RBAC, permissions, h)
	}

	// --- tenant theme (public read for branded login/application shell) ---
	mux.Handle("GET /api/v1/theme", http.HandlerFunc(api.ThemeHandler.Get))
	mux.Handle("PUT /api/v1/theme", protected("permission.read", api.ThemeHandler.Update))

	// --- tenant media library ---
	mux.Handle("GET /api/v1/media/public/{publicID}", http.HandlerFunc(api.MediaHandler.Public))
	mux.Handle("POST /api/v1/media", protected("permission.read", api.MediaHandler.Upload))
	mux.Handle("GET /api/v1/media", protected("permission.read", api.MediaHandler.List))
	mux.Handle("DELETE /api/v1/media/{id}", protected("permission.read", api.MediaHandler.Delete))

	// --- auth ---
	mux.Handle("POST /api/v1/auth/login", api.LoginLimiter.Middleware("login")(http.HandlerFunc(api.AuthHandler.Login)))
	mux.Handle("POST /api/v1/auth/refresh", api.RefreshLimiter.Middleware("refresh")(http.HandlerFunc(api.AuthHandler.Refresh)))
	mux.Handle("POST /api/v1/auth/logout", middleware.Authenticated(api.JWTSecret, api.AuthHandler.Logout))
	mux.Handle("POST /api/v1/auth/change-password", middleware.Authenticated(api.JWTSecret, api.AuthHandler.ChangePassword))
	mux.Handle("GET /api/v1/auth/me", middleware.Authenticated(api.JWTSecret, api.AuthHandler.Me))
	mux.Handle("PATCH /api/v1/auth/me", middleware.Authenticated(api.JWTSecret, api.AuthHandler.UpdateProfile))

	// --- visitors ---
	mux.Handle("POST /api/v1/visitors", protected("visitor.create", api.VisitorHandler.Create))
	mux.Handle("GET /api/v1/visitors", protected("visitor.read", api.VisitorHandler.List))
	mux.Handle("GET /api/v1/visitors/{id}", protected("visitor.read", api.VisitorHandler.Get))
	mux.Handle("PATCH /api/v1/visitors/{id}", protected("visitor.update", api.VisitorHandler.Update))
	mux.Handle("POST /api/v1/visitors/{id}/blacklist", protected("visitor.update", api.VisitorHandler.SetBlacklist))

	// --- visitor configuration ---
	mux.Handle("GET /api/v1/id-types", protected("visitor.read.all", api.VisitorHandler.ListIDTypes))
	mux.Handle("GET /api/v1/id-types/{id}", protected("visitor.read.all", api.VisitorHandler.GetIDType))
	mux.Handle("POST /api/v1/id-types", protected("visitor.update.all", api.VisitorHandler.CreateIDType))
	mux.Handle("PATCH /api/v1/id-types/{id}", protected("visitor.update.all", api.VisitorHandler.UpdateIDType))
	mux.Handle("GET /api/v1/visitor-companies", protected("visitor.read.all", api.VisitorHandler.ListCompanies))
	mux.Handle("GET /api/v1/visitor-companies/{id}", protected("visitor.read.all", api.VisitorHandler.GetCompany))
	mux.Handle("POST /api/v1/visitor-companies", protected("visitor.update.all", api.VisitorHandler.CreateCompany))
	mux.Handle("PATCH /api/v1/visitor-companies/{id}", protected("visitor.update.all", api.VisitorHandler.UpdateCompany))
	mux.Handle("GET /api/v1/visit-types", protected("visitor.read.all", api.VisitTypeHandler.List))
	mux.Handle("GET /api/v1/visit-types/{id}", protected("visitor.read.all", api.VisitTypeHandler.Get))
	mux.Handle("POST /api/v1/visit-types", protected("visitor.update.all", api.VisitTypeHandler.Create))
	mux.Handle("PATCH /api/v1/visit-types/{id}", protected("visitor.update.all", api.VisitTypeHandler.Update))
	mux.Handle("GET /api/v1/settings/visitors", protected("visitor.update.all", api.VisitorSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/visitors", protected("visitor.update.all", api.VisitorSettingsHandler.Update))

	// --- departments and visits ---
	mux.Handle("GET /api/v1/departments", protected("visit.read.all", api.DepartmentHandler.List))
	mux.Handle("GET /api/v1/departments/{id}", protected("visit.read.all", api.DepartmentHandler.Get))
	mux.Handle("POST /api/v1/departments", protected("department.update", api.DepartmentHandler.Create))
	mux.Handle("PATCH /api/v1/departments/{id}", protected("department.update", api.DepartmentHandler.Update))
	mux.Handle("POST /api/v1/visits", protected("visit.create", api.VisitHandler.Create))
	mux.Handle("GET /api/v1/visits", protected("visit.read", api.VisitHandler.List))
	mux.Handle("GET /api/v1/visits/{id}", protected("visit.read", api.VisitHandler.Get))
	mux.Handle("POST /api/v1/visits/{id}/check-in", protected("visit.check_in", api.VisitHandler.CheckIn))
	mux.Handle("POST /api/v1/visits/{id}/check-out", protected("visit.check_out", api.VisitHandler.CheckOut))
	mux.Handle("POST /api/v1/visits/{id}/cancel", protected("visit.cancel.all", api.VisitHandler.Cancel))
	mux.Handle("GET /api/v1/visits/badge/{token}", protected("visit.read.all", api.VisitHandler.BadgeLookup))

	// --- approval workflows ---
	mux.Handle("GET /api/v1/approval-workflows", protected("workflow.read", api.WorkflowHandler.List))
	mux.Handle("GET /api/v1/approval-workflows/{id}", protected("workflow.read", api.WorkflowHandler.Get))
	mux.Handle("POST /api/v1/approval-workflows", protected("workflow.read", api.WorkflowHandler.Create))
	mux.Handle("PATCH /api/v1/approval-workflows/{id}", protected("workflow.read", api.WorkflowHandler.Update))

	// --- gates ---
	mux.Handle("GET /api/v1/gates", protected("gate.read", api.GateHandler.List))
	mux.Handle("GET /api/v1/gates/{id}", protected("gate.read", api.GateHandler.Get))
	mux.Handle("POST /api/v1/gates", protected("gate.create", api.GateHandler.Create))
	mux.Handle("PATCH /api/v1/gates/{id}", protected("gate.update", api.GateHandler.Update))

	// --- authorized gate devices ---
	mux.Handle("GET /api/v1/gate-devices", protected("gate.read", api.GateDeviceHandler.List))
	mux.Handle("POST /api/v1/gate-devices", protected("gate.create", api.GateDeviceHandler.Create))

	// --- gatepass types and operations ---
	// Opaque QR token is a high-entropy capability generated per gatepass.
	mux.Handle("GET /api/v1/gatepasses/qr/image/{token}", http.HandlerFunc(api.GatepassHandler.QRTokenImage))
	mux.Handle("GET /api/v1/gatepass-types", protected("gatepass.read.all", api.GatepassHandler.ListTypes))
	mux.Handle("GET /api/v1/gatepass-types/{id}", protected("gatepass.read.all", api.GatepassHandler.GetType))
	mux.Handle("POST /api/v1/gatepass-types", protected("gatepass.update.all", api.GatepassHandler.CreateType))
	mux.Handle("PATCH /api/v1/gatepass-types/{id}", protected("gatepass.update.all", api.GatepassHandler.UpdateType))
	mux.Handle("POST /api/v1/gatepasses", protected("gatepass.create", api.GatepassHandler.Create))
	mux.Handle("GET /api/v1/gatepasses", protected("gatepass.read", api.GatepassHandler.List))
	mux.Handle("GET /api/v1/gatepasses/{id}", protected("gatepass.read", api.GatepassHandler.Get))
	mux.Handle("POST /api/v1/gatepasses/{id}/cancel", protected("gatepass.cancel", api.GatepassHandler.Cancel))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/approve", protected("approval.approve.assigned", api.GatepassHandler.Approve))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/reject", protected("approval.reject.assigned", api.GatepassHandler.Reject))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-out", protected("gatepass.check_out", api.GatepassHandler.CheckOut))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-in", protected("gatepass.verify", api.GatepassHandler.CheckIn))
	mux.Handle("POST /api/v1/gatepasses/qr/token/{token}/check-out", protected("gatepass.check_out", api.GatepassHandler.QRCheckOut))
	mux.Handle("POST /api/v1/gatepasses/qr/token/{token}/check-in", protected("gatepass.verify", api.GatepassHandler.QRCheckIn))
	mux.Handle("GET /api/v1/gatepasses/{id}/movements", protected("gatepass.read.all", api.GatepassHandler.Movements))
	mux.Handle("GET /api/v1/gatepasses/{id}/qr.png", protected("gatepass.read.all", api.GatepassHandler.QRImage))
	mux.Handle("GET /api/v1/gatepasses/qr/token/{token}", protected("gatepass.read.all", api.GatepassHandler.QRLookup))
	mux.Handle("GET /api/v1/settings/gatepass", protected("gatepass.update.all", api.GatepassSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/gatepass", protected("gatepass.update.all", api.GatepassSettingsHandler.Update))

	// --- employees ---
	// A visit creator needs the employee directory only to select a host. This does
	// not grant employee-management access; either employees.view OR visits.create
	// is sufficient for this read-only lookup.
	mux.Handle("GET /api/v1/employees", protectedAny([]string{"employee.read.all", "employee.read.department", "visit.create"}, api.EmployeeHandler.List))
	mux.Handle("GET /api/v1/employees/{id}", protected("employee.read", api.EmployeeHandler.Get))
	mux.Handle("POST /api/v1/employees", protected("employee.create", api.EmployeeHandler.Create))
	mux.Handle("PATCH /api/v1/employees/{id}", protected("employee.update", api.EmployeeHandler.Update))

	// --- roles, users and invitations ---
	mux.Handle("GET /api/v1/permissions", protected("role.read", api.RoleHandler.ListPermissions))
	mux.Handle("GET /api/v1/roles", protected("role.read", api.RoleHandler.ListRoles))
	mux.Handle("GET /api/v1/access-governance", protected("role.read", api.RoleHandler.AccessGovernance))
	mux.Handle("GET /api/v1/roles/compare", protected("role.read", api.RoleHandler.CompareRoles))
	mux.Handle("GET /api/v1/roles/{id}/impact", protected("role.read", api.RoleHandler.GetRoleImpact))
	mux.Handle("GET /api/v1/roles/{id}", protected("role.read", api.RoleHandler.GetRole))
	mux.Handle("PATCH /api/v1/roles/{id}", protected("role.update", api.RoleHandler.UpdateRole))
	mux.Handle("POST /api/v1/roles", protected("role.create", api.RoleHandler.CreateRole))
	mux.Handle("POST /api/v1/roles/{id}/clone", protected("role.create", api.RoleHandler.CloneRole))
	mux.Handle("DELETE /api/v1/roles/{id}", protected("role.delete", api.RoleHandler.DeleteRole))
	mux.Handle("PUT /api/v1/roles/{id}/permissions", protected("permission.assign", api.RoleHandler.SetRolePermissions))
	mux.Handle("GET /api/v1/users", protected("user.read.all", api.RoleHandler.ListUsers))
	mux.Handle("GET /api/v1/users/{id}/access", protected("user.read.all", api.RoleHandler.GetUserAccess))
	mux.Handle("GET /api/v1/users/{id}", protected("user.read.all", api.RoleHandler.GetUser))
	mux.Handle("POST /api/v1/users", protected("user.create", api.InviteHandler.CreateUser))
	mux.Handle("POST /api/v1/users/invite", protected("user.create", api.InviteHandler.Invite))
	mux.Handle("PATCH /api/v1/users/memberships/{id}", protected("user.update.all", api.RoleHandler.UpdateUserMembership))

	// --- dynamic navigation ---
	mux.Handle("GET /api/v1/navigation", middleware.Authenticated(api.JWTSecret, api.NavigationHandler.List))

	// --- dashboard and personal approval queue ---
	mux.Handle("GET /api/v1/dashboard", protected("report.read.own", api.DashboardHandler.Dashboard))
	mux.Handle("GET /api/v1/dashboard/summary", protected("report.read.own", api.DashboardHandler.Summary))
	mux.Handle("GET /api/v1/approvals/pending", protectedAny([]string{"approval.approve.assigned", "approval.reject.assigned"}, api.GatepassHandler.MyPendingApprovals))
}
