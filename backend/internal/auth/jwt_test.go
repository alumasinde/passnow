package auth

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	secret := []byte("test-secret")
	token, err := IssueAccessToken(secret, 11, 22, 33, time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	claims, err := VerifyAccessToken(secret, token)
	if err != nil {
		t.Fatalf("VerifyAccessToken: %v", err)
	}
	if claims.UserID != 11 || claims.TenantID != 22 || claims.RoleID != 33 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.Expires <= claims.IssuedAt {
		t.Fatalf("expiry must be after issue time: %+v", claims)
	}
}

func TestAccessTokenRejectsWrongSecretAndTampering(t *testing.T) {
	secret := []byte("correct-secret")
	token, err := IssueAccessToken(secret, 1, 2, 3, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyAccessToken([]byte("wrong-secret"), token); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("wrong secret error = %v, want invalid token", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(parts[1]) == 0 {
		t.Fatalf("unexpected JWT format")
	}
	replacement := byte('A')
	if parts[1][0] == replacement {
		replacement = 'B'
	}
	parts[1] = string(replacement) + parts[1][1:]
	tampered := strings.Join(parts, ".")
	if _, err := VerifyAccessToken(secret, tampered); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("tampered token error = %v, want invalid token", err)
	}
}

func TestAccessTokenExpiryAndMalformedTokens(t *testing.T) {
	secret := []byte("expiry-secret")
	token, err := IssueAccessToken(secret, 1, 2, 3, -time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyAccessToken(secret, token); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expired token error = %v, want ErrTokenExpired", err)
	}

	for _, token := range []string{"", "a", "a.b", "a.b.c.d", "not.a.jwt"} {
		if _, err := VerifyAccessToken(secret, token); !errors.Is(err, ErrTokenInvalid) {
			t.Fatalf("token %q error = %v, want ErrTokenInvalid", token, err)
		}
	}
}
