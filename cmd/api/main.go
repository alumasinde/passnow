package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	_ = api // tenant APIs are constructed against the resolved tenant database per request.
	srv, workerCancel := newServer(cfg, db, tenantRepo, bootstrapHandler, platformAdminHandler, platformAdminRepo, []byte(cfg.JWTSecret))
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

func buildApplication(db *sql.DB, cfg *config.Config) (*tenants.Repository, *routes.API, *platform.Handler, *platform.AdminHandler, *platform.AdminRepository) {
	jwtSecret := []byte(cfg.JWTSecret)

	// Repositories.
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

	// Services.
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

	// Handlers and API route dependencies.
	api := routes.NewAPI(jwtSecret, roleRepo)
	api.AuthHandler = auth.NewHandler(authSvc)
	api.VisitorHandler = visitors.NewHandler(visitorSvc, idTypeRepo, companyRepo)
	api.VisitorSettingsHandler = settings.NewVisitorSettingsHandler(settingsRepo)
	api.GatepassSettingsHandler = settings.NewGatepassSettingsHandler(settingsRepo)
	api.VisitTypeHandler = visits.NewVisitTypeHandler(visitTypeRepo)
	api.DepartmentHandler = departments.NewHandler(deptRepo)
	api.VisitHandler = visits.NewHandler(visitSvc)
	api.WorkflowHandler = approvals.NewHandler(workflowRepo)
	api.GatepassHandler = gatepasses.NewHandler(gpSvc, gpTypeRepo)
	api.EmployeeHandler = employees.NewHandler(employeeSvc)
	api.RoleHandler = roles.NewHandler(roleRepo)
	api.InviteHandler = invite.NewHandler(inviteSvc)
	api.DashboardHandler = dashboard.NewHandler(dashboard.NewRepository(db))

	return tenantRepo, api, platform.NewHandler(bootstrapSvc, cfg.PlatformBootstrapToken), platformAdminHandler, platformAdminRepo
}

func newServer(cfg *config.Config, db *sql.DB, tenantRepo *tenants.Repository, bootstrapHandler *platform.Handler, platformAdminHandler *platform.AdminHandler, platformAdminRepo *platform.AdminRepository, jwtSecret []byte) (*http.Server, context.CancelFunc) {
	rootMux := http.NewServeMux()

	tenantMux := &tenantAPIHandler{platformDB: db, cfg: cfg}
	routes.RegisterWeb(rootMux, db, bootstrapHandler, platformAdminHandler, platformAdminRepo, tenantRepo, jwtSecret)
	handler := routes.BuildHandler(cfg, tenantRepo, rootMux, tenantMux)

	// Gatepass operational data lives in isolated tenant databases. A worker
	// connected to the platform database would query tenant tables that do not
	// exist there. Tenant workers are therefore not started from the platform
	// connection.
	workerCancel := func() {}
	log.Printf("gatepass worker: tenant-scoped worker not started on platform database")

	return &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}, workerCancel
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
	platformDB *sql.DB
	cfg *config.Config
}

func (h *tenantAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant context missing", http.StatusInternalServerError)
		return
	}
	if h.cfg.TenantDBEncryptionKey == "" {
		http.Error(w, "tenant database configuration is missing", http.StatusServiceUnavailable)
		return
	}
	cipher, err := tenantdb.NewCipher(h.cfg.TenantDBEncryptionKey)
	if err != nil {
		http.Error(w, "tenant database configuration is invalid", http.StatusInternalServerError)
		return
	}
	conn, err := tenantdb.NewRepository(h.platformDB).Get(r.Context(), tenant.ID)
	if err != nil {
		http.Error(w, "tenant database is not available", http.StatusServiceUnavailable)
		return
	}
	password, err := cipher.Decrypt(conn.EncryptedPassword)
	if err != nil {
		http.Error(w, "tenant database credentials could not be decrypted", http.StatusInternalServerError)
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC", conn.Username, password, conn.Host, conn.Port, conn.DatabaseName)
	tenantDB, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, "tenant database connection could not be opened", http.StatusServiceUnavailable)
		return
	}
	defer tenantDB.Close()
	if err := tenantDB.PingContext(r.Context()); err != nil {
		http.Error(w, "tenant database could not be reached", http.StatusServiceUnavailable)
		return
	}
	api := buildTenantAPI(tenantDB, h.cfg)
	mux := http.NewServeMux()
	routes.RegisterAPI(mux, api)
	mux.ServeHTTP(w, r)
}
