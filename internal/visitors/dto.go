package visitors

// --- Visitor DTOs -----------------------------------------------------

type DTO struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	IDTypeID  int64   `json:"id_type_id"`
	IDNumber  *string `json:"id_number"`
	CompanyID *int64  `json:"company_id"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	PhotoRef  *string `json:"photo_ref"`
	Notes     *string `json:"notes"`
	Source    string  `json:"source"`
	Status    string  `json:"status"`
}

func ToDTO(v *Visitor) DTO {
	return DTO{
		ID: v.ID, FirstName: v.FirstName, LastName: v.LastName,
		IDTypeID: v.IDTypeID, IDNumber: v.IDNumber, CompanyID: v.CompanyID,
		Phone: v.Phone, Email: v.Email, PhotoRef: v.PhotoRef, Notes: v.Notes,
		Source: string(v.Source), Status: string(v.Status),
	}
}

// CreateInput is the explicit allow-list for visitor creation. "source" is
// intentionally a bool flag (WantsPreRegistration), not a free-text field —
// the service decides the resulting Source based on that flag AND whether
// the tenant setting permits it, rather than trusting a client-supplied
// enum value directly (mass-assignment: a client could otherwise just send
// source=pre_registered regardless of the tenant setting).
type CreateInput struct {
	FirstName            string  `json:"first_name"`
	LastName             string  `json:"last_name"`
	IDTypeID             int64   `json:"id_type_id"`
	IDNumber             *string `json:"id_number"`
	CompanyID            *int64  `json:"company_id"`
	Phone                *string `json:"phone"`
	Email                *string `json:"email"`
	Notes                *string `json:"notes"`
	WantsPreRegistration bool    `json:"pre_register"`
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

// --- ID Type DTOs -------------------------------------------------------

type IDTypeDTO struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	RequiresNumber bool   `json:"requires_number"`
	Active         bool   `json:"active"`
}

func IDTypeToDTO(t *IDType) IDTypeDTO {
	return IDTypeDTO{ID: t.ID, Name: t.Name, Code: t.Code, RequiresNumber: t.RequiresNumber, Active: t.Active}
}

type IDTypeInput struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	RequiresNumber *bool  `json:"requires_number"`
	Active         *bool  `json:"active"`
}

// --- Company DTOs ---------------------------------------------------------

type CompanyDTO struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Phone   *string `json:"phone"`
	Email   *string `json:"email"`
	Address *string `json:"address"`
	Active  bool    `json:"active"`
}

func CompanyToDTO(c *Company) CompanyDTO {
	return CompanyDTO{ID: c.ID, Name: c.Name, Phone: c.Phone, Email: c.Email, Address: c.Address, Active: c.Active}
}

type CompanyInput struct {
	Name    string  `json:"name"`
	Phone   *string `json:"phone"`
	Email   *string `json:"email"`
	Address *string `json:"address"`
	Active  *bool   `json:"active"`
}
