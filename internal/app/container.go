package app

import (
	"database/sql"

	"gatepass/internal/approvals"
	"gatepass/internal/audit"
	"gatepass/internal/auth"
	"gatepass/internal/config"
	"gatepass/internal/dashboard"
	"gatepass/internal/departments"
	"gatepass/internal/employees"
	"gatepass/internal/gatepasses"
	"gatepass/internal/invite"
	"gatepass/internal/middleware"
	"gatepass/internal/platform"
	"gatepass/internal/roles"
	"gatepass/internal/settings"
	"gatepass/internal/tenants"
	"gatepass/internal/users"
	"gatepass/internal/visitors"
	"gatepass/internal/visits"
)

// Container owns the process-wide dependencies used by the HTTP application.
// It is intentionally explicit: dependencies are constructed once at startup
// and passed to modules instead of being created inside handlers.
type Container struct {
	Config    *config.Config
	DB        *sql.DB
	JWTSecret []byte

	TenantRepo *tenants.Repository
	RoleRepo   *roles.Repository

	LoginLimiter   *middleware.RateLimiter
	RefreshLimiter *middleware.RateLimiter

	AuthHandler             *auth.Handler
	VisitorHandler          *visitors.Handler
	VisitorSettingsHandler  *settings.VisitorSettingsHandler
	GatepassSettingsHandler *settings.GatepassSettingsHandler
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
}

// NewContainer constructs all application dependencies. This is the single
// composition point for the HTTP process; domain packages remain unaware of
// how the application is wired together.
func NewContainer(cfg *config.Config, db *sql.DB) *Container {
	jwtSecret := []byte(cfg.JWTSecret)

	tenantRepo := tenants.NewRepository(db)
	userRepo := users.NewRepository(db)
	roleRepo := roles.NewRepository(db)
	refreshRepo := auth.NewRefreshTokenRepository(db)
	settingsRepo := settings.NewRepository(db)
	auditRepo := audit.NewRepository(db)

	idTypeRepo := visitors.NewIDTypeRepository(db)
	companyRepo := visitors.NewCompanyRepository(db)
	visitorRepo := visitors.NewRepository(db)
	visitTypeRepo := visits.NewVisitTypeRepository(db)
	deptRepo := departments.NewRepository(db)
	visitRepo := visits.NewRepository(db)
	workflowRepo := approvals.NewRepository(db)
	gpTypeRepo := gatepasses.NewTypeRepository(db)
	gpItemRepo := gatepasses.NewItemRepository(db)
	gpRepo := gatepasses.NewRepository(db, gpItemRepo)
	employeeRepo := employees.NewRepository(db)

	authSvc := auth.NewService(
		userRepo, roleRepo, refreshRepo,
		jwtSecret, cfg.BcryptCost,
		cfg.AccessTokenTTL, cfg.RefreshTokenTTL,
	)
	visitorSvc := visitors.NewService(visitorRepo, idTypeRepo, companyRepo, settingsRepo, auditRepo)
	visitSvc := visits.NewService(visitRepo, visitorRepo, visitTypeRepo, deptRepo, auditRepo)
	gpSvc := gatepasses.NewService(gpRepo, gpTypeRepo, deptRepo, visitorRepo, visitRepo, workflowRepo, roleRepo, settingsRepo, auditRepo, userRepo)
	inviteSvc := invite.NewService(userRepo, roleRepo, cfg.BcryptCost)
	employeeSvc := employees.NewService(employeeRepo, userRepo, roleRepo)
	bootstrapSvc := platform.NewService(tenantRepo, userRepo, roleRepo, cfg.BcryptCost)

	return &Container{
		Config:    cfg,
		DB:        db,
		JWTSecret: jwtSecret,

		TenantRepo: tenantRepo,
		RoleRepo:   roleRepo,

		LoginLimiter:   middleware.NewRateLimiter(10, time.Minute),
		RefreshLimiter: middleware.NewRateLimiter(30, time.Minute),

		AuthHandler:             auth.NewHandler(authSvc),
		VisitorHandler:          visitors.NewHandler(visitorSvc, idTypeRepo, companyRepo),
		VisitorSettingsHandler:  settings.NewVisitorSettingsHandler(settingsRepo),
		GatepassSettingsHandler: settings.NewGatepassSettingsHandler(settingsRepo),
		VisitTypeHandler:        visits.NewVisitTypeHandler(visitTypeRepo),
		DepartmentHandler:       departments.NewHandler(deptRepo),
		VisitHandler:            visits.NewHandler(visitSvc),
		WorkflowHandler:         approvals.NewHandler(workflowRepo),
		GatepassHandler:         gatepasses.NewHandler(gpSvc, gpTypeRepo),
		EmployeeHandler:         employees.NewHandler(employeeSvc),
		RoleHandler:             roles.NewHandler(roleRepo),
		InviteHandler:           invite.NewHandler(inviteSvc),
		BootstrapHandler:        platform.NewHandler(bootstrapSvc, cfg.PlatformBootstrapToken),
		DashboardHandler:        dashboard.NewHandler(dashboard.NewRepository(db)),
	}
}
