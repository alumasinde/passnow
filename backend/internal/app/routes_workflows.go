package app

import (
	"net/http"

	"gatepass/internal/middleware"
)

func registerApprovalRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/approval-workflows", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.approvals", c.WorkflowHandler.List))
	mux.Handle("GET /api/v1/approval-workflows/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.approvals", c.WorkflowHandler.Get))
	mux.Handle("POST /api/v1/approval-workflows", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.approvals", c.WorkflowHandler.Create))
}
