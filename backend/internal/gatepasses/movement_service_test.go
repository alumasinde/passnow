package gatepasses

import "testing"

func TestValidateMovementInput(t *testing.T) {
	if err := validateMovementInput(MovementInput{
		Items: []MovementItemInput{{GatepassItemID: 1, Quantity: 1, Outcome: MovementReturned}},
	}, true); err != nil {
		t.Fatalf("valid check-in rejected: %v", err)
	}
	if err := validateMovementInput(MovementInput{
		FullReturn: true,
		Items:      []MovementItemInput{{GatepassItemID: 1, Quantity: 1, Outcome: MovementReturned}},
	}, true); err == nil {
		t.Fatal("full_return with explicit items must be rejected")
	}
	if err := validateMovementInput(MovementInput{
		Items: []MovementItemInput{{GatepassItemID: 1, Quantity: 0, Outcome: MovementReturned}},
	}, true); err == nil {
		t.Fatal("zero quantity must be rejected")
	}
}

func TestMovementOutcomeValues(t *testing.T) {
	for _, v := range []MovementOutcome{MovementReleased, MovementReturned, MovementDamaged, MovementLost} {
		if v == "" {
			t.Fatal("movement outcome must not be empty")
		}
	}
}
