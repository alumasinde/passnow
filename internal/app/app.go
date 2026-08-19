package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gatepass/internal/approvals"
	"gatepass/internal/audit"
	"gatepass/internal/auth"
	"gatepass/internal/config"
	"gatepass/internal/dashboard"
	"gatepass/internal/database"
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

// Run loads application configuration, constructs dependencies, starts the
// background worker and HTTP server, and handles graceful shutdown.
//
// Phase 1 intentionally keeps the existing wiring and route behavior intact.
// The purpose of this package is to move application/bootstrap concerns out of
// cmd/api/main.go without changing business logic.
func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	jwtSecret := []byte(cfg.JWTSecret)

	// --- repositories ---
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

	// --- services ---
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

	// --- background operations ---
	// The worker owns scheduling only; all Gatepass state rules remain in the
	// Gatepass/domain/repository layers. It is safe to run in each API instance
	// because notification events are idempotent by event_key.
	gatepassWorker := gatepasses.NewWorker(db, cfg.GatepassWorkerInterval, cfg.ApprovedGatepassTTL, log.Default())
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go gatepassWorker.Run(workerCtx)

	// Authentication is deliberately rate-limited before expensive password
	// verification. This is process-local; production multi-instance deployments
	// should also enforce the same policy at the API gateway/WAF.
	loginLimiter := middleware.NewRateLimiter(10, time.Minute)
	refreshLimiter := middleware.NewRateLimiter(30, time.Minute)

	// --- handlers ---
	authHandler := auth.NewHandler(authSvc)
	visitorHandler := visitors.NewHandler(visitorSvc, idTypeRepo, companyRepo)
	visitorSettingsHandler := settings.NewVisitorSettingsHandler(settingsRepo)
	gatepassSettingsHandler := settings.NewGatepassSettingsHandler(settingsRepo)
	visitTypeHandler := visits.NewVisitTypeHandler(visitTypeRepo)
	departmentHandler := departments.NewHandler(deptRepo)
	visitHandler := visits.NewHandler(visitSvc)
	workflowHandler := approvals.NewHandler(workflowRepo)
	gpHandler := gatepasses.NewHandler(gpSvc, gpTypeRepo)
	employeeHandler := employees.NewHandler(employeeSvc)
	roleHandler := roles.NewHandler(roleRepo)
	inviteHandler := invite.NewHandler(inviteSvc)
	bootstrapHandler := platform.NewHandler(bootstrapSvc, cfg.PlatformBootstrapToken)
	dashboardHandler := dashboard.NewHandler(dashboard.NewRepository(db))

	// tenantMux holds every route that needs a resolved tenant. rootMux holds
	// routes that must work with no tenant context: health checks and the
	// platform bootstrap endpoint.
	tenantMux := http.NewServeMux()
	rootMux := http.NewServeMux()

	rootMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	rootMux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	rootMux.HandleFunc("POST /api/v1/platform/bootstrap-tenant", bootstrapHandler.BootstrapTenant)

	mux := tenantMux

	// --- auth (public within a tenant) ---
	mux.Handle("POST /api/v1/auth/login", loginLimiter.Middleware("login")(http.HandlerFunc(authHandler.Login)))
	mux.Handle("POST /api/v1/auth/refresh", refreshLimiter.Middleware("refresh")(http.HandlerFunc(authHandler.Refresh)))
	mux.Handle("POST /api/v1/auth/logout", middleware.Authenticated(jwtSecret, authHandler.Logout))

	// --- visitors ---
	mux.Handle("POST /api/v1/visitors", middleware.Protected(jwtSecret, roleRepo, "visitors.create", visitorHandler.Create))
	mux.Handle("GET /api/v1/visitors", middleware.Protected(jwtSecret, roleRepo, "visitors.view", visitorHandler.List))
	mux.Handle("GET /api/v1/visitors/{id}", middleware.Protected(jwtSecret, roleRepo, "visitors.view", visitorHandler.Get))
	mux.Handle("PATCH /api/v1/visitors/{id}", middleware.Protected(jwtSecret, roleRepo, "visitors.update", visitorHandler.Update))
	mux.Handle("POST /api/v1/visitors/{id}/blacklist", middleware.Protected(jwtSecret, roleRepo, "visitors.update", visitorHandler.SetBlacklist))

	// --- id types (visitor config) ---
	mux.Handle("GET /api/v1/id-types", middleware.Protected(jwtSecret, roleRepo, "visitors.view", visitorHandler.ListIDTypes))
	mux.Handle("POST /api/v1/id-types", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitorHandler.CreateIDType))
	mux.Handle("PATCH /api/v1/id-types/{id}", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitorHandler.UpdateIDType))

	// --- visitor companies (visitor config) ---
	mux.Handle("GET /api/v1/visitor-companies", middleware.Protected(jwtSecret, roleRepo, "visitors.view", visitorHandler.ListCompanies))
	mux.Handle("POST /api/v1/visitor-companies", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitorHandler.CreateCompany))
	mux.Handle("PATCH /api/v1/visitor-companies/{id}", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitorHandler.UpdateCompany))

	// --- visit types (shared config, used later by Visits module) ---
	mux.Handle("GET /api/v1/visit-types", middleware.Protected(jwtSecret, roleRepo, "visitors.view", visitTypeHandler.List))
	mux.Handle("POST /api/v1/visit-types", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitTypeHandler.Create))
	mux.Handle("PATCH /api/v1/visit-types/{id}", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitTypeHandler.Update))

	// --- Platform Admin: visitor settings (pre-registration toggle etc.) ---
	mux.Handle("GET /api/v1/settings/visitors", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitorSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/visitors", middleware.Protected(jwtSecret, roleRepo, "settings.visitors", visitorSettingsHandler.Update))

	// --- departments (visit config) ---
	mux.Handle("GET /api/v1/departments", middleware.Protected(jwtSecret, roleRepo, "visits.view", departmentHandler.List))
	mux.Handle("POST /api/v1/departments", middleware.Protected(jwtSecret, roleRepo, "settings.visits", departmentHandler.Create))
	mux.Handle("PATCH /api/v1/departments/{id}", middleware.Protected(jwtSecret, roleRepo, "settings.visits", departmentHandler.Update))

	// --- visits ---
	mux.Handle("POST /api/v1/visits", middleware.Protected(jwtSecret, roleRepo, "visits.create", visitHandler.Create))
	mux.Handle("GET /api/v1/visits", middleware.Protected(jwtSecret, roleRepo, "visits.view", visitHandler.List))
	mux.Handle("GET /api/v1/visits/{id}", middleware.Protected(jwtSecret, roleRepo, "visits.view", visitHandler.Get))
	mux.Handle("POST /api/v1/visits/{id}/check-in", middleware.Protected(jwtSecret, roleRepo, "visits.checkin", visitHandler.CheckIn))
	mux.Handle("POST /api/v1/visits/{id}/check-out", middleware.Protected(jwtSecret, roleRepo, "visits.checkout", visitHandler.CheckOut))
	mux.Handle("POST /api/v1/visits/{id}/cancel", middleware.Protected(jwtSecret, roleRepo, "visits.cancel", visitHandler.Cancel))
	mux.Handle("GET /api/v1/visits/badge/{token}", middleware.Protected(jwtSecret, roleRepo, "visits.view", visitHandler.BadgeLookup))

	// --- approval workflow templates ---
	mux.Handle("GET /api/v1/approval-workflows", middleware.Protected(jwtSecret, roleRepo, "settings.approvals", workflowHandler.List))
	mux.Handle("GET /api/v1/approval-workflows/{id}", middleware.Protected(jwtSecret, roleRepo, "settings.approvals", workflowHandler.Get))
	mux.Handle("POST /api/v1/approval-workflows", middleware.Protected(jwtSecret, roleRepo, "settings.approvals", workflowHandler.Create))

	// --- gatepass types (config) ---
	mux.Handle("GET /api/v1/gatepass-types", middleware.Protected(jwtSecret, roleRepo, "gatepasses.view", gpHandler.ListTypes))
	mux.Handle("POST /api/v1/gatepass-types", middleware.Protected(jwtSecret, roleRepo, "settings.gatepass", gpHandler.CreateType))
	mux.Handle("PATCH /api/v1/gatepass-types/{id}", middleware.Protected(jwtSecret, roleRepo, "settings.gatepass", gpHandler.UpdateType))

	// --- gatepasses ---
	mux.Handle("POST /api/v1/gatepasses", middleware.Protected(jwtSecret, roleRepo, "gatepasses.create", gpHandler.Create))
	mux.Handle("GET /api/v1/gatepasses", middleware.Protected(jwtSecret, roleRepo, "gatepasses.view", gpHandler.List))
	mux.Handle("GET /api/v1/gatepasses/{id}", middleware.Protected(jwtSecret, roleRepo, "gatepasses.view", gpHandler.Get))
	mux.Handle("POST /api/v1/gatepasses/{id}/cancel", middleware.Protected(jwtSecret, roleRepo, "gatepasses.cancel", gpHandler.Cancel))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/approve", middleware.Protected(jwtSecret, roleRepo, "gatepasses.approve", gpHandler.Approve))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/reject", middleware.Protected(jwtSecret, roleRepo, "gatepasses.reject", gpHandler.Reject))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-out", middleware.Protected(jwtSecret, roleRepo, "gatepasses.issue", gpHandler.CheckOut))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-in", middleware.Protected(jwtSecret, roleRepo, "gatepasses.verify", gpHandler.CheckIn))
	mux.Handle("GET /api/v1/gatepasses/{id}/movements", middleware.Protected(jwtSecret, roleRepo, "gatepasses.view", gpHandler.Movements))
	mux.Handle("GET /api/v1/gatepasses/{id}/qr.png", middleware.Protected(jwtSecret, roleRepo, "gatepasses.view", gpHandler.QRImage))
	mux.Handle("GET /api/v1/gatepasses/qr/{token}", middleware.Protected(jwtSecret, roleRepo, "gatepasses.view", gpHandler.QRLookup))

	// --- Platform Admin: gatepass numbering settings ---
	mux.Handle("GET /api/v1/settings/gatepass", middleware.Protected(jwtSecret, roleRepo, "settings.gatepass", gatepassSettingsHandler.Get))
	mux.Handle("PUT /api/v1/settings/gatepass", middleware.Protected(jwtSecret, roleRepo, "settings.gatepass", gatepassSettingsHandler.Update))

	// --- employees ---
	mux.Handle("GET /api/v1/employees", middleware.Protected(jwtSecret, roleRepo, "employees.view", employeeHandler.List))
	mux.Handle("GET /api/v1/employees/{id}", middleware.Protected(jwtSecret, roleRepo, "employees.view", employeeHandler.Get))
	mux.Handle("POST /api/v1/employees", middleware.Protected(jwtSecret, roleRepo, "employees.create", employeeHandler.Create))
	mux.Handle("PATCH /api/v1/employees/{id}", middleware.Protected(jwtSecret, roleRepo, "employees.update", employeeHandler.Update))

	// --- roles & permissions (Settings) ---
	mux.Handle("GET /api/v1/permissions", middleware.Protected(jwtSecret, roleRepo, "settings.roles", roleHandler.ListPermissions))
	mux.Handle("GET /api/v1/roles", middleware.Protected(jwtSecret, roleRepo, "settings.roles", roleHandler.ListRoles))
	mux.Handle("POST /api/v1/roles", middleware.Protected(jwtSecret, roleRepo, "settings.roles", roleHandler.CreateRole))
	mux.Handle("PUT /api/v1/roles/{id}/permissions", middleware.Protected(jwtSecret, roleRepo, "settings.permissions", roleHandler.SetRolePermissions))

	// --- users (tenant memberships, Settings) ---
	mux.Handle("GET /api/v1/users", middleware.Protected(jwtSecret, roleRepo, "settings.users", roleHandler.ListUsers))
	mux.Handle("POST /api/v1/users/invite", middleware.Protected(jwtSecret, roleRepo, "settings.users", inviteHandler.Invite))
	mux.Handle("PATCH /api/v1/users/memberships/{id}", middleware.Protected(jwtSecret, roleRepo, "settings.users", roleHandler.UpdateUserMembership))

	// --- dashboard ---
	mux.Handle("GET /api/v1/dashboard/summary", middleware.Protected(jwtSecret, roleRepo, "dashboard.view", dashboardHandler.Summary))

	// --- approvals work queue (personal, cuts across gatepasses) ---
	mux.Handle("GET /api/v1/approvals/pending", middleware.Protected(jwtSecret, roleRepo, "gatepasses.approve", gpHandler.MyPendingApprovals))

	// Every tenant-scoped request passes through tenant resolution first so
	// reqctx.TenantFromContext is populated before anything else runs.
	rootMux.Handle("/", middleware.ResolveTenant(tenantRepo, cfg.BaseDomain)(tenantMux))
	handler := http.Handler(rootMux)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		log.Printf("gatepass api listening on %s (env=%s)", cfg.HTTPAddr, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	workerCancel()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("shutdown complete")
}
