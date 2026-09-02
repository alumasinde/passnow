package gatepasses

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"gatepass/internal/approvals"
	"time"

	"gatepass/internal/database"
	"gatepass/internal/httpx"
	"gatepass/internal/numbering"
)

var (
	ErrNotFound            = errors.New("gatepasses: not found")
	ErrInvalidTransition   = errors.New("gatepasses: this action is not valid for the gatepass's current status")
	ErrDuplicatePassNumber = errors.New("gatepasses: pass number collision, retry")
)

type Repository struct {
	db    *sql.DB
	items *ItemRepository
}

func NewRepository(db *sql.DB, items *ItemRepository) *Repository {
	return &Repository{db: db, items: items}
}

const gpCols = `
	id, gatepass_type_id, pass_number, department_id,
	requester_type, requester_user_id, requester_visitor_id, visit_id,
	purpose, is_returnable, expected_return_at, requires_approval, workflow_id,
	status, qr_token, checked_out_at, checked_out_by, checked_in_at, checked_in_by,
	cancelled_at, cancelled_by, cancel_reason, issued_by, issued_at,
	created_by, created_at, updated_at, deleted_at
`

func (r *Repository) scan(row interface{ Scan(dest ...any) error }) (*Gatepass, error) {
	var g Gatepass
	if err := row.Scan(
		&g.ID, &g.GatepassTypeID, &g.PassNumber, &g.DepartmentID,
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

func (r *Repository) ByID(ctx context.Context, id int64) (*Gatepass, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+gpCols+" FROM gatepasses WHERE id = ? AND deleted_at IS NULL LIMIT 1",
		id)
	return r.scan(row)
}

func (r *Repository) ByQRToken(ctx context.Context, token string) (*Gatepass, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+gpCols+" FROM gatepasses WHERE qr_token = ? AND deleted_at IS NULL LIMIT 1", token)
	return r.scan(row)
}

type ListFilter struct {
	Status *Status
}

