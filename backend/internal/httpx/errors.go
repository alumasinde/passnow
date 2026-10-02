package httpx

import "net/http"


type AppError struct {
	Code    string 
	Message string 
	Status  int    
}

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
	ErrServiceUnavailable = AppError{
		Code: "service_unavailable", Message: "service temporarily unavailable, please try again", Status: http.StatusServiceUnavailable,
	}
)

func (e AppError) WithMessage(msg string) AppError {
	e.Message = msg
	return e
}
