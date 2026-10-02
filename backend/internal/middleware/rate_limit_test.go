package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPTrustsHeaderOnlyFromTrustedProxy(t *testing.T) {
	if err := SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { trustedProxies = nil })

	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "127.0.0.1:5555"
	r.Header.Set("X-Client-IP", "203.0.113.9")
	if got := clientIP(r); got != "203.0.113.9" {
		t.Fatalf("trusted proxy: got %q", got)
	}

	r.RemoteAddr = "198.51.100.7:5555" // untrusted: header must be ignored
	if got := clientIP(r); got != "198.51.100.7" {
		t.Fatalf("untrusted peer: got %q", got)
	}

	r.RemoteAddr = "127.0.0.1:5555"
	r.Header.Set("X-Client-IP", "not-an-ip")
	if got := clientIP(r); got != "127.0.0.1" {
		t.Fatalf("garbage header: got %q", got)
	}
}

func TestSetTrustedProxiesRejectsGarbage(t *testing.T) {
	if err := SetTrustedProxies([]string{"nope"}); err == nil {
		t.Fatal("expected error")
	}
}