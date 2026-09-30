package middleware

import "testing"

func TestNormalizeHost(t *testing.T) {
	if got := normalizeHost(" Acme.Example.COM. "); got != "acme.example.com" {
		t.Fatalf("unexpected normalized host: %q", got)
	}
}

func TestRateLimiter(t *testing.T) {
	l := NewRateLimiter(2, 3600*1e9)
	if !l.Allow("x") || !l.Allow("x") || l.Allow("x") {
		t.Fatal("rate limiter did not enforce limit")
	}
}
