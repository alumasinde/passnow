package middleware

import (
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// RequestLogger provides one consistent backend trace for every request.
// It logs failures with the method, path, HTTP status and elapsed time.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rw, r)
		status := rw.status
		if status == 0 {
			status = http.StatusOK
		}
		if status >= 400 {
			log.Printf("HTTP REQUEST FAILED: method=%s path=%q status=%d duration=%s", r.Method, r.URL.Path, status, time.Since(start).Round(time.Millisecond))
			return
		}
		log.Printf("HTTP REQUEST: method=%s path=%q status=%d duration=%s", r.Method, r.URL.Path, status, time.Since(start).Round(time.Millisecond))
	})
}
