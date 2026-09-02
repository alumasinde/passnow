package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"sync"

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
	"gatepass/internal/media"
	"gatepass/internal/navigation"
	"gatepass/internal/platform"
	"gatepass/internal/roles"
	"gatepass/internal/routes"
	"gatepass/internal/reqctx"
	"gatepass/internal/settings"
	"gatepass/internal/tenants"
	"gatepass/internal/tenantdb"
	"gatepass/internal/users"
	"gatepass/internal/visitors"
	"gatepass/internal/visits"
)

func main() {
	cfg := mustLoadConfig()
	db := mustConnectDatabase(cfg)
	defer db.Close()

	tenantRepo, api, bootstrapHandler, platformAdminHandler, platformAdminRepo := buildApplication(db, cfg)
	_ = api
	tenantManager := mustTenantDBManager(db, cfg)
	defer tenantManager.Close()
	srv, workerCancel := newServer(cfg, db, tenantRepo, tenantManager, bootstrapHandler, platformAdminHandler, platformAdminRepo, tenantManager, []byte(cfg.JWTSecret))
	defer workerCancel()

	go serve(srv, cfg)
	waitForShutdown(srv, cfg)
}

func mustLoadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	return cfg
}

func mustConnectDatabase(cfg *config.Config) *sql.DB {
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	return db
}

func mustTenantDBManager(db *sql.DB, cfg *config.Config) *tenantdb.Manager {
	if cfg.TenantDBEncryptionKey == "" {
		log.Fatal("tenant database encryption key is required")
	}
	cipher, err := tenantdb.NewCipher(cfg.TenantDBEncryptionKey)
	if err != nil {
		log.Fatalf("tenant database encryption: %v", err)
	}
	return tenantdb.NewManager(
		tenantdb.NewRepository(db),
		cipher,
		cfg.TenantDBMaxOpenConns,
		cfg.TenantDBMaxIdleConns,
		cfg.TenantDBConnMaxLifetime,
	)
}

