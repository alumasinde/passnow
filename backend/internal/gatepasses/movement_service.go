package gatepasses

import (
	"errors"
	"fmt"
)

var (
	ErrMovementInvalid        = errors.New("gatepasses: invalid movement")
	ErrMovementNotAllowed     = errors.New("gatepasses: movement is not allowed for the current gatepass")
	ErrReturnQuantityExceeded = errors.New("gatepasses: return quantity exceeds outstanding quantity")
	ErrReturnItemInvalid      = errors.New("gatepasses: return item does not belong to this gatepass")
	ErrFullReturnWithItems    = errors.New("gatepasses: full_return cannot be combined with item quantities")
)

func validateMovementInput(in MovementInput, checkIn bool) error {
	if len(in.GateName) > 120 {
		return fmt.Errorf("%w: gate_name is too long", ErrMovementInvalid)
	}
	for _, item := range in.Items {
		if item.GatepassItemID <= 0 || item.Quantity <= 0 {
			return ErrMovementInvalid
		}
		if checkIn {
			switch item.Outcome {
			case MovementReturned, MovementDamaged, MovementLost:
			default:
				return ErrMovementInvalid
			}
		}
	}
	if checkIn && in.FullReturn && len(in.Items) > 0 {
		return ErrFullReturnWithItems
	}
	return nil
}
