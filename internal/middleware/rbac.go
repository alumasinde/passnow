package middleware

import (
    "net/http"
    "gatepass/internal/httpx"
    "gatepass/internal/rbac"
)

func RequireRBAC(engine *rbac.Engine, code string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
            c,ok:=ClaimsFromContext(r.Context()); if !ok { httpx.WriteError(w,httpx.ErrAuthRequired); return }
            d,err:=engine.Authorize(r.Context(),c.UserID,c.RoleID,code)
            if err!=nil { httpx.WriteError(w,httpx.ErrForbidden); return }
            if !d.Allowed { httpx.WriteError(w,httpx.ErrForbidden); return }
            next.ServeHTTP(w,r)
        })
    }
}
func RequireRBACAny(engine *rbac.Engine,codes ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
            c,ok:=ClaimsFromContext(r.Context()); if !ok { httpx.WriteError(w,httpx.ErrAuthRequired); return }
            d,err:=engine.AuthorizeAny(r.Context(),c.UserID,c.RoleID,codes...)
            if err!=nil || !d.Allowed { httpx.WriteError(w,httpx.ErrForbidden); return }
            next.ServeHTTP(w,r)
        })
    }
}
