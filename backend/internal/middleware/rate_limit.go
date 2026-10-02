package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
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

var trustedProxies []*net.IPNet

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
// SetTrustedProxies parses CIDRs or bare IPs, e.g. "127.0.0.1", "10.0.0.0/8".
func SetTrustedProxies(entries []string) error {
	var nets []*net.IPNet
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if !strings.Contains(e, "/") {
			if ip := net.ParseIP(e); ip != nil && ip.To4() != nil {
				e += "/32"
			} else {
				e += "/128"
			}
		}
		_, n, err := net.ParseCIDR(e)
		if err != nil {
			return fmt.Errorf("trusted proxy %q: %w", e, err)
		}
		nets = append(nets, n)
	}
	trustedProxies = nets
	return nil
}

func isTrustedProxy(ip net.IP) bool {
	for _, n := range trustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// clientIP returns the peer address, unless the peer is a trusted proxy
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if peer := net.ParseIP(host); peer != nil && isTrustedProxy(peer) {
		if v := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Client-IP"))); v != nil {
			return v.String()
		}
	}
	return host
}
