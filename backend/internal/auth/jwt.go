// JWT access tokens. Deliberately minimal (HS256 only) and built on
// crypto/hmac + crypto/sha256 from the standard library — this is standard
// JWT *encoding*, not a custom cryptographic primitive.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrTokenExpired = errors.New("auth: token expired")
	ErrTokenInvalid = errors.New("auth: token invalid")
)

type Claims struct {
	UserID   int64 `json:"sub"`
	TenantID int64 `json:"tid"`
	RoleID   int64 `json:"rid"`
	IssuedAt int64 `json:"iat"`
	Expires  int64 `json:"exp"`
}

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func b64(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func unb64(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// IssueAccessToken mints a short-lived token scoped to exactly one tenant +
// role. A user who belongs to multiple tenants gets a DIFFERENT token per
// tenant (issued after they select which tenant to act as) — a token is
// never valid across tenants.
func IssueAccessToken(secret []byte, userID, tenantID, roleID int64, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		RoleID:   roleID,
		IssuedAt: now.Unix(),
		Expires:  now.Add(ttl).Unix(),
	}
	return sign(secret, claims)
}

func sign(secret []byte, claims Claims) (string, error) {
	h := header{Alg: "HS256", Typ: "JWT"}
	hb, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	cb, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := b64(hb) + "." + b64(cb)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)

	return signingInput + "." + b64(sig), nil
}

// VerifyAccessToken checks the signature (constant-time) and expiry, and
// returns the claims. Callers MUST still re-validate that the user's
// tenant_memberships row for claims.TenantID is still active/has that role
// before trusting RoleID for anything privileged — a token issued before a
// role change or membership revocation must not retain old privileges past
// its (short) TTL, and revocation checks close that window immediately.
func VerifyAccessToken(secret []byte, token string) (*Claims, error) {
	var parts [3]string
	n := splitJWT(token, parts[:])
	if n != 3 {
		return nil, ErrTokenInvalid
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)

	gotSig, err := unb64(parts[2])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if subtle.ConstantTimeCompare(expectedSig, gotSig) != 1 {
		return nil, ErrTokenInvalid
	}

	cb, err := unb64(parts[1])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	var claims Claims
	if err := json.Unmarshal(cb, &claims); err != nil {
		return nil, ErrTokenInvalid
	}

	if time.Now().UTC().Unix() > claims.Expires {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func splitJWT(token string, out []string) int {
	n := 0
	start := 0
	for i := 0; i < len(token) && n < len(out); i++ {
		if token[i] == '.' {
			out[n] = token[start:i]
			n++
			start = i + 1
		}
	}
	if n < len(out) {
		out[n] = token[start:]
		n++
	}
	return n
}
