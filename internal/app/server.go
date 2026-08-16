package app

import "net/http"

// NewServer creates the HTTP server without starting it. Keeping construction
// separate from lifecycle makes Run small and the server independently testable.
func NewServer(c *Container) *http.Server {
	return &http.Server{
		Addr:         c.Config.HTTPAddr,
		Handler:      RegisterRoutes(c),
		ReadTimeout:  c.Config.ReadTimeout,
		WriteTimeout: c.Config.WriteTimeout,
		IdleTimeout:  c.Config.IdleTimeout,
	}
}
