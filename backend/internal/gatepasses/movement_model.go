package gatepasses

import "time"

type MovementType string

const (
	MovementCheckout MovementType = "checkout"
	MovementCheckin  MovementType = "checkin"
)

type MovementOutcome string

const (
	MovementReleased MovementOutcome = "released"
	MovementReturned MovementOutcome = "returned"
	MovementDamaged  MovementOutcome = "damaged"
	MovementLost     MovementOutcome = "lost"
)

type Movement struct {
	ID          int64
	TenantID    int64
	GatepassID  int64
	Type        MovementType
	ActorUserID int64
	GateName    *string
	Notes       *string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

type MovementItem struct {
	ID             int64
	TenantID       int64
	MovementID     int64
	GatepassItemID int64
	Quantity       float64
	Outcome        MovementOutcome
	Condition      *string
	Notes          *string
	CreatedAt      time.Time
}

type MovementItemInput struct {
	GatepassItemID int64           `json:"gatepass_item_id"`
	Quantity       float64         `json:"quantity"`
	Outcome        MovementOutcome `json:"outcome"`
	Condition      *string         `json:"condition"`
	Notes          *string         `json:"notes"`
}

type MovementInput struct {
	GateName string              `json:"gate_name"`
	Notes    *string             `json:"notes"`
	Items    []MovementItemInput `json:"items"`
	// FullReturn closes all outstanding returnable quantities. If false,
	// Items must describe the quantities being returned.
	FullReturn bool `json:"full_return"`
}

type MovementDTO struct {
	ID          int64             `json:"id"`
	Type        string            `json:"type"`
	ActorUserID int64             `json:"actor_user_id"`
	GateName    *string           `json:"gate_name,omitempty"`
	Notes       *string           `json:"notes,omitempty"`
	OccurredAt  time.Time         `json:"occurred_at"`
	Items       []MovementItemDTO `json:"items,omitempty"`
}

type MovementItemDTO struct {
	ID             int64   `json:"id"`
	GatepassItemID int64   `json:"gatepass_item_id"`
	Quantity       float64 `json:"quantity"`
	Outcome        string  `json:"outcome"`
	Condition      *string `json:"condition,omitempty"`
	Notes          *string `json:"notes,omitempty"`
}
