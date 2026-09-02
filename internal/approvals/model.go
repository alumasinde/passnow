package approvals

type ApproverType string

const (
	ApproverRole         ApproverType = "role"
	ApproverSpecificUser ApproverType = "specific_user"
)

type Workflow struct {
	ID       int64
	Name     string
	Active   bool
}

// Step is a template step (e.g. "Level 1: HOD"). Snapshotted onto each
// gatepass at creation time as a gatepass_approvals row — see the
// gatepasses package for why.
type Step struct {
	ID           int64
	WorkflowID   int64
	StepOrder    int
	Label        string
	ApproverType ApproverType
	RoleID       *int64
	UserID       *int64
	Required     bool
}

type WorkflowDTO struct {
	ID     int64     `json:"id"`
	Name   string    `json:"name"`
	Active bool      `json:"active"`
	Steps  []StepDTO `json:"steps"`
}

type StepDTO struct {
	ID           int64  `json:"id"`
	StepOrder    int    `json:"step_order"`
	Label        string `json:"label"`
	ApproverType string `json:"approver_type"`
	RoleID       *int64 `json:"role_id"`
	UserID       *int64 `json:"user_id"`
	Required     bool   `json:"required"`
}

func StepToDTO(s *Step) StepDTO {
	return StepDTO{
		ID: s.ID, StepOrder: s.StepOrder, Label: s.Label,
		ApproverType: string(s.ApproverType), RoleID: s.RoleID, UserID: s.UserID, Required: s.Required,
	}
}

// StepInput is the allow-list for defining one workflow step at creation
// time. Exactly one of RoleID/UserID must be set, matching ApproverType —
// validated in the service, not left to the DB CHECK constraint alone.
type StepInput struct {
	Label        string `json:"label"`
	ApproverType string `json:"approver_type"`
	RoleID       *int64 `json:"role_id"`
	UserID       *int64 `json:"user_id"`
	Required     *bool  `json:"required"`
}

type CreateWorkflowInput struct {
	Name  string      `json:"name"`
	Steps []StepInput `json:"steps"`
}
