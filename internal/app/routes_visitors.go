package app

import (
	"net/http"

	"gatepass/internal/middleware"
)

func registerVisitorRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("POST /api/v1/visitors", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.create", c.VisitorHandler.Create))
	mux.Handle("GET /api/v1/visitors", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.List))
	mux.Handle("GET /api/v1/visitors/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.Get))
	mux.Handle("PATCH /api/v1/visitors/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.update", c.VisitorHandler.Update))
	mux.Handle("POST /api/v1/visitors/{id}/blacklist", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.update", c.VisitorHandler.SetBlacklist))

	mux.Handle("GET /api/v1/id-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.ListIDTypes))
	mux.Handle("POST /api/v1/id-types", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.CreateIDType))
	mux.Handle("PATCH /api/v1/id-types/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.UpdateIDType))

	mux.Handle("GET /api/v1/visitor-companies", middleware.Protected(c.JWTSecret, c.RoleRepo, "visitors.view", c.VisitorHandler.ListCompanies))
	mux.Handle("POST /api/v1/visitor-companies", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.CreateCompany))
	mux.Handle("PATCH /api/v1/visitor-companies/{id}", middleware.Protected(c.JWTSecret, c.RoleRepo, "settings.visitors", c.VisitorHandler.UpdateCompany))
}
