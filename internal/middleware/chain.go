package middleware

import (
	"net/http"

	"gatepass/internal/roles"
)

// Protected wraps a handler with auth + a required permission — the
// combination almost every privileged route needs. One helper here
// instead of the same two-middleware chain typed out at every
// mux.Handle call in main.go.
func Protected(secret []byte, roleRepo *roles.Repository, permission string, h http.HandlerFunc) http.Handler {
	return RequireAuth(secret)(RequirePermission(roleRepo, permission)(h))
}

// Authenticated wraps a handler with auth only, no specific permission
// requirement (e.g. logout — any authenticated user may log themself out).
func Authenticated(secret []byte, h http.HandlerFunc) http.Handler {
	return RequireAuth(secret)(h)
}
