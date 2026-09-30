package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type rateBucket struct {
	Count   int
	ResetAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]rateBucket
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 10
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{buckets: make(map[string]rateBucket), limit: limit, window: window}
}

func (l *RateLimiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok || now.After(b.ResetAt) {
		l.buckets[key] = rateBucket{Count: 1, ResetAt: now.Add(l.window)}
		return true
	}
	if b.Count >= l.limit {
		return false
	}
	b.Count++
	l.buckets[key] = b
	return true
}

func (l *RateLimiter) Middleware(keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := int64(0)
			if t, ok := reqctx.TenantFromContext(r.Context()); ok {
				tenantID = t.ID
			}
			key := keyPrefix + ":" + strconv.FormatInt(tenantID, 10) + ":" + clientIP(r)
			if !l.Allow(key) {
				httpx.WriteError(w, httpx.AppError{
					Code: "rate_limited", Message: "too many requests; try again later",
					Status: http.StatusTooManyRequests,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	// Do not trust X-Forwarded-For here. If the deployment is behind a
	// trusted proxy, normalize the proxy chain there before it reaches this
	// middleware. RemoteAddr is the safe default.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
