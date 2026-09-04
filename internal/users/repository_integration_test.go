package users

import (
	"context"
	"errors"
	"testing"

	"gatepass/internal/testutil"
)

func TestRepositoryByEmailAndByID(t *testing.T) {
	db := testutil.OpenMySQL(t)
	testutil.RequireTable(t, db, "users")
	ctx := context.Background()

	var id int64
	var email string
	if err := db.QueryRowContext(ctx, `SELECT id, email FROM users WHERE deleted_at IS NULL LIMIT 1`).Scan(&id, &email); err != nil {
		t.Skipf("no test user available: %v", err)
	}

	r := NewRepository(db)
	byID, err := r.ByID(ctx, id)
	if err != nil { t.Fatalf("ByID: %v", err) }
	if byID.ID != id || byID.Email != email { t.Fatalf("ByID returned %+v", byID) }

	byEmail, err := r.ByEmail(ctx, email)
	if err != nil { t.Fatalf("ByEmail: %v", err) }
	if byEmail.ID != id { t.Fatalf("ByEmail ID = %d, want %d", byEmail.ID, id) }

	if _, err := r.ByID(ctx, -999999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing ByID error = %v, want ErrNotFound", err)
	}
}

func TestRepositoryCreateDuplicateEmailIsTranslated(t *testing.T) {
	db := testutil.OpenMySQL(t)
	testutil.RequireTable(t, db, "users")
	ctx := context.Background()

	var existing string
	if err := db.QueryRowContext(ctx, `SELECT email FROM users WHERE deleted_at IS NULL LIMIT 1`).Scan(&existing); err != nil {
		t.Skipf("no existing user available: %v", err)
	}

	r := NewRepository(db)
	_, err := r.Create(ctx, &User{
		Email: existing,
		PasswordHash: "$2a$04$C6UzMDM.H6dfI/f/IKcEe.4XWn0k6A0pB9Vq2K0vRk3GQmQ6B8B7G",
		FirstName: "Test",
		LastName: "Duplicate",
	})
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate email error = %v, want ErrEmailTaken", err)
	}
}

func TestRepositoryPasswordUpdateRoundTrip(t *testing.T) {
	db := testutil.OpenMySQL(t)
	testutil.RequireTable(t, db, "users")
	ctx := context.Background()

	var id int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1`).Scan(&id); err != nil {
		t.Skipf("no test user available: %v", err)
	}

	r := NewRepository(db)
	old, err := r.ByID(ctx, id)
	if err != nil { t.Fatal(err) }
	newHash := old.PasswordHash + "x"
	if err := r.SetPasswordHash(ctx, id, newHash); err != nil { t.Fatalf("SetPasswordHash: %v", err) }
	updated, err := r.ByID(ctx, id)
	if err != nil { t.Fatal(err) }
	if updated.PasswordHash != newHash { t.Fatal("password hash was not updated") }
	if err := r.SetPasswordHash(ctx, id, old.PasswordHash); err != nil { t.Fatalf("restore password hash: %v", err) }
}
