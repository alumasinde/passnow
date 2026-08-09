// Package numbering generates sequential, human-readable numbers (visit
// badge numbers, and later gatepass numbers) that are safe under
// concurrency. It works by row-locking a counter in number_sequences
// inside the CALLER's transaction (SELECT ... FOR UPDATE), incrementing,
// then formatting. Callers MUST run Next inside a transaction that also
// commits the row using the returned number — that's what makes "two
// simultaneous check-ins never get the same badge number" hold even
// though nothing here is idempotent on its own. The target table's own
// UNIQUE constraint on the formatted number is the second, independent
// safety net (never rely on the counter alone — see spec: "do not rely on
// application-level counters alone").
package numbering

import (
	"context"
	"database/sql"
	"fmt"
)

// Next increments and returns the next value in (tenantID, scope, period).
// period lets a sequence reset per year (e.g. "2026") — pass "" for a
// sequence that never resets.
func Next(ctx context.Context, tx *sql.Tx, tenantID int64, scope, period string) (int64, error) {
	// Ensure the row exists, then lock and read it. INSERT IGNORE +
	// SELECT...FOR UPDATE (rather than a single UPSERT) so the row is
	// definitely present and definitely locked before we compute the next
	// value, regardless of driver-specific upsert-and-return support.
	if _, err := tx.ExecContext(ctx, `
		INSERT IGNORE INTO number_sequences (tenant_id, scope, period, last_value)
		VALUES (?, ?, ?, 0)`, tenantID, scope, period); err != nil {
		return 0, err
	}

	var current int64
	if err := tx.QueryRowContext(ctx, `
		SELECT last_value FROM number_sequences
		WHERE tenant_id = ? AND scope = ? AND period = ? FOR UPDATE`,
		tenantID, scope, period,
	).Scan(&current); err != nil {
		return 0, err
	}

	next := current + 1
	if _, err := tx.ExecContext(ctx, `
		UPDATE number_sequences SET last_value = ?
		WHERE tenant_id = ? AND scope = ? AND period = ?`,
		next, tenantID, scope, period); err != nil {
		return 0, err
	}

	return next, nil
}

// Format renders a sequence value as "PREFIX-PERIOD-000123" (6-digit
// zero-padded), or "PREFIX-000123" if period is "". Adjust here if a
// tenant ever needs a different pattern — see settings package for making
// this tenant-configurable if/when that's needed.
func Format(prefix, period string, value int64) string {
	if period == "" {
		return fmt.Sprintf("%s-%06d", prefix, value)
	}
	return fmt.Sprintf("%s-%s-%06d", prefix, period, value)
}
