package visitors

import "time"

// DTO is the API shape used by both visitor lists and detail screens.
// Display fields are enriched server-side so the frontend never has to guess
// names from IDs or issue N+1 lookup requests.
type DTO struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	FullName  string  `json:"full_name"`
	IDTypeID  int64   `json:"id_type_id"`
	IDTypeName string `json:"id_type_name,omitempty"`
	IDNumber  *string `json:"id_number"`
	CompanyID *int64  `json:"company_id"`
	CompanyName string `json:"company_name,omitempty"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	PhotoRef  *string `json:"photo_ref"`
	Notes     *string `json:"notes"`
	Source    string  `json:"source"`
	Status    string  `json:"status"`
	Blacklisted bool  `json:"blacklisted"`
	BlacklistReason *string `json:"blacklist_reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToDTO(v *Visitor) DTO {
	return DTO{
		ID: v.ID, FirstName: v.FirstName, LastName: v.LastName, FullName: v.FullName(),
		IDTypeID: v.IDTypeID, IDNumber: v.IDNumber, CompanyID: v.CompanyID,
		Phone: v.Phone, Email: v.Email, PhotoRef: v.PhotoRef, Notes: v.Notes,
		Source: string(v.Source), Status: string(v.Status),
		Blacklisted: v.Status == StatusBlacklisted, BlacklistReason: v.BlacklistReason,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}

type CreateInput struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	IDTypeID int64 `json:"id_type_id"`
	IDNumber *string `json:"id_number"`
	CompanyID *int64 `json:"company_id"`
	Phone *string `json:"phone"`
	Email *string `json:"email"`
	Notes *string `json:"notes"`
	PhotoRef *string `json:"photo_ref"`
	WantsPreRegistration bool `json:"pre_register"`
}

type UpdateInput struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	IDTypeID  *int64  `json:"id_type_id"`
	IDNumber  *string `json:"id_number"`
	CompanyID *int64  `json:"company_id"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	Notes     *string `json:"notes"`
}

type BlacklistInput struct {
	Blacklisted bool    `json:"blacklisted"`
	Reason      *string `json:"reason"`
}

type IDTypeDTO struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	RequiresNumber bool `json:"requires_number"`
	Active bool `json:"active"`
}
func IDTypeToDTO(t *IDType) IDTypeDTO { return IDTypeDTO{ID:t.ID,Name:t.Name,Code:t.Code,RequiresNumber:t.RequiresNumber,Active:t.Active} }
type IDTypeInput struct {
	Name string `json:"name"`
	Code string `json:"code"`
	RequiresNumber *bool `json:"requires_number"`
	Active *bool `json:"active"`
}

type CompanyDTO struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	Phone *string `json:"phone"`
	Email *string `json:"email"`
	Address *string `json:"address"`
	Active bool `json:"active"`
}
func CompanyToDTO(c *Company) CompanyDTO { return CompanyDTO{ID:c.ID,Name:c.Name,Phone:c.Phone,Email:c.Email,Address:c.Address,Active:c.Active} }
type CompanyInput struct {
	Name string `json:"name"`
	Phone *string `json:"phone"`
	Email *string `json:"email"`
	Address *string `json:"address"`
	Active *bool `json:"active"`
}


type IdentityMatch struct { ID int64 `json:"id"`; FullName string `json:"full_name"`; IDTypeID int64 `json:"id_type_id"`; IDNumber *string `json:"id_number"`; Phone *string `json:"phone"`; Email *string `json:"email"`; CompanyName string `json:"company_name,omitempty"`; Status string `json:"status"`; Blacklisted bool `json:"blacklisted"` }
