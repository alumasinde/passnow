package middleware

import (
    "net/http"
    "gatepass/internal/rbac"
    "gatepass/internal/roles"
)

func Protected(secret []byte, roleRepo *roles.Repository, permission string, h http.HandlerFunc) http.Handler {
    return RequireAuth(secret)(RequirePermission(roleRepo, permission)(h))
}

func ProtectedRBAC(secret []byte, engine *rbac.Engine, permission string, h http.HandlerFunc) http.Handler {
    return RequireAuth(secret)(RequireRBAC(engine, permission)(h))
}

func ProtectedRBACAny(secret []byte, engine *rbac.Engine, permissions []string, h http.HandlerFunc) http.Handler {
    return RequireAuth(secret)(RequireRBACAny(engine, permissions...)(h))
}

func Authenticated(secret []byte, h http.HandlerFunc) http.Handler { return RequireAuth(secret)(h) }
