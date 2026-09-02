package gatepasses

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

// Worker owns scheduling/retries only. It does NOT contain Gatepass business
// rules; those remain in the repository/domain layer. This keeps concurrency
// and operational concerns separate from request-time business logic.
type Worker struct {
	db          *sql.DB
	interval    time.Duration
	approvedTTL time.Duration
	batch       int
	logger      *log.Logger
}

func NewWorker(db *sql.DB, interval, approvedTTL time.Duration, logger *log.Logger) *Worker {
	if interval <= 0 {
		interval = time.Minute
	}
	if logger == nil {
		logger = log.Default()
	}
	if approvedTTL <= 0 {
		approvedTTL = 24 * time.Hour
	}
	return &Worker{db: db, interval: interval, approvedTTL: approvedTTL, batch: 100, logger: logger}
}

func (w *Worker) Run(ctx context.Context) {
	// Run once immediately so restarts do not wait for the first tick.
	w.runOnce(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *Worker) runOnce(ctx context.Context) {
	if err := w.expireApproved(ctx); err != nil {
		w.logger.Printf("gatepass worker: expire approved passes: %v", err)
	}
	if err := w.markOverdueReturns(ctx); err != nil {
		w.logger.Printf("gatepass worker: mark overdue returns: %v", err)
	}
}

// expireApproved expires approved passes that have remained unused beyond
// the configured operational TTL. issued_at is the approval/issuance timestamp;
// checked-out passes are no longer eligible for this expiry path. Per-tenant business settings can be
// introduced later without changing the worker architecture.
func (w *Worker) expireApproved(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-w.approvedTTL)
	_, err := w.db.ExecContext(ctx, `
		UPDATE gatepasses
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'approved'
		  AND issued_at IS NOT NULL
		  AND issued_at <= ?
		LIMIT 100`, cutoff)
	if err != nil {
		return err
	}

	rows, err := w.db.QueryContext(ctx, `
		SELECT id
		FROM gatepasses
		WHERE status = 'expired'
		  AND updated_at >= UTC_TIMESTAMP() - INTERVAL 1 MINUTE
		ORDER BY id DESC
		LIMIT ?`, w.batch)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := w.enqueue(ctx, "gatepass.expired", "gatepass", id,
			"Gatepass expired before checkout", map[string]any{"gatepass_id": id}); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (w *Worker) markOverdueReturns(ctx context.Context) error {
	// Only passes that are physically out can become overdue. Partial returns
	// are included so the system can continue to alert until all quantities
	// are accounted for.
	_, err := w.db.ExecContext(ctx, `
		UPDATE gatepasses
		SET status = 'return_overdue', updated_at = NOW()
		WHERE is_returnable = 1
		  AND expected_return_at IS NOT NULL
		  AND expected_return_at < UTC_TIMESTAMP()
		  AND status IN ('awaiting_return','partially_returned')
		LIMIT 100`)
	if err != nil {
		return err
	}

	rows, err := w.db.QueryContext(ctx, `
		SELECT id
		FROM gatepasses
		WHERE status = 'return_overdue'
		  AND updated_at >= UTC_TIMESTAMP() - INTERVAL 1 MINUTE
		ORDER BY id DESC
		LIMIT ?`, w.batch)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := w.enqueue(ctx, "gatepass.return_overdue", "gatepass", id,
			"Gatepass return is overdue", map[string]any{"gatepass_id": id}); err != nil {
			return err
		}
	}
	return rows.Err()
}

// enqueue is idempotent. Multiple application instances can run the same
// worker safely because the unique event key prevents duplicate notifications.
func (w *Worker) enqueue(ctx context.Context, eventType, entityType string, entityID int64, title string, payload map[string]any) error {
	eventKey := eventType + ":" + entityType + ":" + formatInt(entityID)
	data, err := jsonMarshal(payload)
	if err != nil {
		return err
	}

	_, err = w.db.ExecContext(ctx, `
		INSERT INTO notification_outbox
			(event_key, event_type, entity_type, entity_id, title, payload, status, attempts, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 'pending', 0, NOW())
		ON DUPLICATE KEY UPDATE event_key = event_key`,
		eventKey, eventType, entityType, entityID, title, data)
	return err
}

func formatInt(v int64) string {
	// Avoid fmt import solely for this helper in the worker hot path.
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = digits[v%10]
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func jsonMarshal(v any) ([]byte, error) {
	// Local wrapper keeps worker's persistence code easy to test/replace.
	return json.Marshal(v)
}
