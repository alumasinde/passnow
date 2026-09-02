// Package dashboard aggregates tenant-scoped operational metrics from
// visits, gatepasses, gatepass_items, and audit_logs. It queries those
// tables directly (rather than depending on the visits/gatepasses
// packages' repositories) — a dashboard is read-only aggregation, not
// domain logic, so it doesn't need those packages' business rules, only
// their tables.
package dashboard

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) count(ctx context.Context, query string, args ...any) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&n)
	return n, err
}

// Summary runs the full set of dashboard metric queries against the current tenant database.
func (r *Repository) Summary(ctx context.Context) (*Summary, error) {
	s := &Summary{}
	var err error

	if s.VisitorsToday, err = r.count(ctx, `
		SELECT COUNT(DISTINCT visitor_id) FROM visits
		WHERE status != 'cancelled'
		  AND (DATE(expected_time) = CURDATE() OR DATE(created_at) = CURDATE())`); err != nil {
		return nil, err
	}
	if s.CurrentlyOnPremises, err = r.count(ctx,
		`SELECT COUNT(*) FROM visits WHERE status = 'checked_in'`); err != nil {
		return nil, err
	}
	if s.CheckedInToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM visits
		WHERE checked_in_at IS NOT NULL AND DATE(checked_in_at) = CURDATE()`); err != nil {
		return nil, err
	}
	if s.ExpectedToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM visits
		WHERE status IN ('scheduled','expected') AND DATE(expected_time) = CURDATE()`); err != nil {
		return nil, err
	}
	if s.CompletedVisitsToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM visits
		WHERE status = 'checked_out' AND DATE(checked_out_at) = CURDATE()`); err != nil {
		return nil, err
	}
	if s.OverstayedVisits, err = r.count(ctx, `
		SELECT COUNT(*) FROM visits
		WHERE status = 'checked_in' AND expected_time IS NOT NULL AND expected_time < NOW()`); err != nil {
		return nil, err
	}

	if s.ActiveGatepasses, err = r.count(ctx, `
		SELECT COUNT(*) FROM gatepasses WHERE status IN ('approved','checked_out')`); err != nil {
		return nil, err
	}
	if s.PendingApprovals, err = r.count(ctx,
		`SELECT COUNT(*) FROM gatepasses WHERE status = 'pending_approval'`); err != nil {
		return nil, err
	}
	if s.RejectedGatepassesToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM gatepasses
		WHERE status = 'rejected' AND DATE(updated_at) = CURDATE()`); err != nil {
		return nil, err
	}
	if s.EmployeeGatepassesToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM gatepasses
		WHERE requester_type = 'employee' AND DATE(created_at) = CURDATE()`); err != nil {
		return nil, err
	}
	if s.VisitorGatepassesToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM gatepasses
		WHERE requester_type = 'visitor' AND DATE(created_at) = CURDATE()`); err != nil {
		return nil, err
	}

	if s.ItemsEnteringToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM gatepass_items gi JOIN gatepasses g ON g.id = gi.gatepass_id
		WHERE gi.direction = 'entering' AND DATE(g.created_at) = CURDATE()`); err != nil {
		return nil, err
	}
	if s.ItemsLeavingToday, err = r.count(ctx, `
		SELECT COUNT(*) FROM gatepass_items gi JOIN gatepasses g ON g.id = gi.gatepass_id
		WHERE gi.direction = 'leaving' AND DATE(g.created_at) = CURDATE()`); err != nil {
		return nil, err
	}

	if s.OverdueGatepasses, err = r.count(ctx, `
		SELECT COUNT(*) FROM gatepasses
		WHERE is_returnable = 1 AND status = 'checked_out'
		  AND expected_return_at IS NOT NULL AND expected_return_at < NOW()`); err != nil {
		return nil, err
	}

	activity, err := r.recentActivity(ctx, 15)
	if err != nil {
		return nil, err
	}
	s.RecentActivity = activity

	breakdown, err := r.gatepassTypeBreakdown(ctx)
	if err != nil {
		return nil, err
	}
	s.GatepassTypeBreakdown = breakdown

	return s, nil
}

func (r *Repository) recentActivity(ctx context.Context, limit int) ([]ActivityEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT action, entity_type, entity_id, created_at FROM audit_logs
		ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ActivityEntry
	for rows.Next() {
		var e ActivityEntry
		if err := rows.Scan(&e.Action, &e.EntityType, &e.EntityID, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}


func (r *Repository) gatepassTypeBreakdown(ctx context.Context) ([]BreakdownEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT requester_type, COUNT(*)
		FROM gatepasses
		WHERE DATE(created_at) = CURDATE()
		GROUP BY requester_type
		ORDER BY COUNT(*) DESC, requester_type ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BreakdownEntry
	for rows.Next() {
		var entry BreakdownEntry
		if err := rows.Scan(&entry.Key, &entry.Value); err != nil {
			return nil, err
		}
		entry.Label = entry.Key
		out = append(out, entry)
	}
	return out, rows.Err()
}
