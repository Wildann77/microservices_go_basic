package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents structured error codes
type ErrorCode string

const (
	// 4xx Client Errors
	ErrInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrForbidden        ErrorCode = "FORBIDDEN"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrConflict         ErrorCode = "CONFLICT"
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrRateLimit        ErrorCode = "RATE_LIMIT_EXCEEDED"

	// 5xx Server Errors
	ErrInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrDatabaseError  ErrorCode = "DATABASE_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Service    string    `json:"service,omitempty"`
	StatusCode int       `json:"-"`
	Err        error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// ToJSON returns JSON representation of the error
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(map[string]interface{}{
		"error": e,
	})
	return data
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getStatusCode(code),
	}
}

// Wrap wraps an existing error with AppError
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    err.Error(),
		StatusCode: getStatusCode(code),
		Err:        err,
	}
}

// WithService adds service name to error
func (e *AppError) WithService(service string) *AppError {
	e.Service = service
	return e
}

// WithDetails adds details to error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// IsNotFound checks if error is not found
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrNotFound
	}
	return false
}

// IsValidationError checks if error is validation error
func IsValidationError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrValidationFailed || appErr.Code == ErrInvalidInput
	}
	return false
}

// WriteHTTPResponse writes error as HTTP response
func (e *AppError) WriteHTTPResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	w.Write(e.ToJSON())
}

func getStatusCode(code ErrorCode) int {
	switch code {
	case ErrInvalidInput, ErrValidationFailed:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrRateLimit:
		return http.StatusTooManyRequests
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Common errors
var (
	ErrUserNotFound     = New(ErrNotFound, "User not found")
	ErrOrderNotFound    = New(ErrNotFound, "Order not found")
	ErrPaymentNotFound  = New(ErrNotFound, "Payment not found")
	ErrInvalidToken     = New(ErrUnauthorized, "Invalid or expired token")
	ErrInsufficientFund = New(ErrForbidden, "Insufficient funds")
)