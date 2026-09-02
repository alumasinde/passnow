package gatepasses

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

// Worker owns tenant-local gatepass maintenance. The supplied database must
// already belong to exactly one tenant; no tenant_id column is required or
// consulted by this worker.
type Worker struct {
	db          *sql.DB
	approvedTTL time.Duration
	batch       int
	logger      *log.Logger
}

func NewWorker(db *sql.DB, approvedTTL time.Duration, logger *log.Logger) *Worker {
	if approvedTTL <= 0 {
		approvedTTL = 24 * time.Hour
	}
	if logger == nil {
		logger = log.Default()
	}
	return &Worker{db: db, approvedTTL: approvedTTL, batch: 100, logger: logger}
}

func (w *Worker) SetBatchSize(batch int) {
	if batch > 0 {
		w.batch = batch
	}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	if err := w.expireApproved(ctx); err != nil {
		w.logger.Printf("gatepass worker: expire approved passes: %v", err)
		return err
	}
	if err := w.markOverdueReturns(ctx); err != nil {
		w.logger.Printf("gatepass worker: mark overdue returns: %v", err)
		return err
	}
	return nil
}

func (w *Worker) expireApproved(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-w.approvedTTL)
	rows, err := w.db.QueryContext(ctx, `
		SELECT id
		FROM gatepasses
		WHERE status = 'approved'
		  AND issued_at IS NOT NULL
		  AND issued_at <= ?
		ORDER BY id
		LIMIT ?`, cutoff, w.batch)
	if err != nil {
		return err
	}

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(ids) == 0 {
		return nil
	}

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		res, err := tx.ExecContext(ctx, `
			UPDATE gatepasses
			SET status = 'expired', updated_at = NOW()
			WHERE id = ? AND status = 'approved'`, id)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			continue
		}
		if err := w.enqueueTx(ctx, tx, "gatepass.expired", "gatepass", id,
			"Gatepass expired before checkout", map[string]any{"gatepass_id": id}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (w *Worker) markOverdueReturns(ctx context.Context) error {
	rows, err := w.db.QueryContext(ctx, `
		SELECT id
		FROM gatepasses
		WHERE is_returnable = 1
		  AND expected_return_at IS NOT NULL
		  AND expected_return_at < UTC_TIMESTAMP()
		  AND status IN ('awaiting_return','partially_returned')
		ORDER BY id
		LIMIT ?`, w.batch)
	if err != nil {
		return err
	}

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(ids) == 0 {
		return nil
	}

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		res, err := tx.ExecContext(ctx, `
			UPDATE gatepasses
			SET status = 'return_overdue', updated_at = NOW()
			WHERE id = ?
			  AND status IN ('awaiting_return','partially_returned')`, id)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			continue
		}
		if err := w.enqueueTx(ctx, tx, "gatepass.return_overdue", "gatepass", id,
			"Gatepass return is overdue", map[string]any{"gatepass_id": id}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (w *Worker) enqueueTx(ctx context.Context, tx *sql.Tx, eventType, entityType string, entityID int64, title string, payload map[string]any) error {
	eventKey := eventType + ":" + entityType + ":" + formatInt(entityID)
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO notification_outbox
			(event_key, event_type, entity_type, entity_id, title, payload, status, attempts, available_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'pending', 0, NOW(), NOW(), NOW())
		ON DUPLICATE KEY UPDATE event_key = event_key`,
		eventKey, eventType, entityType, entityID, title, data)
	return err
}

func formatInt(v int64) string {
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
