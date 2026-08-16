package users

// DTO is what the API exposes. PasswordHash, FailedLoginCount, LockedUntil,
// DeletedAt etc. never leave this package in a response.
type DTO struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
}

func ToDTO(u *User) DTO {
	return DTO{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    string(u.Status),
	}
}

// CreateInput is the explicit allow-list of fields a caller may set when
// creating a user. Never bind request JSON straight onto the User model —
// that's how mass-assignment bugs (e.g. a client setting status=active or
// failed_login_count) happen.
type CreateInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdateInput struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}
