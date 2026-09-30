// Package settings is the generic key/value config store for a tenant database.
// New admin-toggleable behavior (a feature flag, a numeric limit, a
// per-module option) should be a new key here, not a new database column
// or a new table. Tenant isolation is provided by the database connection.
package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

var ErrNotFound = errors.New("settings: not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Get returns the raw JSON value for a key, or ErrNotFound.
func (r *Repository) Get(ctx context.Context, key string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := r.db.QueryRowContext(ctx,
		`SELECT value FROM tenant_settings WHERE setting_key = ? LIMIT 1`,
		key,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return raw, nil
}

func (r *Repository) GetBool(ctx context.Context, key string, fallback bool) bool {
	raw, err := r.Get(ctx, key)
	if err != nil || len(raw) == 0 { return fallback }
	var v bool
	if err := json.Unmarshal(raw, &v); err == nil { return v }
	// Be tolerant of legacy/manual SQL values stored without JSON quoting.
	s := strings.Trim(strings.TrimSpace(string(raw)), "\\\"")
	if parsed, err := strconv.ParseBool(s); err == nil { return parsed }
	if s == "1" { return true }
	if s == "0" { return false }
	return fallback
}

func (r *Repository) GetString(ctx context.Context, key string, fallback string) string {
	raw, err := r.Get(ctx, key)
	if err != nil || len(raw) == 0 { return fallback }
	var v string
	if err := json.Unmarshal(raw, &v); err == nil {
		v = strings.TrimSpace(v)
		if v != "" { return v }
	}
	// Be tolerant of legacy/manual SQL values stored without JSON quoting.
	v = strings.Trim(strings.TrimSpace(string(raw)), "\\\"")
	if v != "" && v != "null" { return v }
	return fallback
}

// Set upserts a key's value. updatedBy is the acting user's ID.
func (r *Repository) Set(ctx context.Context, key string, value any, updatedBy int64) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO tenant_settings (setting_key, value, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE value = VALUES(value), updated_by = VALUES(updated_by), updated_at = NOW()`,
		key, raw, updatedBy,
	)
	return err
}

// All returns every setting as key -> raw JSON.
func (r *Repository) All(ctx context.Context) (map[string]json.RawMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT setting_key, value FROM tenant_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]json.RawMessage)
	for rows.Next() {
		var k string
		var v json.RawMessage
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}
