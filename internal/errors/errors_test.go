package errors_test

import (
	"errors"
	"net/http"
	"testing"

	apperrors "musicapp/internal/errors"
)

func TestAppError(t *testing.T) {
	err := apperrors.New(apperrors.ErrCodeNotFound, "Resource not found")
	
	if err.Code != apperrors.ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeNotFound, err.Code)
	}
	
	if err.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got '%s'", err.Message)
	}
	
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}
}

func TestAppErrorErrorMethod(t *testing.T) {
	err := apperrors.New(apperrors.ErrCodeNotFound, "Resource not found")
	
	errorString := err.Error()
	if errorString == "" {
		t.Error("Expected non-empty error string")
	}
	
	// Test with wrapped error
	originalErr := errors.New("original error")
	wrappedErr := apperrors.Wrap(originalErr, apperrors.ErrCodeDatabaseError, "Database operation failed")
	
	wrappedErrorString := wrappedErr.Error()
	if wrappedErrorString == "" {
		t.Error("Expected non-empty wrapped error string")
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := apperrors.Wrap(originalErr, apperrors.ErrCodeDatabaseError, "Database operation failed")
	
	unwrappedErr := wrappedErr.Unwrap()
	if unwrappedErr != originalErr {
		t.Errorf("Expected unwrapped error to be original error")
	}
	
	// Test unwrap on non-wrapped error
	nonWrappedErr := apperrors.New(apperrors.ErrCodeNotFound, "Not found")
	if nonWrappedErr.Unwrap() != nil {
		t.Error("Expected unwrapped error to be nil for non-wrapped error")
	}
}

func TestNewWithDetails(t *testing.T) {
	err := apperrors.NewWithDetails(apperrors.ErrCodeValidationFailed, "Validation failed", "Field 'username' is invalid")
	
	if err.Code != apperrors.ErrCodeValidationFailed {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeValidationFailed, err.Code)
	}
	
	if err.Message != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got '%s'", err.Message)
	}
	
	if err.Details != "Field 'username' is invalid" {
		t.Errorf("Expected details 'Field 'username' is invalid', got '%s'", err.Details)
	}
}

func TestWrapWithDetails(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := apperrors.WrapWithDetails(originalErr, apperrors.ErrCodeDatabaseError, "Database operation failed", "Insert operation failed")
	
	if wrappedErr.Code != apperrors.ErrCodeDatabaseError {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeDatabaseError, wrappedErr.Code)
	}
	
	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error")
	}
	
	if wrappedErr.Details != "Insert operation failed" {
		t.Errorf("Expected details 'Insert operation failed', got '%s'", wrappedErr.Details)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := apperrors.Wrap(originalErr, apperrors.ErrCodeDatabaseError, "Database operation failed")
	
	if wrappedErr.Code != apperrors.ErrCodeDatabaseError {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeDatabaseError, wrappedErr.Code)
	}
	
	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error")
	}
	
	if !errors.Is(wrappedErr, originalErr) {
		t.Errorf("Expected wrapped error to unwrap to original error")
	}
}

func TestGetAppError(t *testing.T) {
	appErr := apperrors.New(apperrors.ErrCodeNotFound, "Not found")
	regularErr := errors.New("regular error")
	
	if apperrors.GetAppError(appErr) == nil {
		t.Error("Expected to get AppError from AppError")
	}
	
	if apperrors.GetAppError(regularErr) != nil {
		t.Error("Expected nil from regular error")
	}
}

func TestIsAppError(t *testing.T) {
	appErr := apperrors.New(apperrors.ErrCodeNotFound, "Not found")
	regularErr := errors.New("regular error")
	
	if !apperrors.IsAppError(appErr) {
		t.Error("Expected IsAppError to return true for AppError")
	}
	
	if apperrors.IsAppError(regularErr) {
		t.Error("Expected IsAppError to return false for regular error")
	}
}

func TestSpecificErrorFunctions(t *testing.T) {
	// Test NewUserNotFound
	userNotFoundErr := apperrors.NewUserNotFound("user123")
	if userNotFoundErr.Code != apperrors.ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeNotFound, userNotFoundErr.Code)
	}
	
	// Test NewUserAlreadyExists
	userExistsErr := apperrors.NewUserAlreadyExists("email", "test@example.com")
	if userExistsErr.Code != apperrors.ErrCodeAlreadyExists {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeAlreadyExists, userExistsErr.Code)
	}
	
	// Test NewBandNotFound
	bandNotFoundErr := apperrors.NewBandNotFound("band123")
	if bandNotFoundErr.Code != apperrors.ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeNotFound, bandNotFoundErr.Code)
	}
	
	// Test NewPostNotFound
	postNotFoundErr := apperrors.NewPostNotFound("post123")
	if postNotFoundErr.Code != apperrors.ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeNotFound, postNotFoundErr.Code)
	}
	
	// Test NewValidationError
	validationErr := apperrors.NewValidationError("username", "must be at least 3 characters")
	if validationErr.Code != apperrors.ErrCodeValidationFailed {
		t.Errorf("Expected code %s, got %s", apperrors.ErrCodeValidationFailed, validationErr.Code)
	}
}

func TestPredefinedErrors(t *testing.T) {
	if apperrors.ErrNotFound.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected ErrNotFound to have status %d, got %d", http.StatusNotFound, apperrors.ErrNotFound.HTTPStatus)
	}
	
	if apperrors.ErrInvalidCredentials.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("Expected ErrInvalidCredentials to have status %d, got %d", http.StatusUnauthorized, apperrors.ErrInvalidCredentials.HTTPStatus)
	}
}
