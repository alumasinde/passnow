package middleware

import (
    "net/http"
    "strings"

    "gatepass/internal/auth"
    "gatepass/internal/httpx"
    "gatepass/internal/platform"
    "gatepass/internal/reqctx"
)

// PlatformAdmin authenticates a PassNow platform operator. Platform tokens
// are deliberately marked with tenant_id=0, so tenant tokens can never be
// replayed against platform routes.
func PlatformAdmin(secret []byte, admins *platform.AdminRepository, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authz := r.Header.Get("Authorization")
        if !strings.HasPrefix(authz, "Bearer ") {
            httpx.WriteError(w, httpx.ErrAuthRequired)
            return
        }

        claims, err := auth.VerifyAccessToken(secret, strings.TrimPrefix(authz, "Bearer "))
        if err != nil || claims.TenantID != 0 {
            httpx.WriteError(w, httpx.ErrAuthRequired)
            return
        }

        admin, err := admins.ByUserID(r.Context(), claims.UserID)
        if err != nil || admin.Status != "active" {
            httpx.WriteError(w, httpx.ErrAuthRequired)
            return
        }

        ctx := reqctx.WithClaims(r.Context(), reqctx.Claims{UserID: claims.UserID, TenantID: 0, RoleID: claims.RoleID})
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
