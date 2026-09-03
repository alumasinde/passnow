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

type Repository struct { db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func validateSteps(in CreateWorkflowInput) error {
	if len(in.Steps) == 0 { return ErrInvalidStepConfig }
	hasRequired := false
	for _, s := range in.Steps {
		if s.Label == "" { return ErrInvalidStepConfig }
		required := s.Required == nil || *s.Required
		if required { hasRequired = true }
		switch ApproverType(s.ApproverType) {
		case ApproverRole:
			if s.RoleID == nil || s.UserID != nil || *s.RoleID <= 0 { return ErrInvalidStepConfig }
		case ApproverSpecificUser:
			if s.UserID == nil || s.RoleID != nil || *s.UserID <= 0 { return ErrInvalidStepConfig }
		default:
			return ErrInvalidStepConfig
		}
	}
	if !hasRequired { return ErrWorkflowNeedsRequiredStep }
	return nil
}

func validateApprovers(ctx context.Context, tx *sql.Tx, in CreateWorkflowInput) error {
	for _, s := range in.Steps {
		switch ApproverType(s.ApproverType) {
		case ApproverRole:
			var count int
			if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM roles WHERE id = ?`, *s.RoleID).Scan(&count); err != nil || count != 1 {
				return ErrApproverNotInTenant
			}
		case ApproverSpecificUser:
			var status string
			if err := tx.QueryRowContext(ctx, `SELECT status FROM tenant_memberships WHERE user_id = ? LIMIT 1`, *s.UserID).Scan(&status); err != nil || status != "active" {
				return ErrApproverNotInTenant
			}
		}
	}
	return nil
}

func (r *Repository) CreateWithSteps(ctx context.Context, tenantID int64, in CreateWorkflowInput) (int64, error) {
	_ = tenantID
	if err := validateSteps(in); err != nil { return 0, err }
	tx, err := r.db.BeginTx(ctx, nil); if err != nil { return 0, err }
	defer tx.Rollback()
	if err := validateApprovers(ctx, tx, in); err != nil { return 0, err }
	active := true
	if in.Active != nil { active = *in.Active }
	res, err := tx.ExecContext(ctx, `
		INSERT INTO approval_workflows (name, active, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())`, in.Name, active)
	if err != nil { return 0, err }
	id, err := res.LastInsertId(); if err != nil { return 0, err }
	if err := insertSteps(ctx, tx, id, in.Steps); err != nil { return 0, err }
	if err := tx.Commit(); err != nil { return 0, err }
	return id, nil
}

func insertSteps(ctx context.Context, tx *sql.Tx, workflowID int64, steps []StepInput) error {
	for i, s := range steps {
		required := s.Required == nil || *s.Required
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO approval_workflow_steps
				(workflow_id, step_order, label, approver_type, role_id, user_id, required)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			workflowID, i+1, s.Label, s.ApproverType, s.RoleID, s.UserID, required,
		); err != nil { return err }
	}
	return nil
}

// UpdateWithSteps replaces the template atomically. Existing gatepass approvals
// are unaffected because they were snapshotted when each gatepass was created.
func (r *Repository) UpdateWithSteps(ctx context.Context, tenantID, id int64, in UpdateWorkflowInput) error {
	_ = tenantID
	if err := validateSteps(in); err != nil { return err }
	tx, err := r.db.BeginTx(ctx, nil); if err != nil { return err }
	defer tx.Rollback()
	if err := validateApprovers(ctx, tx, in); err != nil { return err }
	var currentActive bool
	if err := tx.QueryRowContext(ctx, `SELECT active FROM approval_workflows WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, id).Scan(&currentActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return ErrNotFound }
		return err
	}
	active := currentActive
	if in.Active != nil { active = *in.Active }
	if _, err := tx.ExecContext(ctx, `UPDATE approval_workflows SET name = ?, active = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL`, in.Name, active, id); err != nil { return err }
	if _, err := tx.ExecContext(ctx, `DELETE FROM approval_workflow_steps WHERE workflow_id = ?`, id); err != nil { return err }
	if err := insertSteps(ctx, tx, id, in.Steps); err != nil { return err }
	return tx.Commit()
}

func (r *Repository) ByID(ctx context.Context, tenantID, id int64) (*Workflow, []Step, error) {
	var w Workflow
	err := r.db.QueryRowContext(ctx, `
		SELECT w.id, w.name, w.active,
		       (SELECT COUNT(*) FROM approval_workflow_steps s WHERE s.workflow_id = w.id)
		FROM approval_workflows w WHERE w.id = ? AND w.deleted_at IS NULL LIMIT 1`, id,
	).Scan(&w.ID, &w.Name, &w.Active, &w.StepCount)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, nil, ErrNotFound }; return nil, nil, err }
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workflow_id, step_order, label, approver_type, role_id, user_id, required
		FROM approval_workflow_steps WHERE workflow_id = ? ORDER BY step_order`, id)
	if err != nil { return nil, nil, err }
	defer rows.Close()
	var steps []Step
	for rows.Next() {
		var s Step
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.StepOrder, &s.Label, &s.ApproverType, &s.RoleID, &s.UserID, &s.Required); err != nil { return nil, nil, err }
		steps = append(steps, s)
	}
	return &w, steps, rows.Err()
}

func (r *Repository) List(ctx context.Context, tenantID int64, activeOnly bool) ([]Workflow, error) {
	_ = tenantID
	q := `SELECT w.id, w.name, w.active, COUNT(s.id)
		FROM approval_workflows w LEFT JOIN approval_workflow_steps s ON s.workflow_id = w.id
		WHERE w.deleted_at IS NULL`
	if activeOnly { q += " AND w.active = 1" }
	q += " GROUP BY w.id, w.name, w.active ORDER BY w.name"
	rows, err := r.db.QueryContext(ctx, q); if err != nil { return nil, err }
	defer rows.Close()
	var out []Workflow
	for rows.Next() {
		var w Workflow
		if err := rows.Scan(&w.ID, &w.Name, &w.Active, &w.StepCount); err != nil { return nil, err }
		out = append(out, w)
	}
	return out, rows.Err()
}
