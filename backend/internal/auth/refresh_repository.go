package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

var ErrRefreshTokenInvalid = errors.New("auth: refresh token invalid or revoked")

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// NewRefreshToken generates a random 256-bit token. The raw value is
// returned once to the caller (to send to the client); only its SHA-256
// hash is ever persisted, so a DB leak doesn't hand out usable tokens.
func NewRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return raw, hash, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (r *RefreshTokenRepository) Store(ctx context.Context, userID, tenantID int64, tokenHash string, ttl time.Duration) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, tenant_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, DATE_ADD(NOW(), INTERVAL ? SECOND), NOW())`,
		userID, tenantID, tokenHash, int(ttl.Seconds()))
	return err
}

// Consume validates a raw refresh token and atomically revokes it (rotation
// on every use), returning the associated user ID. Reusing a revoked token
// is treated as invalid — this also means a stolen-and-reused token after
// the legitimate rotation will fail, which is a detectable signal worth
// alerting on (not implemented here — hook into audit).
func (r *RefreshTokenRepository) Consume(ctx context.Context, tenantID int64, raw string) (int64, error) {
	hash := hashToken(raw)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var userID, tokenTenantID int64
	var revokedAt sql.NullTime
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, tenant_id, revoked_at, expires_at
		FROM refresh_tokens
		WHERE token_hash = ? FOR UPDATE`, hash).Scan(&userID, &tokenTenantID, &revokedAt, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrRefreshTokenInvalid
		}
		return 0, err
	}

	// Validate the tenant binding before revoking the token. A token from
	// tenant A must not be usable at tenant B, and an attacker at B should
	// not be able to burn a legitimate tenant-A refresh token.
	if tokenTenantID != tenantID {
		return 0, ErrRefreshTokenInvalid
	}
	if revokedAt.Valid || time.Now().UTC().After(expiresAt) {
		return 0, ErrRefreshTokenInvalid
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = ? AND tenant_id = ?`,
		hash, tenantID); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = ? AND revoked_at IS NULL`, userID)
	return err
}
