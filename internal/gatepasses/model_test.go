package gatepasses

import "testing"

func TestGatepassStateMachine(t *testing.T) {
	cases := []struct {
		from Status
		to   Status
		ok   bool
	}{
		{StatusDraft, StatusSubmitted, true},
		{StatusSubmitted, StatusPendingApproval, true},
		{StatusPendingApproval, StatusApproved, true},
		{StatusPendingApproval, StatusRejected, true},
		{StatusApproved, StatusCheckedOut, true},
		{StatusApproved, StatusExpired, true},
		{StatusCheckedOut, StatusAwaitingReturn, true},
		{StatusAwaitingReturn, StatusPartiallyReturn, true},
		{StatusPartiallyReturn, StatusCompleted, false},
		{StatusRejected, StatusApproved, false},
		{StatusCompleted, StatusCheckedOut, false},
		{StatusCancelled, StatusApproved, false},
	}
	for _, tc := range cases {
		if got := tc.from.CanTransitionTo(tc.to); got != tc.ok {
			t.Fatalf("%s -> %s: got %v want %v", tc.from, tc.to, got, tc.ok)
		}
	}
}

func TestNextMovementStatuses(t *testing.T) {
	g := &Gatepass{IsReturnable: true}
	if got := g.NextCheckOutStatus(); got != StatusAwaitingReturn {
		t.Fatalf("returnable checkout: got %s", got)
	}
	if got := g.NextCheckInStatus(DirectionOut); got != StatusCompleted {
		t.Fatalf("returnable check-in: got %s", got)
	}
}
