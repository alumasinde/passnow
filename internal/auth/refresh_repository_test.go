package auth

import (
	"context"
	"testing"
	"time"

	"gatepass/internal/testutil"
)

func TestNewRefreshTokenProperties(t *testing.T) {
	raw, hash, err := NewRefreshToken()
	if err != nil { t.Fatalf("NewRefreshToken: %v", err) }
	if len(raw) != 64 { t.Fatalf("raw token length = %d, want 64 hex chars", len(raw)) }
	if len(hash) != 64 { t.Fatalf("hash length = %d, want 64 hex chars", len(hash)) }
	if raw == hash { t.Fatal("raw token must differ from persisted hash") }
	if got := hashToken(raw); got != hash { t.Fatalf("hashToken(raw) = %q, want %q", got, hash) }
}

func TestRefreshTokenRepositoryLifecycle(t *testing.T) {
	db := testutil.OpenMySQL(t)
	testutil.RequireTable(t, db, "refresh_tokens")
	testutil.RequireTable(t, db, "users")

	ctx := context.Background()
	userID := testutil.MustQueryInt(t, db, `SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1`)
	if userID < 1 { t.Skip("no active test user available in configured MySQL database") }

	repo := NewRefreshTokenRepository(db)
	raw, hash, err := NewRefreshToken()
	if err != nil { t.Fatal(err) }

	if err := repo.Store(ctx, userID, hash, time.Hour); err != nil { t.Fatalf("Store: %v", err) }
	gotUser, err := repo.Consume(ctx, raw)
	if err != nil { t.Fatalf("Consume: %v", err) }
	if gotUser != userID { t.Fatalf("Consume userID = %d, want %d", gotUser, userID) }

	if _, err := repo.Consume(ctx, raw); err != ErrRefreshTokenInvalid {
		t.Fatalf("reused token error = %v, want ErrRefreshTokenInvalid", err)
	}
}

func TestRefreshTokenRepositoryRejectsUnknownAndExpiredTokens(t *testing.T) {
	db := testutil.OpenMySQL(t)
	testutil.RequireTable(t, db, "refresh_tokens")
	testutil.RequireTable(t, db, "users")

	ctx := context.Background()
	userID := testutil.MustQueryInt(t, db, `SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1`)
	if userID < 1 { t.Skip("no active test user available in configured MySQL database") }

	repo := NewRefreshTokenRepository(db)
	if _, err := repo.Consume(ctx, "unknown-token"); err != ErrRefreshTokenInvalid {
		t.Fatalf("unknown token error = %v, want ErrRefreshTokenInvalid", err)
	}

	raw, hash, err := NewRefreshToken()
	if err != nil { t.Fatal(err) }
	if err := repo.Store(ctx, userID, hash, -time.Second); err != nil { t.Fatal(err) }
	if _, err := repo.Consume(ctx, raw); err != ErrRefreshTokenInvalid {
		t.Fatalf("expired token error = %v, want ErrRefreshTokenInvalid", err)
	}
}

func TestRefreshTokenRepositoryRevokesAllForUser(t *testing.T) {
	db := testutil.OpenMySQL(t)
	testutil.RequireTable(t, db, "refresh_tokens")
	testutil.RequireTable(t, db, "users")

	ctx := context.Background()
	userID := testutil.MustQueryInt(t, db, `SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1`)
	if userID < 1 { t.Skip("no active test user available in configured MySQL database") }

	repo := NewRefreshTokenRepository(db)
	raw, hash, err := NewRefreshToken()
	if err != nil { t.Fatal(err) }
	if err := repo.Store(ctx, userID, hash, time.Hour); err != nil { t.Fatal(err) }
	if err := repo.RevokeAllForUser(ctx, userID); err != nil { t.Fatalf("RevokeAllForUser: %v", err) }
	if _, err := repo.Consume(ctx, raw); err != ErrRefreshTokenInvalid {
		t.Fatalf("revoked token error = %v, want ErrRefreshTokenInvalid", err)
	}
}
