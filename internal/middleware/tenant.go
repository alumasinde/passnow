package middleware

import (
	"context"
	"net/http"
	"strings"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
	"gatepass/internal/tenants"
)

// TenantFromContext returns the resolved tenant. Every handler and
// repository call downstream of this middleware MUST use this — never trust
// a tenant_id from the request body/query/header. If this returns false,
// the middleware chain is misconfigured; fail closed (401/500), never proceed.
// (Thin re-export of reqctx so existing callers in this package don't change.)
func TenantFromContext(ctx context.Context) (*tenants.Tenant, bool) {
	return reqctx.TenantFromContext(ctx)
}

// ResolveTenant identifies the tenant for an inbound request in this order:
//  1. Custom domain: Host header matches a verified tenants.custom_domain.
//  2. Subdomain: Host is "<slug>.<baseDomain>".
//  3. Path prefix: first path segment is "/<slug>/...". Used as a fallback
//     for local/dev access or clients that can't do per-tenant DNS.
//
// baseDomain is the platform's own domain, e.g. "gatepass.example.com".
func ResolveTenant(repo *tenants.Repository, baseDomain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			host := normalizeHost(stripPort(r.Host))
			base := normalizeHost(baseDomain)

			var (
				t   *tenants.Tenant
				err error
			)

			switch {
			case host != "" && !strings.HasSuffix(host, "."+base) && host != base:
				// Resolve every registered domain first, including PassNow subdomains.
				t, err = repo.ByDomain(ctx, host)
				if err != nil { t, err = repo.ByCustomDomain(ctx, host) }

			case strings.HasSuffix(host, "."+base):
				t, err = repo.ByDomain(ctx, host)
				if t == nil {
					sub := strings.TrimSuffix(host, "."+baseDomain)
					if sub != "" && sub != "www" { t, err = repo.BySlug(ctx, sub) }
				}
			}

			// Path-prefix fallback if no host-based match was found/attempted.
			if t == nil {
				slug, rest, ok := firstPathSegment(r.URL.Path)
				if ok {
					if pt, perr := repo.BySlug(ctx, slug); perr == nil {
						t = pt
						// Strip the tenant segment so downstream routers see
						// a normal "/api/v1/..." path.
						r.URL.Path = rest
					} else {
						err = perr
					}
				}
			}

			if t == nil || !t.IsActive() {
				httpx.WriteError(w, httpx.ErrTenantNotFound)
				return
			}
			if err != nil && t == nil {
				httpx.WriteError(w, httpx.ErrTenantNotFound)
				return
			}

			ctx = reqctx.WithTenant(ctx, t)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func stripPort(host string) string {
	if i := strings.IndexByte(host, ':'); i != -1 {
		return host[:i]
	}
	return host
}

func normalizeHost(host string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
}

// firstPathSegment splits "/acme/api/v1/visitors" into ("acme",
// "/api/v1/visitors", true). Returns ok=false for paths like "/api/v1/..."
// that aren't tenant-prefixed (e.g. platform-level admin routes).
func firstPathSegment(path string) (slug string, rest string, ok bool) {
	trimmed := strings.TrimPrefix(path, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 || parts[0] == "" || parts[0] == "api" {
		return "", path, false
	}
	slug = strings.ToLower(parts[0])
	if len(parts) == 2 {
		rest = "/" + parts[1]
	} else {
		rest = "/"
	}
	return slug, rest, true
}
