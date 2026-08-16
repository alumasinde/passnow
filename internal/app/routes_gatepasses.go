package app

import (
	"net/http"

	"gatepass/internal/middleware"
)

func registerGatepassRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/gatepass-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.ListTypes))
	mux.Handle("POST /api/v1/gatepass-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.gatepass", c.GatepassHandler.CreateType))
	mux.Handle("PATCH /api/v1/gatepass-types/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.gatepass", c.GatepassHandler.UpdateType))

	mux.Handle("POST /api/v1/gatepasses", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.create", c.GatepassHandler.Create))
	mux.Handle("GET /api/v1/gatepasses", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.List))
	mux.Handle("GET /api/v1/gatepasses/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.Get))
	mux.Handle("POST /api/v1/gatepasses/{id}/cancel", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.cancel", c.GatepassHandler.Cancel))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/approve", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.approve", c.GatepassHandler.Approve))
	mux.Handle("POST /api/v1/gatepasses/{id}/approvals/{stepId}/reject", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.reject", c.GatepassHandler.Reject))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-out", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.issue", c.GatepassHandler.CheckOut))
	mux.Handle("POST /api/v1/gatepasses/{id}/check-in", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.verify", c.GatepassHandler.CheckIn))
	mux.Handle("GET /api/v1/gatepasses/{id}/movements", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.Movements))
	mux.Handle("GET /api/v1/gatepasses/{id}/qr.png", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.QRImage))
	mux.Handle("GET /api/v1/gatepasses/qr/{token}", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.view", c.GatepassHandler.QRLookup))

	mux.Handle("GET /api/v1/approvals/pending", middleware.Protected(c.JWTSecret, c.RoleRepo, "gatepasses.approve", c.GatepassHandler.MyPendingApprovals))
}
