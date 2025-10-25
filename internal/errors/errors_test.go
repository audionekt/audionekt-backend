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

func TestPredefinedErrors(t *testing.T) {
	if apperrors.ErrNotFound.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected ErrNotFound to have status %d, got %d", http.StatusNotFound, apperrors.ErrNotFound.HTTPStatus)
	}
	
	if apperrors.ErrInvalidCredentials.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("Expected ErrInvalidCredentials to have status %d, got %d", http.StatusUnauthorized, apperrors.ErrInvalidCredentials.HTTPStatus)
	}
}
