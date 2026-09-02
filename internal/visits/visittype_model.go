package visits

// VisitType is a tenant-configurable visit purpose/category.
type VisitType struct {
	ID     int64
	Name   string
	Code   string
	Active bool
}

type VisitTypeDTO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Active bool   `json:"active"`
}

func VisitTypeToDTO(t *VisitType) VisitTypeDTO {
	return VisitTypeDTO{ID: t.ID, Name: t.Name, Code: t.Code, Active: t.Active}
}

type VisitTypeInput struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Active *bool  `json:"active"`
}
