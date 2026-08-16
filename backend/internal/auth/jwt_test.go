package auth

import (
	"strings"
	"testing"
	"time"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	token, err := IssueAccessToken(secret, 7, 11, 13, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := VerifyAccessToken(secret, token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 7 || claims.TenantID != 11 || claims.RoleID != 13 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestAccessTokenRejectsWrongSecret(t *testing.T) {
	token, err := IssueAccessToken([]byte("01234567890123456789012345678901"), 1, 2, 3, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyAccessToken([]byte("abcdefghijklmnopqrstuvwxyz123456"), token); err != ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestAccessTokenRejectsTampering(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	token, err := IssueAccessToken(secret, 1, 2, 3, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	parts[1] = parts[1] + "x"
	tampered := strings.Join(parts, ".")
	if _, err := VerifyAccessToken(secret, tampered); err != ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestAccessTokenRejectsExpiredToken(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	token, err := IssueAccessToken(secret, 1, 2, 3, -time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyAccessToken(secret, token); err != ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}
