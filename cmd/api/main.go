package main

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
	"gatepass/internal/platform"
	"gatepass/internal/roles"
	"gatepass/internal/routes"
	"gatepass/internal/settings"
	"gatepass/internal/tenants"
	"gatepass/internal/users"
	"gatepass/internal/visitors"
	"gatepass/internal/visits"
)

func main() {
	cfg := mustLoadConfig()
	db := mustConnectDatabase(cfg)
	defer db.Close()

	tenantRepo, api, bootstrapHandler := buildApplication(db, cfg)
	srv, workerCancel := newServer(cfg, db, tenantRepo, api, bootstrapHandler)
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

func buildApplication(db *sql.DB, cfg *config.Config) (*tenants.Repository, *routes.API, *platform.Handler) {
	jwtSecret := []byte(cfg.JWTSecret)

	// Repositories.
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

	// Services.
	authSvc := auth.NewService(userRepo, roleRepo, refreshRepo, jwtSecret, cfg.BcryptCost, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	visitorSvc := visitors.NewService(visitorRepo, idTypeRepo, companyRepo, settingsRepo, auditRepo)
	visitSvc := visits.NewService(visitRepo, visitorRepo, visitTypeRepo, deptRepo, auditRepo)
	gpSvc := gatepasses.NewService(gpRepo, gpTypeRepo, deptRepo, visitorRepo, visitRepo, workflowRepo, roleRepo, settingsRepo, auditRepo, userRepo)
	inviteSvc := invite.NewService(userRepo, roleRepo, cfg.BcryptCost)
	employeeSvc := employees.NewService(employeeRepo, userRepo, roleRepo)
	bootstrapSvc := platform.NewService(tenantRepo, userRepo, roleRepo, cfg.BcryptCost)

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

	return tenantRepo, api, platform.NewHandler(bootstrapSvc, cfg.PlatformBootstrapToken)
}

func newServer(cfg *config.Config, db *sql.DB, tenantRepo *tenants.Repository, api *routes.API, bootstrapHandler *platform.Handler) (*http.Server, context.CancelFunc) {
	tenantMux := http.NewServeMux()
	rootMux := http.NewServeMux()

	routes.RegisterAPI(tenantMux, api)
	routes.RegisterWeb(rootMux, db, bootstrapHandler)
	handler := routes.BuildHandler(cfg, tenantRepo, rootMux, tenantMux)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	worker := gatepasses.NewWorker(db, cfg.GatepassWorkerInterval, cfg.ApprovedGatepassTTL, log.Default())
	go worker.Run(workerCtx)

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
