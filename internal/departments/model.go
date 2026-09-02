package departments

type Department struct {
	ID     int64
	Name   string
	Code   string
	Active bool
}

type DTO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Active bool   `json:"active"`
}

func ToDTO(d *Department) DTO {
	return DTO{ID: d.ID, Name: d.Name, Code: d.Code, Active: d.Active}
}

type Input struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Active *bool  `json:"active"`
}
