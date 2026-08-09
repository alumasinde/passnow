// Package settings is the generic tenant-scoped key/value config store.
// New admin-toggleable behavior (a feature flag, a numeric limit, a
// per-module option) should be a new KEY here, not a new database column
// or a new table — this is what keeps the platform "dynamic" instead of
// requiring a migration every time Platform Admin needs a new switch.
package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

var ErrNotFound = errors.New("settings: not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Get returns the raw JSON value for a key, or ErrNotFound.
func (r *Repository) Get(ctx context.Context, tenantID int64, key string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := r.db.QueryRowContext(ctx,
		`SELECT value FROM tenant_settings WHERE tenant_id = ? AND setting_key = ? LIMIT 1`,
		tenantID, key,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return raw, nil
}

// GetBool returns the boolean value for key, or fallback if unset/invalid.
// Used for simple on/off Platform Admin toggles like pre-registration.
func (r *Repository) GetBool(ctx context.Context, tenantID int64, key string, fallback bool) bool {
	raw, err := r.Get(ctx, tenantID, key)
	if err != nil {
		return fallback
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err != nil {
		return fallback
	}
	return v
}

// GetString returns the string value for key, or fallback if unset/invalid.
func (r *Repository) GetString(ctx context.Context, tenantID int64, key string, fallback string) string {
	raw, err := r.Get(ctx, tenantID, key)
	if err != nil {
		return fallback
	}
	var v string
	if err := json.Unmarshal(raw, &v); err != nil {
		return fallback
	}
	return v
}

// Set upserts a key's value. updatedBy is the acting user's ID, for audit
// trail purposes (who flipped this switch).
func (r *Repository) Set(ctx context.Context, tenantID int64, key string, value any, updatedBy int64) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO tenant_settings (tenant_id, setting_key, value, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE value = VALUES(value), updated_by = VALUES(updated_by), updated_at = NOW()`,
		tenantID, key, raw, updatedBy,
	)
	return err
}

// All returns every setting for a tenant as key -> raw JSON, e.g. for a
// settings page that lists everything at once.
func (r *Repository) All(ctx context.Context, tenantID int64) (map[string]json.RawMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT setting_key, value FROM tenant_settings WHERE tenant_id = ?`, tenantID)
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
