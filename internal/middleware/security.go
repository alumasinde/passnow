package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
)

// SecurityHeaders applies headers that are safe for the API regardless of
// whether TLS is terminated by PassNow or by an upstream reverse proxy.
// HSTS is intentionally not set here because TLS may be terminated upstream.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// RequestID ensures every request has a short opaque correlation identifier.
// Clients may supply X-Request-ID for tracing, but values are capped and
// restricted to printable ASCII so the response header cannot be abused for
// header injection.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := sanitizeRequestID(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// Recover converts an otherwise process-crashing handler panic into a generic
// 500 response. The panic value is logged server-side but never exposed to a
// client because it can contain credentials, SQL, filesystem paths or other
// internal details.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Printf("http panic request_id=%s method=%s path=%s: %v", w.Header().Get("X-Request-ID"), r.Method, r.URL.Path, v)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":{"code":"internal_error","message":"something went wrong, please try again"}}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func sanitizeRequestID(value string) string {
	if len(value) == 0 || len(value) > 64 {
		return ""
	}
	for _, b := range []byte(value) {
		if b < 0x21 || b > 0x7e {
			return ""
		}
	}
	return value
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b[:])
}
