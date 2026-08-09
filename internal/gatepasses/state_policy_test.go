package gatepasses

import "testing"

func TestReturnabilityPolicy(t *testing.T) {
	required := true
	falseValue := false

	got, err := ResolveReturnability(ReturnabilityRequired, &required, false)
	if err != nil || !got {
		t.Fatalf("required policy: got=%v err=%v", got, err)
	}

	if _, err := ResolveReturnability(ReturnabilityRequired, &falseValue, true); err == nil {
		t.Fatal("required policy must reject disabling return")
	}

	got, err = ResolveReturnability(ReturnabilityNotAllowed, &falseValue, true)
	if err != nil || got {
		t.Fatalf("not_allowed policy: got=%v err=%v", got, err)
	}

	if _, err := ResolveReturnability(ReturnabilityNotAllowed, &required, false); err == nil {
		t.Fatal("not_allowed policy must reject enabling return")
	}
}

func TestGatepassTransitions(t *testing.T) {
	valid := [][2]Status{
		{StatusDraft, StatusSubmitted},
		{StatusSubmitted, StatusPendingApproval},
		{StatusPendingApproval, StatusApproved},
		{StatusApproved, StatusCheckedOut},
		{StatusCheckedOut, StatusAwaitingReturn},
		{StatusAwaitingReturn, StatusPartiallyReturn},
		{StatusPartiallyReturn, StatusCheckedIn},
		{StatusCheckedIn, StatusCompleted},
	}
	for _, tt := range valid {
		if !tt[0].CanTransitionTo(tt[1]) {
			t.Errorf("expected transition %s -> %s", tt[0], tt[1])
		}
	}
	if StatusRejected.CanTransitionTo(StatusCheckedOut) {
		t.Fatal("rejected gatepass must not be checkable out")
	}
	if StatusExpired.CanTransitionTo(StatusCheckedOut) {
		t.Fatal("expired gatepass must not be checkable out")
	}
}
