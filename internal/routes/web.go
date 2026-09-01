package routes

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"gatepass/internal/config"
	"gatepass/internal/middleware"
	"gatepass/internal/platform"
	"gatepass/internal/tenants"
)

// RegisterWeb registers routes that are intentionally outside tenant resolution:
// health checks and first-tenant bootstrap.
func RegisterWeb(rootMux *http.ServeMux, db *sql.DB, bootstrapHandler *platform.Handler, platformAdminHandler *platform.AdminHandler, platformAdminRepo *platform.AdminRepository, tenantRepo *tenants.Repository, jwtSecret []byte) {
	rootMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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

	// Platform routes are intentionally outside tenant resolution.
	rootMux.HandleFunc("POST /api/v1/platform/auth/login", platformAdminHandler.Login)
	rootMux.Handle("GET /api/v1/platform/me", middleware.PlatformAdmin(jwtSecret, platformAdminRepo, http.HandlerFunc(platformAdminHandler.Me)))
	tenantHandler := platform.NewTenantHandler(tenantRepo)
	rootMux.Handle("GET /api/v1/platform/tenants", middleware.PlatformAdmin(jwtSecret, platformAdminRepo, http.HandlerFunc(tenantHandler.List)))
	rootMux.Handle("GET /api/v1/platform/tenants/{id}", middleware.PlatformAdmin(jwtSecret, platformAdminRepo, http.HandlerFunc(tenantHandler.Get)))
	rootMux.Handle("PATCH /api/v1/platform/tenants/{id}/status", middleware.PlatformAdmin(jwtSecret, platformAdminRepo, http.HandlerFunc(tenantHandler.UpdateStatus)))
	rootMux.HandleFunc("POST /api/v1/platform/bootstrap-tenant", bootstrapHandler.BootstrapTenant)
}

// BuildHandler layers public routes over the tenant-scoped API. Health checks
// and bootstrap remain available even when no tenant can be resolved.
func BuildHandler(cfg *config.Config, tenantRepo *tenants.Repository, rootMux, tenantMux *http.ServeMux) http.Handler {
	rootMux.Handle("/", middleware.ResolveTenant(tenantRepo, cfg.BaseDomain)(tenantMux))
	return rootMux
}
