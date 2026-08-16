package app

import (
	"net/http"

	"gatepass/internal/middleware"
)

func registerAuthRoutes(mux *http.ServeMux, c *Container) {
	mux.Handle("POST /api/v1/auth/login", c.LoginLimiter.Middleware("login")(http.HandlerFunc(c.AuthHandler.Login)))
	mux.Handle("POST /api/v1/auth/refresh", c.RefreshLimiter.Middleware("refresh")(http.HandlerFunc(c.AuthHandler.Refresh)))
	mux.Handle("POST /api/v1/auth/logout", middleware.Authenticated(c.JWTSecret, c.AuthHandler.Logout))
}
