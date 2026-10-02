package middleware

import (
	"context"
	"net/http"
	"strings"
	"errors"
	"log"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
	"gatepass/internal/tenants"
)


func TenantFromContext(ctx context.Context) (*tenants.Tenant, bool) {
	return reqctx.TenantFromContext(ctx)
}

func ResolveTenant(repo *tenants.Repository, baseDomain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			host := normalizeHost(stripPort(r.Host))
			base := normalizeHost(baseDomain)

			var t *tenants.Tenant

			// dbErr remembers the first REAL lookup failure (anything other than
			// "not found"), so a platform-database outage is answered with 503
			// instead of being reported as "tenant not found".
			var dbErr error
			note := func(err error) {
				if err != nil && !errors.Is(err, tenants.ErrNotFound) && dbErr == nil {
					dbErr = err
				}
			}

			switch {
			case host != "" && !strings.HasSuffix(host, "."+base) && host != base:
				// Resolve every registered domain first, including PassNow subdomains.
				var err error
				t, err = repo.ByDomain(ctx, host)
				note(err)
				if t == nil {
					t, err = repo.ByCustomDomain(ctx, host)
					note(err)
				}

			case strings.HasSuffix(host, "."+base):
				var err error
				t, err = repo.ByDomain(ctx, host)
				note(err)
				if t == nil {
					sub := strings.TrimSuffix(host, "."+base)
					if sub != "" && sub != "www" {
						t, err = repo.BySlug(ctx, sub)
						note(err)
					}
				}
			}

			// Explicit tenant slug is useful for local development where the
			// browser is served from one IP/host for every tenant.
			if t == nil {
				if slug := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Tenant-Slug"))); slug != "" {
					var err error
					t, err = repo.BySlug(ctx, slug)
					note(err)
				}
			}

			// Path-prefix fallback supports direct browser access on one local/IP
			// host. It also strips a matching slug even when the tenant was already
			// resolved from the Host header, allowing the same public-media URL to
			// work in development and production.
			if slug, rest, ok := firstPathSegment(r.URL.Path); ok {
				if t == nil {
					pt, perr := repo.BySlug(ctx, slug)
					note(perr)
					if perr == nil {
						t = pt
						r.URL.Path = rest
					}
				} else if strings.EqualFold(slug, t.Slug) {
					r.URL.Path = rest
				}
			}

			if t == nil {
				if dbErr != nil {
					log.Printf("tenant resolution failed: host=%q err=%v", host, dbErr)
					httpx.WriteError(w, httpx.ErrServiceUnavailable)
					return
				}
				httpx.WriteError(w, httpx.ErrTenantNotFound)
				return
			}
			if !t.IsActive() {
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
