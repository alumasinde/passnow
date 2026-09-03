package gatepasses

import "time"

type DTO struct {
	ID                 int64             `json:"id"`
	GatepassTypeID     int64             `json:"gatepass_type_id"`
	AssignedGateID      *int64            `json:"assigned_gate_id,omitempty"`
	AssignedGateName    string            `json:"assigned_gate_name,omitempty"`
	PassNumber         string            `json:"pass_number"`
	GatepassTypeName   string            `json:"gatepass_type_name,omitempty"`
	Direction          string            `json:"direction,omitempty"`
	DepartmentName     string            `json:"department_name,omitempty"`
	RequesterName      string            `json:"requester_name,omitempty"`
	SubjectName        string            `json:"subject_name,omitempty"`
	DepartmentID       *int64            `json:"department_id"`
	RequesterType      string            `json:"requester_type"`
	RequesterUserID    *int64            `json:"requester_user_id,omitempty"`
	RequesterVisitorID *int64            `json:"requester_visitor_id,omitempty"`
	VisitID            *int64            `json:"visit_id"`
	Purpose            *string           `json:"purpose"`
	IsReturnable       bool              `json:"is_returnable"`
	ExpectedReturnAt   *time.Time        `json:"expected_return_at"`
	RequiresApproval   bool              `json:"requires_approval"`
	WorkflowID         *int64            `json:"workflow_id,omitempty"`
	Status             string            `json:"status"`
	QRToken            string            `json:"qr_token,omitempty"`
	CheckedOutAt       *time.Time        `json:"checked_out_at"`
	CheckedInAt        *time.Time        `json:"checked_in_at"`
	CancelReason       *string           `json:"cancel_reason,omitempty"`
	Items              []ItemDTO         `json:"items,omitempty"`
	Approvals          []ApprovalStepDTO `json:"approvals,omitempty"`
	Movements          []MovementDTO     `json:"movements,omitempty"`
}

func ToDTO(g *Gatepass) DTO {
	return DTO{
		ID: g.ID, GatepassTypeID: g.GatepassTypeID, AssignedGateID: g.AssignedGateID, PassNumber: g.PassNumber,
		DepartmentID: g.DepartmentID, RequesterType: string(g.RequesterType),
		RequesterUserID: g.RequesterUserID, RequesterVisitorID: g.RequesterVisitorID,
		VisitID: g.VisitID, Purpose: g.Purpose, IsReturnable: g.IsReturnable,
		ExpectedReturnAt: g.ExpectedReturnAt, RequiresApproval: g.RequiresApproval,
		WorkflowID: g.WorkflowID, Status: string(g.Status), CheckedOutAt: g.CheckedOutAt, CheckedInAt: g.CheckedInAt,
		CancelReason: g.CancelReason, QRToken: g.QRToken,
	}
}

// CreateInput is the explicit allow-list for creating a gatepass.
//
// RequestsApproval is the requester's "needs approval" checkbox — it is
// only ever an OPT-IN. If the resolved GatepassType mandates approval
// (RequiresApproval=true), the workflow runs regardless of this flag; the
// requester can never uncheck their way out of a mandatory approval. See
// service.go.
type CreateInput struct {
	GatepassTypeID     int64       `json:"gatepass_type_id"`
	AssignedGateID      *int64      `json:"assigned_gate_id"`
	DepartmentID       *int64      `json:"department_id"`
	RequesterType      string      `json:"requester_type"`       // "employee" | "visitor"
	RequesterVisitorID *int64      `json:"requester_visitor_id"` // required if requester_type=visitor
	VisitID            *int64      `json:"visit_id"`
	Purpose            *string     `json:"purpose"`
	IsReturnable       *bool       `json:"is_returnable"` // constrained by the gatepass type returnability policy
	ExpectedReturnAt   *time.Time  `json:"expected_return_at"`
	RequestsApproval   bool        `json:"needs_approval"`
	Items              []ItemInput `json:"items"`
}

type CancelInput struct {
	Reason string `json:"reason"`
}

type ApprovalActionInput struct {
	Comments string `json:"comments"`
}

type ApprovalStepDTO struct {
	StepOrder int     `json:"step_order"`
	Label     string  `json:"label"`
	Status    string  `json:"status"`
	ActedBy   *int64  `json:"acted_by,omitempty"`
	Comments  *string `json:"comments,omitempty"`
}

// --- Items ---------------------------------------------------------------

type ItemDTO struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Category     *string `json:"category"`
	Quantity     float64 `json:"quantity"`
	Unit         *string `json:"unit"`
	SerialNumber *string `json:"serial_number"`
	AssetNumber  *string `json:"asset_number"`
	Condition    *string `json:"condition"`
	Direction    string  `json:"direction"`
}

type ItemInput struct {
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Category     *string `json:"category"`
	Quantity     float64 `json:"quantity"`
	Unit         *string `json:"unit"`
	SerialNumber *string `json:"serial_number"`
	AssetNumber  *string `json:"asset_number"`
	Condition    *string `json:"condition"`
	Direction    string  `json:"direction"` // "entering" | "leaving" | "returning"
}

// QRDTO is what a QR scan/lookup returns — deliberately minimal, no
// contact info or ID numbers.
type QRDTO struct {
	GatepassID   int64   `json:"gatepass_id"`
	PassNumber   string  `json:"pass_number"`
	Status       string  `json:"status"`
	Requester    string  `json:"requester"`
	Purpose      *string `json:"purpose"`
	IsReturnable bool    `json:"is_returnable"`
}
