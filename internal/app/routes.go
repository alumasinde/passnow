package app

import (
	"context"
	"net/http"
	"time"

	"gatepass/internal/middleware"
)

// RegisterRoutes builds the HTTP router from the application container.
// Domain route groups live in dedicated files so startup wiring stays small
// while the public /api/v1 contract remains unchanged.
func RegisterRoutes(c *Container) http.Handler {
	rootMux := http.NewServeMux()
	tenantMux := http.NewServeMux()

	rootMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
