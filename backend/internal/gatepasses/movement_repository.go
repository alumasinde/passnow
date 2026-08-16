package gatepasses

import (
	"context"
	"database/sql"
	"errors"
)

type MovementRepository struct {
	db *sql.DB
}

func NewMovementRepository(db *sql.DB) *MovementRepository {
	return &MovementRepository{db: db}
}

func (r *MovementRepository) List(ctx context.Context, tenantID, gatepassID int64) ([]MovementDTO, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, actor_user_id, gate_name, notes, occurred_at
		FROM gatepass_movements
		WHERE tenant_id = ? AND gatepass_id = ?
		ORDER BY occurred_at, id`, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MovementDTO
	for rows.Next() {
		var m MovementDTO
		if err := rows.Scan(&m.ID, &m.Type, &m.ActorUserID, &m.GateName, &m.Notes, &m.OccurredAt); err != nil {
			return nil, err
		}
		items, err := r.listItems(ctx, tenantID, m.ID)
		if err != nil {
			return nil, err
		}
		m.Items = items
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MovementRepository) listItems(ctx context.Context, tenantID, movementID int64) ([]MovementItemDTO, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, gatepass_item_id, quantity, outcome, condition, notes
		FROM gatepass_movement_items
		WHERE tenant_id = ? AND movement_id = ?
		ORDER BY id`, tenantID, movementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MovementItemDTO
	for rows.Next() {
		var i MovementItemDTO
		if err := rows.Scan(&i.ID, &i.GatepassItemID, &i.Quantity, &i.Outcome, &i.Condition, &i.Notes); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// Checkout records the physical departure and atomically changes the
// gatepass status. The gatepass row is locked first; item rows are then
// locked before the movement ledger is written.
func (r *MovementRepository) Checkout(ctx context.Context, tenantID, gatepassID, actorUserID int64, direction string, in MovementInput) (*Gatepass, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	g, err := r.lockGatepass(ctx, tx, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}
	if !g.CanCheckOut(Direction(direction)) {
		return nil, ErrMovementNotAllowed
	}

	itemIDs, err := r.lockGatepassItems(ctx, tx, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}
	// A checkout does not accept arbitrary item quantities. The authorized
	// gatepass items are the source of truth; the movement ledger records all
	// leaving quantities in one atomic operation.
	if len(in.Items) > 0 {
		for _, line := range in.Items {
			if _, ok := itemIDs[line.GatepassItemID]; !ok {
				return nil, ErrReturnItemInvalid
			}
		}
	}
	if len(itemIDs) == 0 {
		// A physical checkout can still be valid for a visitor-only pass;
		// there simply are no item movement lines to write.
	}

	movementID, err := r.insertMovement(ctx, tx, tenantID, gatepassID, MovementCheckout, actorUserID, in)
	if err != nil {
		return nil, err
	}

	for itemID, qty := range itemIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO gatepass_movement_items
				(tenant_id, movement_id, gatepass_item_id, quantity, outcome, created_at)
			VALUES (?, ?, ?, ?, 'released', NOW())`,
			tenantID, movementID, itemID, qty); err != nil {
			return nil, err
		}
	}

	next := g.NextCheckOutStatus()
	if _, err := tx.ExecContext(ctx, `
		UPDATE gatepasses
		SET status = ?, checked_out_at = NOW(), checked_out_by = ?, updated_at = NOW()
		WHERE id = ? AND tenant_id = ?`, next, actorUserID, gatepassID, tenantID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.getGatepass(ctx, tenantID, gatepassID)
}

// Checkin applies a full or partial physical return. Returned/damaged/lost
// are treated as accounted-for quantities. This means a missing item can be
// explicitly resolved without falsely recording it as physically returned.
func (r *MovementRepository) Checkin(ctx context.Context, tenantID, gatepassID, actorUserID int64, direction string, in MovementInput) (*Gatepass, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	g, err := r.lockGatepass(ctx, tx, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}
	if !g.CanCheckIn(Direction(direction)) {
		return nil, ErrMovementNotAllowed
	}
	if err := validateMovementInput(in, true); err != nil {
		return nil, err
	}

	items, err := r.lockGatepassItemsWithQuantities(ctx, tx, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}

	if in.FullReturn {
		in.Items = make([]MovementItemInput, 0, len(items))
		for id, outstanding := range items {
			if outstanding > 0 {
				in.Items = append(in.Items, MovementItemInput{
					GatepassItemID: id, Quantity: outstanding, Outcome: MovementReturned,
				})
			}
		}
	} else if len(in.Items) == 0 && len(items) > 0 {
		return nil, ErrMovementInvalid
	}

	seen := make(map[int64]bool)
	for _, line := range in.Items {
		if seen[line.GatepassItemID] {
			return nil, ErrMovementInvalid
		}
		seen[line.GatepassItemID] = true
		outstanding, ok := items[line.GatepassItemID]
		if !ok {
			return nil, ErrReturnItemInvalid
		}
		if line.Quantity > outstanding {
			return nil, ErrReturnQuantityExceeded
		}
	}

	movementID, err := r.insertMovement(ctx, tx, tenantID, gatepassID, MovementCheckin, actorUserID, in)
	if err != nil {
		return nil, err
	}
	for _, line := range in.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO gatepass_movement_items
				(tenant_id, movement_id, gatepass_item_id, quantity, outcome, condition, notes, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, NOW())`,
			tenantID, movementID, line.GatepassItemID, line.Quantity,
			line.Outcome, line.Condition, line.Notes); err != nil {
			return nil, err
		}
	}

	remaining := float64(0)
	for id, outstanding := range items {
		remaining += outstanding
		// subtract this transaction's accounted quantity
		for _, line := range in.Items {
			if line.GatepassItemID == id {
				remaining -= line.Quantity
				break
			}
		}
	}

	next := StatusPartiallyReturn
	if remaining <= 0.000001 {
		next = StatusCompleted
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE gatepasses
		SET status = ?, checked_in_at = CASE WHEN ? = 'checked_in' THEN NOW() ELSE checked_in_at END,
		    checked_in_by = CASE WHEN ? = 'checked_in' THEN ? ELSE checked_in_by END,
		    updated_at = NOW()
		WHERE id = ? AND tenant_id = ?`,
		next, next, next, actorUserID, gatepassID, tenantID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.getGatepass(ctx, tenantID, gatepassID)
}

func (r *MovementRepository) lockGatepass(ctx context.Context, tx *sql.Tx, tenantID, id int64) (*Gatepass, error) {
	row := tx.QueryRowContext(ctx,
		"SELECT "+gpCols+" FROM gatepasses WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL FOR UPDATE",
		id, tenantID)
	return scanGatepass(row)
}

func (r *MovementRepository) lockGatepassItems(ctx context.Context, tx *sql.Tx, tenantID, gatepassID int64) (map[int64]float64, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, quantity
		FROM gatepass_items
		WHERE tenant_id = ? AND gatepass_id = ?
		  AND direction IN ('leaving','returning')
		ORDER BY id
		FOR UPDATE`, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64]float64)
	for rows.Next() {
		var id int64
		var qty float64
		if err := rows.Scan(&id, &qty); err != nil {
			return nil, err
		}
		out[id] = qty
	}
	return out, rows.Err()
}

