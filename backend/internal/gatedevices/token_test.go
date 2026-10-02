package gatedevices

import (
	"testing"
	"time"
)

func TestTokenRoundTrip(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	tok, err := issueToken(secret, tokenClaims{DeviceID: 7, TenantID: 3, GateID: 2, Expires: time.Now().Add(time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	c, err := verifyToken(secret, tok)
	if err != nil || c.DeviceID != 7 || c.TenantID != 3 || c.GateID != 2 {
		t.Fatalf("round trip failed: %+v %v", c, err)
	}
	if _, err := verifyToken([]byte("a-different-secret-a-different-xx"), tok); err == nil {
		t.Fatal("token verified under the wrong secret")
	}
	if _, err := verifyToken(secret, tok+"A"); err == nil {
		t.Fatal("tampered token verified")
	}
	expired, _ := issueToken(secret, tokenClaims{DeviceID: 1, TenantID: 1, GateID: 1, Expires: time.Now().Add(-time.Second).Unix()})
	if _, err := verifyToken(secret, expired); err == nil {
		t.Fatal("expired token verified")
	}
}