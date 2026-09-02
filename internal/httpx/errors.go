package httpx

import "net/http"

// AppError is a named, reusable error response. Add new ones here as the
// system grows (e.g. ErrGatepassNotFound, ErrApprovalStepMismatch) instead
// of typing status codes and message strings at each call site.
type AppError struct {
	Code    string // stable machine-readable code, e.g. "auth_required"
	Message string // human-readable text shown to the client
	Status  int    // HTTP status code
}

// ---------------------------------------------------------------------
// EDIT THIS FILE to change wording, status codes, or add new error types.
// Every handler references these by name (e.g. httpx.ErrAuthRequired) —
// nothing else in the codebase should hardcode a message string or status.
// ---------------------------------------------------------------------

var (
	ErrAuthRequired = AppError{
		Code: "auth_required", Message: "authentication required", Status: http.StatusUnauthorized,
	}
	ErrInvalidCredentials = AppError{
		Code: "invalid_credentials", Message: "invalid email or password", Status: http.StatusUnauthorized,
	}
	ErrAccountLocked = AppError{
		Code: "account_locked", Message: "account temporarily locked, try again later", Status: http.StatusTooManyRequests,
	}
	ErrAccountDisabled = AppError{
		Code: "account_disabled", Message: "account disabled", Status: http.StatusForbidden,
	}
	ErrInvalidRefreshToken = AppError{
		Code: "invalid_refresh_token", Message: "invalid or expired refresh token", Status: http.StatusUnauthorized,
	}
	ErrForbidden = AppError{
		Code: "forbidden", Message: "insufficient permissions", Status: http.StatusForbidden,
	}
	ErrTenantNotFound = AppError{
		Code: "tenant_not_found", Message: "tenant not found", Status: http.StatusNotFound,
	}
	ErrBadRequestBody = AppError{
		Code: "bad_request", Message: "invalid request body", Status: http.StatusBadRequest,
	}
	ErrValidation = AppError{
		Code: "validation_failed", Message: "one or more fields are invalid", Status: http.StatusUnprocessableEntity,
	}
	ErrNotFound = AppError{
		Code: "not_found", Message: "resource not found", Status: http.StatusNotFound,
	}
	ErrConflict = AppError{
		Code: "conflict", Message: "request conflicts with the current resource state", Status: http.StatusConflict,
	}
	ErrInternal = AppError{
		Code: "internal_error", Message: "something went wrong, please try again", Status: http.StatusInternalServerError,
	}
)

// WithMessage returns a copy of the AppError with a different message —
// use for validation errors where the detail varies per request
// (e.g. httpx.ErrValidation.WithMessage("email and password are required"))
// while keeping the same code/status.
func (e AppError) WithMessage(msg string) AppError {
	e.Message = msg
	return e
}
