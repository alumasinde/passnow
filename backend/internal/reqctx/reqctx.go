// Package reqctx holds per-request values (resolved tenant, verified
// claims) in context.Context. It exists as a standalone package — with no
// dependency on auth or middleware's HTTP wiring — specifically so both
// the auth package (handlers) and the middleware package (which populates
// these values) can depend on it without an import cycle.
package reqctx

import (
	"context"

	"gatepass/internal/tenants"
)

type ctxKey string

const (
	tenantKey ctxKey = "reqctx.tenant"
	claimsKey ctxKey = "reqctx.claims"
)

// Claims mirrors the fields callers actually need from a verified access
// token. Kept separate from auth.Claims so this package never has to
// import auth.
type Claims struct {
	UserID   int64
	TenantID int64
	RoleID   int64
}

func WithTenant(ctx context.Context, t *tenants.Tenant) context.Context {
	return context.WithValue(ctx, tenantKey, t)
}

func TenantFromContext(ctx context.Context) (*tenants.Tenant, bool) {
	t, ok := ctx.Value(tenantKey).(*tenants.Tenant)
	return t, ok
}

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(claimsKey).(Claims)
	return c, ok
}
