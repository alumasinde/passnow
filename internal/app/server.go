package app

import (
	"net/http"

	"gatepass/internal/config"
)

// NewServer creates the HTTP server without starting it. Keeping construction
// separate from lifecycle makes the server easy to test and keeps Run small.
func NewServer(c *Container) *http.Server {
	return &http.Server{
		Addr:         c.Config.HTTPAddr,
		Handler:      RegisterRoutes(c),
		ReadTimeout:  c.Config.ReadTimeout,
		WriteTimeout: c.Config.WriteTimeout,
		IdleTimeout:  c.Config.IdleTimeout,
		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}
}

var _ = config.Config{}