func buildApplication(db *sql.DB, cfg *config.Config) (*tenants.Repository, *routes.API, *platform.Handler, *platform.AdminHandler, *platform.AdminRepository) {
	jwtSecret := []byte(cfg.JWTSecret)

	tenantRepo := tenants.NewRepository(db)
	if err := tenantRepo.ReconcilePlatformDomains(context.Background(), cfg.BaseDomain); err != nil {
		log.Fatalf("tenant domains: %v", err)
	}
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

	authSvc := auth.NewService(userRepo, roleRepo, refreshRepo, jwtSecret, cfg.BcryptCost, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	visitorSvc := visitors.NewService(visitorRepo, idTypeRepo, companyRepo, settingsRepo, auditRepo)
	visitSvc := visits.NewService(visitRepo, visitorRepo, visitTypeRepo, deptRepo, auditRepo)
	gpSvc := gatepasses.NewService(gpRepo, gpTypeRepo, deptRepo, visitorRepo, visitRepo, workflowRepo, roleRepo, settingsRepo, auditRepo, userRepo)
	inviteSvc := invite.NewService(userRepo, roleRepo, cfg.BcryptCost)
	employeeSvc := employees.NewService(employeeRepo, userRepo, roleRepo)
	bootstrapSvc := platform.NewService(tenantRepo, userRepo, roleRepo, cfg.BcryptCost).WithBaseDomain(cfg.BaseDomain)
	if cfg.TenantDBEncryptionKey != "" {
		if cipher, err := tenantdb.NewCipher(cfg.TenantDBEncryptionKey); err == nil {
			bootstrapSvc.WithTenantDatabase(
				tenantdb.NewRepository(db),
				cipher,
				tenantdb.NewInstaller("migrations/tenant"),
				tenantdb.NewProvisioner(cfg.DBProvisionHost, cfg.DBProvisionPort, cfg.DBProvisionUser, cfg.DBProvisionPassword),
			)
		} else {
			log.Fatalf("tenant database encryption: %v", err)
		}
	}
	platformAdminRepo := platform.NewAdminRepository(db)
	platformAdminSvc := platform.NewAdminService(platformAdminRepo, userRepo, jwtSecret, cfg.AccessTokenTTL)
	platformAdminHandler := platform.NewAdminHandler(platformAdminSvc)

	api := routes.NewAPI(jwtSecret, roleRepo)
	api.AuthHandler = auth.NewHandler(authSvc)
	api.VisitorHandler = visitors.NewHandler(visitorSvc, idTypeRepo, companyRepo)
	api.VisitorSettingsHandler = settings.NewVisitorSettingsHandler(settingsRepo)
	api.GatepassSettingsHandler = settings.NewGatepassSettingsHandler(settingsRepo)
	api.ThemeHandler = settings.NewThemeHandler(settingsRepo)
	api.MediaHandler = media.NewHandler(media.NewRepository(db), cfg.MediaStoragePath, cfg.MediaPublicBaseURL, cfg.MediaMaxUploadBytes)
	api.VisitTypeHandler = visits.NewVisitTypeHandler(visitTypeRepo)
	api.DepartmentHandler = departments.NewHandler(deptRepo)
	api.VisitHandler = visits.NewHandler(visitSvc)
	api.WorkflowHandler = approvals.NewHandler(workflowRepo)
	api.GatepassHandler = gatepasses.NewHandler(gpSvc, gpTypeRepo)
	api.EmployeeHandler = employees.NewHandler(employeeSvc)
	api.RoleHandler = roles.NewHandler(roleRepo)
	api.InviteHandler = invite.NewHandler(inviteSvc)
	dashboardRepo := dashboard.NewRepository(db)
	api.DashboardHandler = dashboard.NewHandler(dashboardRepo, dashboard.NewService(dashboardRepo, roleRepo))
	api.NavigationHandler = navigation.NewHandler(navigation.NewService(roleRepo))

	return tenantRepo, api, platform.NewHandler(bootstrapSvc, cfg.PlatformBootstrapToken), platformAdminHandler, platformAdminRepo
}

func buildTenantAPI(db *sql.DB, cfg *config.Config) *routes.API {
	jwtSecret := []byte(cfg.JWTSecret)
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

	authSvc := auth.NewService(userRepo, roleRepo, refreshRepo, jwtSecret, cfg.BcryptCost, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	visitorSvc := visitors.NewService(visitorRepo, idTypeRepo, companyRepo, settingsRepo, auditRepo)
	visitSvc := visits.NewService(visitRepo, visitorRepo, visitTypeRepo, deptRepo, auditRepo)
	gpSvc := gatepasses.NewService(gpRepo, gpTypeRepo, deptRepo, visitorRepo, visitRepo, workflowRepo, roleRepo, settingsRepo, auditRepo, userRepo)
	inviteSvc := invite.NewService(userRepo, roleRepo, cfg.BcryptCost)
	employeeSvc := employees.NewService(employeeRepo, userRepo, roleRepo)

	api := routes.NewAPI(jwtSecret, roleRepo)
	api.AuthHandler = auth.NewHandler(authSvc)
	api.VisitorHandler = visitors.NewHandler(visitorSvc, idTypeRepo, companyRepo)
	api.VisitorSettingsHandler = settings.NewVisitorSettingsHandler(settingsRepo)
	api.GatepassSettingsHandler = settings.NewGatepassSettingsHandler(settingsRepo)
	api.ThemeHandler = settings.NewThemeHandler(settingsRepo)
	api.MediaHandler = media.NewHandler(media.NewRepository(db), cfg.MediaStoragePath, cfg.MediaPublicBaseURL, cfg.MediaMaxUploadBytes)
	api.VisitTypeHandler = visits.NewVisitTypeHandler(visitTypeRepo)
	api.DepartmentHandler = departments.NewHandler(deptRepo)
	api.VisitHandler = visits.NewHandler(visitSvc)
	api.WorkflowHandler = approvals.NewHandler(workflowRepo)
	api.GatepassHandler = gatepasses.NewHandler(gpSvc, gpTypeRepo)
	api.EmployeeHandler = employees.NewHandler(employeeSvc)
	api.RoleHandler = roles.NewHandler(roleRepo)
	api.InviteHandler = invite.NewHandler(inviteSvc)
	dashboardRepo := dashboard.NewRepository(db)
	api.DashboardHandler = dashboard.NewHandler(dashboardRepo, dashboard.NewService(dashboardRepo, roleRepo))
	api.NavigationHandler = navigation.NewHandler(navigation.NewService(roleRepo))
	return api
}

func newServer(cfg *config.Config, db *sql.DB, tenantRepo *tenants.Repository, tenantManager *tenantdb.Manager, bootstrapHandler *platform.Handler, platformAdminHandler *platform.AdminHandler, platformAdminRepo *platform.AdminRepository, tenantManager *tenantdb.Manager, jwtSecret []byte) (*http.Server, context.CancelFunc) {
	rootMux := http.NewServeMux()
	tenantMux := newTenantAPIHandler(tenantManager, cfg)
	routes.RegisterWeb(rootMux, db, bootstrapHandler, platformAdminHandler, platformAdminRepo, tenantRepo, tenantManager, jwtSecret)
	handler := routes.BuildHandler(cfg, tenantRepo, rootMux, tenantMux)

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	tenantWorker := gatepasses.NewTenantWorker(tenantRepo, tenantManager, cfg.GatepassWorkerInterval, cfg.ApprovedGatepassTTL, 100, log.Default())
	go tenantWorker.Run(workerCtx)
	log.Printf("gatepass worker: tenant-scoped worker started")

	return &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}, cancelWorkers
}

func serve(srv *http.Server, cfg *config.Config) {
	log.Printf("passnow api listening on %s (env=%s)", cfg.HTTPAddr, cfg.Env)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}

func waitForShutdown(srv *http.Server, cfg *config.Config) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("shutdown complete")
}

type tenantAPIHandler struct {
	manager  *tenantdb.Manager
	cfg      *config.Config
	mu       sync.Mutex
	handlers map[int64]http.Handler
}

func newTenantAPIHandler(manager *tenantdb.Manager, cfg *config.Config) *tenantAPIHandler {
	return &tenantAPIHandler{manager: manager, cfg: cfg, handlers: make(map[int64]http.Handler)}
}

func (h *tenantAPIHandler) handler(ctx context.Context, tenantID int64) (http.Handler, error) {
	h.mu.Lock()
	if handler := h.handlers[tenantID]; handler != nil {
		h.mu.Unlock()
		return handler, nil
	}
	h.mu.Unlock()

	db, err := h.manager.DB(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	api := buildTenantAPI(db, h.cfg)
	mux := http.NewServeMux()
	routes.RegisterAPI(mux, api)

	h.mu.Lock()
	if existing := h.handlers[tenantID]; existing != nil {
		h.mu.Unlock()
		return existing, nil
	}
	h.handlers[tenantID] = mux
	h.mu.Unlock()
	return mux, nil
}

func (h *tenantAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant context missing", http.StatusInternalServerError)
		return
	}
	handler, err := h.handler(r.Context(), tenant.ID)
	if err != nil {
		http.Error(w, "tenant database is not available", http.StatusServiceUnavailable)
		return
	}
	handler.ServeHTTP(w, r)
}
