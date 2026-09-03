package gates

import "time"

type Gate struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Location    *string   `json:"location,omitempty"`
	AllowsEntry bool      `json:"allows_entry"`
	AllowsExit  bool      `json:"allows_exit"`
	IsDefault   bool      `json:"is_default"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Input struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Location    *string `json:"location"`
	AllowsEntry *bool   `json:"allows_entry"`
	AllowsExit  *bool   `json:"allows_exit"`
	IsDefault   *bool   `json:"is_default"`
	Active      *bool   `json:"active"`
}