package gatedevices

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("gatedevices: device token invalid or expired")

const TokenTTL = 12 * time.Hour

type tokenClaims struct {
	DeviceID int64 `json:"d"`
	TenantID int64 `json:"t"`
	GateID   int64 `json:"g"`
	Expires  int64 `json:"e"`
}

// Domain-separated key so a gate-device token can never be confused with a user JWT.
func signingKey(secret []byte) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte("gate-device-token-v1"))
	return m.Sum(nil)
}

func sign(secret []byte, payload string) []byte {
	m := hmac.New(sha256.New, signingKey(secret))
	m.Write([]byte(payload))
	return m.Sum(nil)
}

func issueToken(secret []byte, c tokenClaims) (string, error) {
	body, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	p := base64.RawURLEncoding.EncodeToString(body)
	return "gd1." + p + "." + base64.RawURLEncoding.EncodeToString(sign(secret, p)), nil
}

func verifyToken(secret []byte, token string) (*tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "gd1" {
		return nil, ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(sig, sign(secret, parts[1])) {
		return nil, ErrInvalidToken
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var c tokenClaims
	if err := json.Unmarshal(body, &c); err != nil || time.Now().Unix() > c.Expires {
		return nil, ErrInvalidToken
	}
	return &c, nil
}

// Verifier issues and checks device tokens. A token only proves identity;
// the device row is re-read on every use, so deactivating a device kills its
// tokens immediately.
type Verifier struct {
	repo   *Repository
	secret []byte
}

func NewVerifier(repo *Repository, secret []byte) *Verifier {
	return &Verifier{repo: repo, secret: secret}
}

func (v *Verifier) IssueToken(d *Device, tenantID int64) (string, time.Time, error) {
	exp := time.Now().UTC().Add(TokenTTL)
	tok, err := issueToken(v.secret, tokenClaims{DeviceID: d.ID, TenantID: tenantID, GateID: d.GateID, Expires: exp.Unix()})
	return tok, exp, err
}

// VerifyDevice satisfies the small interfaces declared in gatepasses and visits.
func (v *Verifier) VerifyDevice(ctx context.Context, tenantID int64, token string) (deviceID, gateID int64, deviceKey string, err error) {
	c, err := verifyToken(v.secret, token)
	if err != nil || c.TenantID != tenantID {
		return 0, 0, "", ErrInvalidToken
	}
	d, err := v.repo.ByID(ctx, c.DeviceID)
	if err != nil || !d.Active || d.GateID != c.GateID {
		return 0, 0, "", ErrInvalidToken
	}
	return d.ID, d.GateID, d.DeviceKey, nil
}