package users

import (
	"context"
	"database/sql"
	"errors"

	"gatepass/internal/database"
)

var (
	ErrNotFound   = errors.New("users: not found")
	ErrEmailTaken = errors.New("users: email already registered")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const selectCols = `
	id, email, password_hash, first_name, last_name, status,
	failed_login_count, locked_until, created_at, updated_at, deleted_at
`

func (r *Repository) scan(row interface{ Scan(dest ...any) error }) (*User, error) {
	var u User
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status,
		&u.FailedLoginCount, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) ByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+selectCols+" FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1", email)
	return r.scan(row)
}

func (r *Repository) ByID(ctx context.Context, id int64) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+selectCols+" FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1", id)
	return r.scan(row)
}

// Create inserts a new user with an already-hashed password. Returns
// ErrEmailTaken on the unique-key violation rather than leaking the raw
// MySQL error (which would confirm email existence to a probing caller in
// logs, not response — the handler layer decides what the client sees).
func (r *Repository) Create(ctx context.Context, u *User) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', NOW(), NOW())`,
		u.Email, u.PasswordHash, u.FirstName, u.LastName,
	)
	if err != nil {
		if database.IsDuplicateKeyErr(err) {
			return 0, ErrEmailTaken
		}
		return 0, err
	}
	return res.LastInsertId()
}

// CreateTx is Create run inside an existing transaction (platform
// bootstrap use).
func (r *Repository) CreateTx(ctx context.Context, tx *sql.Tx, u *User) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', NOW(), NOW())`,
		u.Email, u.PasswordHash, u.FirstName, u.LastName,
	)
	if err != nil {
		if database.IsDuplicateKeyErr(err) {
			return 0, ErrEmailTaken
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateName(ctx context.Context, id int64, firstName, lastName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET first_name = ?, last_name = ?, updated_at = NOW() WHERE id = ?`,
		firstName, lastName, id)
	return err
}

// RegisterFailedLogin increments the counter and, once it crosses the
// threshold, sets locked_until. Called inside the same transaction as the
// credential check to avoid a race that lets an attacker exceed the limit
// via concurrent requests.
func (r *Repository) RegisterFailedLogin(ctx context.Context, tx *sql.Tx, userID int64, lockDurationSeconds, threshold int) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE users
		SET failed_login_count = failed_login_count + 1,
		    locked_until = CASE
		        WHEN failed_login_count + 1 >= ? THEN DATE_ADD(NOW(), INTERVAL ? SECOND)
		        ELSE locked_until
		    END
		WHERE id = ?`, threshold, lockDurationSeconds, userID)
	return err
}

func (r *Repository) ResetFailedLogins(ctx context.Context, tx *sql.Tx, userID int64) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET failed_login_count = 0, locked_until = NULL WHERE id = ?`, userID)
	return err
}

func (r *Repository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
