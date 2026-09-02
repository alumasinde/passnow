package gatepasses

type Direction string

const (
	DirectionIn   Direction = "in"
	DirectionOut  Direction = "out"
	DirectionBoth Direction = "both"
)

type ReturnabilityPolicy string

const (
	ReturnabilityOptional   ReturnabilityPolicy = "optional"
	ReturnabilityRequired   ReturnabilityPolicy = "required"
	ReturnabilityNotAllowed ReturnabilityPolicy = "not_allowed"
)

// GatepassType is tenant-local configuration.
type GatepassType struct {
	ID                  int64
	Name                string
	Code                string
	Description         *string
	Direction           Direction
	IsReturnableDefault bool
	ReturnabilityPolicy ReturnabilityPolicy
	RequiresItems       bool
	RequiresApproval    bool
	WorkflowID          *int64
	Active              bool
}

type TypeDTO struct {
	ID                  int64   `json:"id"`
	Name                string  `json:"name"`
	Code                string  `json:"code"`
	Description         *string `json:"description,omitempty"`
	Direction           string  `json:"direction"`
	IsReturnableDefault bool    `json:"is_returnable_default"`
	ReturnabilityPolicy string  `json:"returnability_policy"`
	RequiresItems       bool    `json:"requires_items"`
	RequiresApproval    bool    `json:"requires_approval"`
	WorkflowID          *int64  `json:"workflow_id,omitempty"`
	Active              bool    `json:"active"`
}

func TypeToDTO(t *GatepassType) TypeDTO {
	return TypeDTO{
		ID:t.ID, Name:t.Name, Code:t.Code, Description:t.Description, Direction:string(t.Direction),
		IsReturnableDefault:t.IsReturnableDefault, ReturnabilityPolicy:string(t.ReturnabilityPolicy),
		RequiresItems:t.RequiresItems, RequiresApproval:t.RequiresApproval, WorkflowID:t.WorkflowID, Active:t.Active,
	}
}

type TypeInput struct {
	Name                string  `json:"name"`
	Code                string  `json:"code"`
	Description         *string `json:"description"`
	Direction           string  `json:"direction"`
	IsReturnableDefault *bool   `json:"is_returnable_default"`
	ReturnabilityPolicy *string `json:"returnability_policy"`
	RequiresItems       *bool   `json:"requires_items"`
	RequiresApproval    *bool   `json:"requires_approval"`
	WorkflowID          *int64  `json:"workflow_id"`
	Active              *bool   `json:"active"`
}
