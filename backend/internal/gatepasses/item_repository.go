package gatepasses

import (
	"context"
	"database/sql"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// CreateForGatepass inserts items as part of the SAME transaction as
// gatepass creation (see Repository.Create) — either all items and the
// gatepass are saved, or none are.
func (r *ItemRepository) CreateForGatepass(ctx context.Context, tx *sql.Tx, gatepassID int64, items []ItemInput) error {
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO gatepass_items
				(gatepass_id, name, description, category, quantity, unit,
				 serial_number, asset_number, item_condition, direction, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			gatepassID, it.Name, it.Description, it.Category, it.Quantity, it.Unit,
			it.SerialNumber, it.AssetNumber, it.Condition, it.Direction,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *ItemRepository) ListForGatepass(ctx context.Context, gatepassID int64) ([]ItemDTO, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, category, quantity, unit, serial_number, asset_number, item_condition, direction
		FROM gatepass_items WHERE gatepass_id = ? ORDER BY id`, gatepassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ItemDTO
	for rows.Next() {
		var it ItemDTO
		if err := rows.Scan(&it.ID, &it.Name, &it.Description, &it.Category, &it.Quantity, &it.Unit,
			&it.SerialNumber, &it.AssetNumber, &it.Condition, &it.Direction); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
