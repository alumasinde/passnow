package gatepasses

import "time"

type RequesterType string

const (
	RequesterEmployee RequesterType = "employee"
	RequesterVisitor  RequesterType = "visitor"
)

type Status string

type Gatepass struct {
	ID             int64
	GatepassTypeID int64
	AssignedGateID *int64
	PassNumber     string

	DepartmentID *int64

	RequesterType      RequesterType
	RequesterUserID    *int64
	RequesterVisitorID *int64

	VisitID *int64

	Purpose *string

	IsReturnable     bool
	ExpectedReturnAt *time.Time

	RequiresApproval bool
	WorkflowID       *int64

	Status Status

	QRToken string

	CheckedOutAt *time.Time
	CheckedOutBy *int64
	CheckedInAt  *time.Time
	CheckedInBy  *int64

	CancelledAt  *time.Time
	CancelledBy  *int64
	CancelReason *string

	IssuedBy *int64
	IssuedAt *time.Time

	CreatedBy *int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// --- State machine ---------------------------------------------------
//
// Authorization and physical movement are deliberately separate.
// APPROVED means the requested movement has been authorized.
// CHECKED_OUT means a security officer confirmed physical departure.
// A returnable pass then enters AWAITING_RETURN until security confirms
// the return. COMPLETED is the business terminal state.
//
// Actual state changes are transactionally persisted by the repository.

const (
	StatusDraft           Status = "draft"
	StatusSubmitted       Status = "submitted"
	StatusPendingApproval Status = "pending_approval"
	StatusApproved        Status = "approved"
	StatusRejected        Status = "rejected"
	StatusCancelled       Status = "cancelled"
	StatusExpired         Status = "expired"
	StatusCheckedOut      Status = "checked_out"
	StatusAwaitingReturn  Status = "awaiting_return"
	StatusPartiallyReturn Status = "partially_returned"
	StatusReturnOverdue   Status = "return_overdue"
	StatusCheckedIn       Status = "checked_in"
	StatusCompleted       Status = "completed"
)

func (s Status) CanTransitionTo(next Status) bool {
	switch s {
	case StatusDraft:
		return next == StatusSubmitted || next == StatusCancelled
	case StatusSubmitted:
		return next == StatusPendingApproval || next == StatusCancelled
	case StatusPendingApproval:
		return next == StatusApproved || next == StatusRejected || next == StatusCancelled
	case StatusApproved:
		return next == StatusCheckedOut || next == StatusCheckedIn || next == StatusExpired || next == StatusCancelled
	case StatusCheckedOut:
		return next == StatusAwaitingReturn || next == StatusCompleted
	case StatusAwaitingReturn:
		return next == StatusPartiallyReturn || next == StatusCheckedIn || next == StatusReturnOverdue
	case StatusPartiallyReturn:
		return next == StatusPartiallyReturn || next == StatusCheckedIn || next == StatusReturnOverdue
	case StatusReturnOverdue:
		return next == StatusPartiallyReturn || next == StatusCheckedIn
	case StatusCheckedIn:
		return next == StatusCompleted
	default:
		return false
	}
}

func (g *Gatepass) CanCheckOut(direction Direction) bool {
	switch g.Status {
	case StatusApproved:
		return direction == DirectionOut || direction == DirectionBoth
	default:
		return false
	}
}

func (g *Gatepass) CanCheckIn(direction Direction) bool {
	switch g.Status {
	case StatusApproved:
		return direction == DirectionIn || direction == DirectionBoth
	case StatusAwaitingReturn, StatusPartiallyReturn, StatusReturnOverdue:
		return g.IsReturnable && (direction == DirectionOut || direction == DirectionBoth)
	case StatusCheckedOut:
		return g.IsReturnable && (direction == DirectionOut || direction == DirectionBoth)
	default:
		return false
	}
}

func (g *Gatepass) NextCheckOutStatus() Status {
	if g.IsReturnable {
		return StatusAwaitingReturn
	}
	return StatusCompleted
}

func (g *Gatepass) NextCheckInStatus(direction Direction) Status {
	// For a returnable outbound/bidirectional pass, check-in is the physical
	// return and therefore completes the lifecycle. For an inbound pass,
	// check-in is the physical arrival and also completes the lifecycle.
	// StatusCheckedIn remains useful as a legacy/intermediate state but is
	// no longer produced by the movement engine.
	return StatusCompleted
}

func (g *Gatepass) CanCancel() bool {
	return g.Status == StatusDraft ||
		g.Status == StatusSubmitted ||
		g.Status == StatusPendingApproval ||
		g.Status == StatusApproved
}
