package app

import (
	"net/http"

	"gatepass/internal/middleware"
)

// NewServer creates the HTTP server without starting it. Keeping construction
// separate from lifecycle makes Run small and the server independently testable.
func NewServer(c *Container) *http.Server {
	handler := middleware.RequestID(
		middleware.Recover(
			middleware.SecurityHeaders(RegisterRoutes(c)),
		),
	)

	return &http.Server{
		Addr:           c.Config.HTTPAddr,
		Handler:        handler,
		ReadTimeout:    c.Config.ReadTimeout,
		WriteTimeout:   c.Config.WriteTimeout,
		IdleTimeout:    c.Config.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}
}
