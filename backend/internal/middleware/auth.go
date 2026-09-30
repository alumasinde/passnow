package middleware

import (
	"context"
	"net/http"
	"strings"

	"gatepass/internal/auth"
	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
	"gatepass/internal/roles"
)

// ClaimsFromContext is a thin re-export of reqctx for callers already in
// this package. The auth package (handlers) reads reqctx directly instead
// of importing middleware, to avoid an import cycle (middleware imports
// auth for token verification).
func ClaimsFromContext(ctx context.Context) (reqctx.Claims, bool) {
	return reqctx.ClaimsFromContext(ctx)
}

// RequireAuth verifies the bearer JWT and confirms its tenant claim
// matches the tenant resolved for THIS request's Host/path (set by
// ResolveTenant, which must run earlier in the chain). A token issued for
// tenant A is rejected outright on tenant B's domain, even if otherwise
// valid and unexpired — this is the hard stop against cross-tenant token
// replay.
func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Bearer ") {
				httpx.WriteError(w, httpx.ErrAuthRequired)
				return
			}
			token := strings.TrimPrefix(authz, "Bearer ")

			claims, err := auth.VerifyAccessToken(secret, token)
			if err != nil {
				httpx.WriteError(w, httpx.ErrAuthRequired)
				return
			}

			tenant, ok := TenantFromContext(r.Context())
			if !ok || tenant.ID != claims.TenantID {
				httpx.WriteError(w, httpx.ErrAuthRequired)
				return
			}

			ctx := reqctx.WithClaims(r.Context(), reqctx.Claims{
				UserID:   claims.UserID,
				TenantID: claims.TenantID,
				RoleID:   claims.RoleID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission re-reads the caller's CURRENT permissions from the DB
// on every call (not just from the JWT) so a permission/role revoked mid-
// session takes effect immediately rather than waiting for token expiry.
// This costs a query per privileged request; that is the correct trade-off
// for an approvals/access-control system over trusting a cached claim.
func RequirePermission(roleRepo *roles.Repository, code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, httpx.ErrAuthRequired)
				return
			}

			membership, err := roleRepo.MembershipFor(r.Context(), claims.UserID)
			if err != nil || !membership.IsActive() || membership.RoleID != claims.RoleID {
				httpx.WriteError(w, httpx.ErrForbidden)
				return
			}

			perms, err := roleRepo.PermissionCodesForRole(r.Context(), membership.RoleID)
			if err != nil {
				httpx.WriteError(w, httpx.ErrInternal)
				return
			}
			if !perms[code] {
				httpx.WriteError(w, httpx.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


// RequireAnyPermission grants access when the caller has at least one of the
// supplied permissions. It is useful when an operation needs a narrow
// cross-module lookup (for example selecting a host while creating a visit)
// without granting full access to the other module.
func RequireAnyPermission(roleRepo *roles.Repository, codes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok { httpx.WriteError(w, httpx.ErrAuthRequired); return }
			membership, err := roleRepo.MembershipFor(r.Context(), claims.UserID)
			if err != nil || !membership.IsActive() || membership.RoleID != claims.RoleID {
				httpx.WriteError(w, httpx.ErrForbidden); return
			}
			perms, err := roleRepo.PermissionCodesForRole(r.Context(), membership.RoleID)
			if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
			for _, code := range codes { if perms[code] { next.ServeHTTP(w, r); return } }
			httpx.WriteError(w, httpx.ErrForbidden)
		})
	}
}
