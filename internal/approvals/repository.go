package approvals

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound                  = errors.New("approvals: workflow not found")
	ErrInvalidStepConfig         = errors.New("approvals: each step needs a label, a valid approver_type, and exactly one of role_id/user_id")
	ErrApproverNotInTenant       = errors.New("approvals: approver must belong to the same tenant")
	ErrWorkflowNeedsRequiredStep = errors.New("approvals: workflow must contain at least one required step")
	ErrNotEligibleApprover       = errors.New("approvals: actor is not eligible for this approval step")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateWithSteps creates a workflow and its ordered steps atomically. All
// steps are validated (exactly one of role_id/user_id, matching
// approver_type) before anything is written.
func (r *Repository) CreateWithSteps(ctx context.Context, tenantID int64, in CreateWorkflowInput) (int64, error) {
	if len(in.Steps) == 0 {
		return 0, ErrInvalidStepConfig
	}

	hasRequired := false
	for _, s := range in.Steps {
		if s.Label == "" {
			return 0, ErrInvalidStepConfig
		}
		required := true
		if s.Required != nil {
			required = *s.Required
		}
		if required {
			hasRequired = true
		}
		switch ApproverType(s.ApproverType) {
		case ApproverRole:
			if s.RoleID == nil || s.UserID != nil || *s.RoleID <= 0 {
				return 0, ErrInvalidStepConfig
			}
		case ApproverSpecificUser:
			if s.UserID == nil || s.RoleID != nil || *s.UserID <= 0 {
				return 0, ErrInvalidStepConfig
			}
		default:
			return 0, ErrInvalidStepConfig
		}
	}
	if !hasRequired {
		return 0, ErrWorkflowNeedsRequiredStep
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Validate every approver inside the same transaction used to create the
	// workflow. This prevents a tenant from referencing a role or user from
	// another tenant and avoids TOCTOU gaps between validation and INSERT.
	for _, s := range in.Steps {
		switch ApproverType(s.ApproverType) {
		case ApproverRole:
			var roleTenantID int64
			if err := tx.QueryRowContext(ctx,
				`SELECT tenant_id FROM roles WHERE id = ? LIMIT 1`, *s.RoleID).Scan(&roleTenantID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return 0, ErrApproverNotInTenant
				}
				return 0, err
			}
			if roleTenantID != tenantID {
				return 0, ErrApproverNotInTenant
			}
		case ApproverSpecificUser:
			var membershipTenantID int64
			var membershipStatus string
			if err := tx.QueryRowContext(ctx, `
				SELECT tenant_id, status
				FROM tenant_memberships
				WHERE tenant_id = ? AND user_id = ?
				LIMIT 1`, tenantID, *s.UserID).Scan(&membershipTenantID, &membershipStatus); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return 0, ErrApproverNotInTenant
				}
				return 0, err
			}
			if membershipTenantID != tenantID || membershipStatus != "active" {
				return 0, ErrApproverNotInTenant
			}
		}
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO approval_workflows (tenant_id, name, active, created_at, updated_at)
		VALUES (?, ?, 1, NOW(), NOW())`, tenantID, in.Name)
	if err != nil {
		return 0, err
	}
	workflowID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for i, s := range in.Steps {
		required := true
		if s.Required != nil {
			required = *s.Required
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO approval_workflow_steps
				(workflow_id, tenant_id, step_order, label, approver_type, role_id, user_id, required)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			workflowID, tenantID, i+1, s.Label, s.ApproverType, s.RoleID, s.UserID, required,
		); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return workflowID, nil
}

func (r *Repository) ByID(ctx context.Context, tenantID, id int64) (*Workflow, []Step, error) {
	var w Workflow
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, active FROM approval_workflows
		WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL LIMIT 1`, id, tenantID,
	).Scan(&w.ID, &w.TenantID, &w.Name, &w.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workflow_id, tenant_id, step_order, label, approver_type, role_id, user_id, required
		FROM approval_workflow_steps WHERE workflow_id = ? AND tenant_id = ? ORDER BY step_order`, id, tenantID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var steps []Step
	for rows.Next() {
		var s Step
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.TenantID, &s.StepOrder, &s.Label, &s.ApproverType, &s.RoleID, &s.UserID, &s.Required); err != nil {
			return nil, nil, err
		}
		steps = append(steps, s)
	}
	return &w, steps, rows.Err()
}

func (r *Repository) List(ctx context.Context, tenantID int64, activeOnly bool) ([]Workflow, error) {
	q := `SELECT id, tenant_id, name, active FROM approval_workflows WHERE tenant_id = ? AND deleted_at IS NULL`
	if activeOnly {
		q += " AND active = 1"
	}
	q += " ORDER BY name"

	rows, err := r.db.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Workflow
	for rows.Next() {
		var w Workflow
		if err := rows.Scan(&w.ID, &w.TenantID, &w.Name, &w.Active); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
