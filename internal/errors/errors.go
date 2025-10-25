package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Authentication errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired      ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid      ErrorCode = "TOKEN_INVALID"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"

	// Resource errors
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"

	// Database errors
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrCodeQueryFailed   ErrorCode = "QUERY_FAILED"

	// External service errors
	ErrCodeExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeS3Error         ErrorCode = "S3_ERROR"
	ErrCodeRedisError      ErrorCode = "REDIS_ERROR"

	// Business logic errors
	ErrCodeBusinessRule ErrorCode = "BUSINESS_RULE_VIOLATION"
	ErrCodeRateLimited ErrorCode = "RATE_LIMITED"

	// Internal errors
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
	}
}

// NewWithDetails creates a new AppError with details
func NewWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getHTTPStatus(code),
	}
}

// Wrap wraps an existing error with an AppError
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		Cause:      err,
	}
}

// WrapWithDetails wraps an existing error with an AppError and details
func WrapWithDetails(err error, code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getHTTPStatus(code),
		Cause:      err,
	}
}

// getHTTPStatus maps error codes to HTTP status codes
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeInvalidCredentials, ErrCodeTokenExpired, ErrCodeTokenInvalid:
		return http.StatusUnauthorized
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeValidationFailed, ErrCodeInvalidInput, ErrCodeMissingField:
		return http.StatusBadRequest
	case ErrCodeNotFound, ErrCodeResourceNotFound:
		return http.StatusNotFound
	case ErrCodeAlreadyExists, ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeDatabaseError, ErrCodeQueryFailed:
		return http.StatusInternalServerError
	case ErrCodeExternalService, ErrCodeS3Error, ErrCodeRedisError:
		return http.StatusServiceUnavailable
	case ErrCodeBusinessRule:
		return http.StatusUnprocessableEntity
	case ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case ErrCodeNotImplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// Predefined common errors
var (
	ErrInvalidCredentials = New(ErrCodeInvalidCredentials, "Invalid email or password")
	ErrTokenExpired       = New(ErrCodeTokenExpired, "Token has expired")
	ErrTokenInvalid       = New(ErrCodeTokenInvalid, "Invalid token")
	ErrUnauthorized       = New(ErrCodeUnauthorized, "Unauthorized access")
	ErrForbidden          = New(ErrCodeForbidden, "Access forbidden")
	ErrValidationFailed   = New(ErrCodeValidationFailed, "Validation failed")
	ErrInvalidInput       = New(ErrCodeInvalidInput, "Invalid input provided")
	ErrMissingField       = New(ErrCodeMissingField, "Required field is missing")
	ErrNotFound           = New(ErrCodeNotFound, "Resource not found")
	ErrAlreadyExists      = New(ErrCodeAlreadyExists, "Resource already exists")
	ErrConflict           = New(ErrCodeConflict, "Resource conflict")
	ErrDatabaseError      = New(ErrCodeDatabaseError, "Database operation failed")
	ErrQueryFailed        = New(ErrCodeQueryFailed, "Database query failed")
	ErrExternalService    = New(ErrCodeExternalService, "External service error")
	ErrS3Error            = New(ErrCodeS3Error, "S3 operation failed")
	ErrRedisError         = New(ErrCodeRedisError, "Redis operation failed")
	ErrBusinessRule       = New(ErrCodeBusinessRule, "Business rule violation")
	ErrRateLimited        = New(ErrCodeRateLimited, "Rate limit exceeded")
	ErrInternalError      = New(ErrCodeInternalError, "Internal server error")
	ErrNotImplemented     = New(ErrCodeNotImplemented, "Feature not implemented")
)

// Helper functions for common error patterns
func NewUserNotFound(userID string) *AppError {
	return NewWithDetails(ErrCodeNotFound, "User not found", fmt.Sprintf("User with ID %s not found", userID))
}

func NewUserAlreadyExists(field, value string) *AppError {
	return NewWithDetails(ErrCodeAlreadyExists, "User already exists", fmt.Sprintf("User with %s '%s' already exists", field, value))
}

func NewBandNotFound(bandID string) *AppError {
	return NewWithDetails(ErrCodeNotFound, "Band not found", fmt.Sprintf("Band with ID %s not found", bandID))
}

func NewPostNotFound(postID string) *AppError {
	return NewWithDetails(ErrCodeNotFound, "Post not found", fmt.Sprintf("Post with ID %s not found", postID))
}

func NewValidationError(field, reason string) *AppError {
	return NewWithDetails(ErrCodeValidationFailed, "Validation failed", fmt.Sprintf("Field '%s': %s", field, reason))
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from an error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}
