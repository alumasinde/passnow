package app

import (
	"net/http"

	"gatepass/internal/middleware"
)

func registerVisitRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("GET /api/v1/visit-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitTypeHandler.List))
	mux.Handle("POST /api/v1/visit-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitTypeHandler.Create))
	mux.Handle("PATCH /api/v1/visit-types/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitTypeHandler.Update))

	mux.Handle("GET /api/v1/departments", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.DepartmentHandler.List))
	mux.Handle("POST /api/v1/departments", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visits", c.DepartmentHandler.Create))
	mux.Handle("PATCH /api/v1/departments/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visits", c.DepartmentHandler.Update))

	mux.Handle("POST /api/v1/visits", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.create", c.VisitHandler.Create))
	mux.Handle("GET /api/v1/visits", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.VisitHandler.List))
	mux.Handle("GET /api/v1/visits/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.VisitHandler.Get))
	mux.Handle("POST /api/v1/visits/{id}/check-in", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.checkin", c.VisitHandler.CheckIn))
	mux.Handle("POST /api/v1/visits/{id}/check-out", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.checkout", c.VisitHandler.CheckOut))
	mux.Handle("POST /api/v1/visits/{id}/cancel", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.cancel", c.VisitHandler.Cancel))
	mux.Handle("GET /api/v1/visits/badge/{token}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visits.view", c.VisitHandler.BadgeLookup))
}