func (r *Repository) List(ctx context.Context, f ListFilter, p httpx.Pagination) ([]Gatepass, int, error) {
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	if f.Status != nil {
		where += " AND status = ?"
		args = append(args, *f.Status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM gatepasses "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT "+gpCols+" FROM gatepasses "+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
		append(args, p.Limit, p.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Gatepass
	for rows.Next() {
		g, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *g)
	}
	return out, total, rows.Err()
}

// ApprovalStep mirrors a gatepass_approvals row.
type ApprovalStep struct {
	ID           int64
	GatepassID   int64
	StepOrder    int
	Label        string
	ApproverType string
	RoleID       *int64
	UserID       *int64
	Required     bool
	Status       string
	ActedBy      *int64
	Comments     *string
}

func (r *Repository) ApprovalSteps(ctx context.Context, gatepassID int64) ([]ApprovalStep, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, gatepass_id, step_order, label, approver_type, role_id, user_id, required, status, acted_by, comments
		FROM gatepass_approvals WHERE gatepass_id = ? ORDER BY step_order`, gatepassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ApprovalStep
	for rows.Next() {
		var s ApprovalStep
		if err := rows.Scan(&s.ID, &s.GatepassID, &s.StepOrder, &s.Label, &s.ApproverType, &s.RoleID, &s.UserID,
			&s.Required, &s.Status, &s.ActedBy, &s.Comments); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CreateInputResolved is what the service hands the repository after all
// validation/defaulting — the repository itself does no business-rule
// checking, only persistence.
type CreateInputResolved struct {
	Gatepass      *Gatepass
	Items         []ItemInput
	WorkflowSteps []WorkflowStepSnapshot // empty if no approval required
	NumberScope   string
	NumberPrefix  string
	NumberPeriod  string
}

// WorkflowStepSnapshot is copied from an approval_workflow_steps template
// row (or built directly) at creation time.
type WorkflowStepSnapshot struct {
	StepOrder    int
	Label        string
	ApproverType string
	RoleID       *int64
	UserID       *int64
	Required     bool
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Create persists the gatepass, its items, and (if any) its snapshotted
// approval steps in a single transaction. The pass number is generated
// INSIDE this same transaction via the numbering package (row-locked
// counter) so numbering and gatepass creation succeed or fail together —
// no risk of a burned number with no gatepass, or vice versa.
func (r *Repository) Create(ctx context.Context, in CreateInputResolved) (id int64, passNumber string, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback()

	token, err := randomToken()
	if err != nil {
		return 0, "", err
	}

	seq, err := numbering.Next(ctx, tx, in.NumberScope, in.NumberPeriod)
	if err != nil {
		return 0, "", err
	}
	passNumber = numbering.Format(in.NumberPrefix, in.NumberPeriod, seq)

	g := in.Gatepass
	status := StatusApproved
	var issuedBy any
	var issuedAt any
	workflowID := g.WorkflowID
	if g.RequiresApproval && len(in.WorkflowSteps) > 0 {
		status = StatusPendingApproval
	} else {
		// No approval is required: authorization happens at creation time.
		// Recording issued_by/issued_at makes the approval/expiry semantics
		// identical to a gatepass that completed an approval workflow.
		issuedBy = g.CreatedBy
		issuedAt = time.Now().UTC()
		workflowID = nil
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO gatepasses
			(gatepass_type_id, pass_number, department_id,
			 requester_type, requester_user_id, requester_visitor_id, visit_id,
			 purpose, is_returnable, expected_return_at, requires_approval, workflow_id,
			 status, qr_token, issued_by, issued_at, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		g.GatepassTypeID, passNumber, g.DepartmentID,
		g.RequesterType, g.RequesterUserID, g.RequesterVisitorID, g.VisitID,
		g.Purpose, g.IsReturnable, g.ExpectedReturnAt, g.RequiresApproval, workflowID,
		status, token, issuedBy, issuedAt, g.CreatedBy,
	)
	if err != nil {
		if database.IsDuplicateKeyErr(err) {
			return 0, "", ErrDuplicatePassNumber
		}
		return 0, "", err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, "", err
	}

	if err := r.items.CreateForGatepass(ctx, tx, id, in.Items); err != nil {
		return 0, "", err
	}

	for _, s := range in.WorkflowSteps {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO gatepass_approvals
				(gatepass_id, step_order, label, approver_type, role_id, user_id, required, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', NOW())`,
			id, s.StepOrder, s.Label, s.ApproverType, s.RoleID, s.UserID, s.Required,
		); err != nil {
			return 0, "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return id, passNumber, nil
}

// ActOnApprovalStep row-locks the gatepass, verifies the step is the
// current pending one (all earlier REQUIRED steps already approved),
// verifies the caller is eligible for THIS step (role or specific user
// match — checked by the caller/service using the returned step, since
// the caller already has the claims), applies approve/reject, and — if
// this was the last required step and approved — flips the gatepass to
// approved; if rejected, flips the gatepass to rejected and marks all
// other pending steps skipped. All inside one transaction so two
// approvers acting on the same step simultaneously can't both succeed.
func (r *Repository) ActOnApprovalStep(ctx context.Context, gatepassID, stepID, actorUserID int64, approve bool, comments string) (*Gatepass, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var gStatus Status
	if err := tx.QueryRowContext(ctx,
		`SELECT status FROM gatepasses WHERE id = ? AND deleted_at IS NULL FOR UPDATE`,
		gatepassID,
	).Scan(&gStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if gStatus != StatusPendingApproval {
		return nil, ErrInvalidTransition
	}

	var stepOrder int
	var stepStatus string
	if err := tx.QueryRowContext(ctx,
		`SELECT step_order, status FROM gatepass_approvals WHERE id = ? AND gatepass_id = ? FOR UPDATE`,
		stepID, gatepassID,
	).Scan(&stepOrder, &stepStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if stepStatus != "pending" {
		return nil, ErrInvalidTransition
	}

	// Re-check actor eligibility while the gatepass/step are locked. The
	// service performs an early check for fast feedback, but this second
	// check closes the race where a membership is deactivated/changed
	// between the service check and this transaction.
	var approverType string
	var roleID, userID sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
		SELECT approver_type, role_id, user_id
		FROM gatepass_approvals
		WHERE id = ? AND gatepass_id = ?
		FOR UPDATE`, stepID, gatepassID).Scan(&approverType, &roleID, &userID); err != nil {
		return nil, err
	}

	eligible := false
	switch approvals.ApproverType(approverType) {
	case approvals.ApproverRole:
		if roleID.Valid {
			var membershipCount int
			if err := tx.QueryRowContext(ctx, `
				SELECT COUNT(*)
				FROM tenant_memberships
				WHERE user_id = ? AND role_id = ? AND status = 'active'`,
				actorUserID, roleID.Int64).Scan(&membershipCount); err != nil {
				return nil, err
			}
			eligible = membershipCount == 1
		}
	case approvals.ApproverSpecificUser:
		if userID.Valid && userID.Int64 == actorUserID {
			var membershipCount int
			if err := tx.QueryRowContext(ctx, `
				SELECT COUNT(*)
				FROM tenant_memberships
				WHERE user_id = ? AND status = 'active'`,
				actorUserID).Scan(&membershipCount); err != nil {
				return nil, err
			}
			eligible = membershipCount == 1
		}
	}
	if !eligible {
		return nil, ErrNotEligibleApprover
	}

	// Enforce strict sequential order: every earlier REQUIRED step must
	// already be approved.
	var unmetEarlier int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM gatepass_approvals
		WHERE gatepass_id = ? AND step_order < ? AND required = 1 AND status != 'approved'`,
		gatepassID, stepOrder,
	).Scan(&unmetEarlier); err != nil {
		return nil, err
	}
	if unmetEarlier > 0 {
		return nil, ErrInvalidTransition
	}

	newStepStatus := "approved"
	if !approve {
		newStepStatus = "rejected"
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE gatepass_approvals SET status = ?, acted_by = ?, acted_at = NOW(), comments = ?
		WHERE id = ?`, newStepStatus, actorUserID, comments, stepID); err != nil {
		return nil, err
	}

	if !approve {
		if _, err := tx.ExecContext(ctx, `
			UPDATE gatepass_approvals SET status = 'skipped' WHERE gatepass_id = ? AND status = 'pending'`,
			gatepassID); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE gatepasses SET status = 'rejected', updated_at = NOW() WHERE id = ?`,
			gatepassID); err != nil {
			return nil, err
		}
	} else {
		var remainingRequired int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM gatepass_approvals
			WHERE gatepass_id = ? AND required = 1 AND status != 'approved'`,
			gatepassID,
		).Scan(&remainingRequired); err != nil {
			return nil, err
		}
		if remainingRequired == 0 {
			if _, err := tx.ExecContext(ctx, `
				UPDATE gatepasses SET status = 'approved', issued_by = ?, issued_at = NOW(), updated_at = NOW()
				WHERE id = ?`, actorUserID, gatepassID); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.ByID(ctx, gatepassID)
}

func (r *Repository) CheckOut(ctx context.Context, id, actorUserID int64, direction string) (*Gatepass, error) {
	return r.transitionGate(ctx, id, actorUserID, direction, true)
}

func (r *Repository) CheckIn(ctx context.Context, id, actorUserID int64, direction string) (*Gatepass, error) {
	return r.transitionGate(ctx, id, actorUserID, direction, false)
}

func (r *Repository) transitionGate(ctx context.Context, id, actorUserID int64, direction string, checkOut bool) (*Gatepass, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx,
		"SELECT "+gpCols+" FROM gatepasses WHERE id = ? AND deleted_at IS NULL FOR UPDATE",
		id)
	g, err := r.scan(row)
	if err != nil {
		return nil, err
	}

	dir := Direction(direction)
	var ok bool
	if checkOut {
		ok = g.CanCheckOut(dir)
	} else {
		ok = g.CanCheckIn(dir)
	}
	if !ok {
		return nil, ErrInvalidTransition
	}

	if checkOut {
		if _, err := tx.ExecContext(ctx, `
			UPDATE gatepasses SET status = 'checked_out', checked_out_at = NOW(), checked_out_by = ?, updated_at = NOW()
			WHERE id = ?`, actorUserID, id); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE gatepasses SET status = 'checked_in', checked_in_at = NOW(), checked_in_by = ?, updated_at = NOW()
			WHERE id = ?`, actorUserID, id); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.ByID(ctx, id)
}

// PendingApprovalItem is one gatepass currently waiting on THIS approver
// specifically — not just "pending approval" in general, but the exact
// current step, already filtered to steps this user is eligible to act on.
type PendingApprovalItem struct {
	GatepassID    int64
	PassNumber    string
	StepID        int64
	StepOrder     int
	StepLabel     string
	RequesterType string
	Purpose       *string
	CreatedAt     string
}

// PendingForApprover returns every gatepass currently sitting at a step
// this specific user (by role membership OR as a named specific_user) is
// eligible to approve/reject RIGHT NOW — i.e. it's their turn (all earlier
// required steps already approved). This is what turns "you need the
// gatepass ID and step ID" into an actual work queue: an approver opens
// one screen and sees exactly what needs their attention.
func (r *Repository) PendingForApprover(ctx context.Context, actorUserID, actorRoleID int64) ([]PendingApprovalItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.pass_number, ga.id, ga.step_order, ga.label, g.requester_type, g.purpose, g.created_at
		FROM gatepass_approvals ga
		JOIN gatepasses g ON g.id = ga.gatepass_id
		WHERE ga.status = 'pending'
		  AND g.status = 'pending_approval'
		  AND (
		        (ga.approver_type = 'role' AND ga.role_id = ?)
		     OR (ga.approver_type = 'specific_user' AND ga.user_id = ?)
		      )
		  AND NOT EXISTS (
		        SELECT 1 FROM gatepass_approvals earlier
		        WHERE earlier.gatepass_id = ga.gatepass_id
		          AND earlier.step_order < ga.step_order
		          AND earlier.required = 1
		          AND earlier.status != 'approved'
		      )
		ORDER BY g.created_at`, actorRoleID, actorUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PendingApprovalItem
	for rows.Next() {
		var it PendingApprovalItem
		if err := rows.Scan(&it.GatepassID, &it.PassNumber, &it.StepID, &it.StepOrder, &it.StepLabel,
			&it.RequesterType, &it.Purpose, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Repository) Cancel(ctx context.Context, id, actorUserID int64, reason string) (*Gatepass, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status Status
	if err := tx.QueryRowContext(ctx,
		`SELECT status FROM gatepasses WHERE id = ? AND deleted_at IS NULL FOR UPDATE`,
		id,
	).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if status != StatusPendingApproval && status != StatusApproved {
		return nil, ErrInvalidTransition
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE gatepasses SET status = 'cancelled', cancelled_at = NOW(), cancelled_by = ?, cancel_reason = ?, updated_at = NOW()
		WHERE id = ?`, actorUserID, reason, id); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.ByID(ctx, id)
}