func (r *MovementRepository) lockGatepassItemsWithQuantities(ctx context.Context, tx *sql.Tx, tenantID, gatepassID int64) (map[int64]float64, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT gi.id, gi.quantity - COALESCE(SUM(
			CASE WHEN gm.type = 'checkin' THEN gmi.quantity ELSE 0 END
		), 0) AS outstanding
		FROM gatepass_items gi
		LEFT JOIN gatepass_movement_items gmi ON gmi.gatepass_item_id = gi.id AND gmi.tenant_id = ?
		LEFT JOIN gatepass_movements gm ON gm.id = gmi.movement_id AND gm.tenant_id = ? AND gm.gatepass_id = ?
		WHERE gi.tenant_id = ? AND gi.gatepass_id = ?
		GROUP BY gi.id, gi.quantity
		ORDER BY gi.id
		FOR UPDATE`, tenantID, tenantID, gatepassID, tenantID, gatepassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64]float64)
	for rows.Next() {
		var id int64
		var outstanding float64
		if err := rows.Scan(&id, &outstanding); err != nil {
			return nil, err
		}
		if outstanding < 0 {
			return nil, ErrReturnQuantityExceeded
		}
		out[id] = outstanding
	}
	return out, rows.Err()
}

func (r *MovementRepository) insertMovement(ctx context.Context, tx *sql.Tx, tenantID, gatepassID int64, typ MovementType, actorUserID int64, in MovementInput) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO gatepass_movements
			(tenant_id, gatepass_id, type, actor_user_id, gate_name, notes, occurred_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		tenantID, gatepassID, typ, actorUserID, nullableString(in.GateName), in.Notes)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func (r *MovementRepository) getGatepass(ctx context.Context, tenantID, id int64) (*Gatepass, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+gpCols+" FROM gatepasses WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL",
		id, tenantID)
	return scanGatepass(row)
}

func scanGatepass(row interface{ Scan(dest ...any) error }) (*Gatepass, error) {
	var g Gatepass
	if err := row.Scan(
		&g.ID, &g.TenantID, &g.GatepassTypeID, &g.PassNumber, &g.DepartmentID,
		&g.RequesterType, &g.RequesterUserID, &g.RequesterVisitorID, &g.VisitID,
		&g.Purpose, &g.IsReturnable, &g.ExpectedReturnAt, &g.RequiresApproval, &g.WorkflowID,
		&g.Status, &g.QRToken, &g.CheckedOutAt, &g.CheckedOutBy, &g.CheckedInAt, &g.CheckedInBy,
		&g.CancelledAt, &g.CancelledBy, &g.CancelReason, &g.IssuedBy, &g.IssuedAt,
		&g.CreatedBy, &g.CreatedAt, &g.UpdatedAt, &g.DeletedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}
