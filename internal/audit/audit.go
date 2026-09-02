// Package audit is the append-only security event log. Every module calls
// Record for its sensitive actions (created/updated/approved/etc.) — this
// package owns the ONE table and ONE shape for that, so audit events don't
// end up scattered and inconsistently formatted across modules.
package audit

import (
	"context"
	"database/sql"
	"encoding/json"
)

// Action codes. Add new ones here as modules are built, matching the
// action list in the system spec (GATEPASS_APPROVED, ITEM_ADDED, etc.).
const (
	ActionVisitorCreated = "VISITOR_CREATED"
	ActionVisitorUpdated = "VISITOR_UPDATED"
)

// Querier is satisfied by both *sql.DB and *sql.Tx, so callers can record
// an audit entry either standalone or as part of the same transaction as
// the mutation it describes (preferred — keeps them atomic).
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// DB satisfies Querier for callers that don't need a transaction.
func (r *Repository) DB() *sql.DB { return r.db }

type Entry struct {
	ActorUserID *int64
	Action      string
	EntityType  string
	EntityID    *int64
	RequestID   string
	IPAddress   string
	UserAgent   string
	Metadata    map[string]any
}

// Record inserts one audit row. Never returns a "soft" failure silently to
// the caller as success — if this errors, the caller decides whether that
// should fail the whole request (recommended: wrap in the same tx as the
// mutating write so a broken audit log write rolls back the action too).
func (r *Repository) Record(ctx context.Context, q Querier, e Entry) error {
	var metaJSON []byte
	if e.Metadata != nil {
		b, err := json.Marshal(e.Metadata)
		if err != nil {
			return err
		}
		metaJSON = b
	}

	_, err := q.ExecContext(ctx, `
		INSERT INTO audit_logs
			(actor_user_id, action, entity_type, entity_id, request_id, ip_address, user_agent, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		e.ActorUserID, e.Action, e.EntityType, e.EntityID,
		nullIfEmpty(e.RequestID), nullIfEmpty(e.IPAddress), nullIfEmpty(e.UserAgent), metaJSON,
	)
	return err
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
