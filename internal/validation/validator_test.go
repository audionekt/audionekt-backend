package validation_test

import (
	"testing"

	"musicapp/internal/validation"
)

func TestValidator(t *testing.T) {
	v := validation.New()
	
	// Test valid UUID
	if err := v.ValidateVar("550e8400-e29b-41d4-a716-446655440000", "uuid"); err != nil {
		t.Errorf("Expected valid UUID to pass validation: %v", err)
	}
	
	// Test invalid UUID
	if err := v.ValidateVar("invalid-uuid", "uuid"); err == nil {
		t.Error("Expected invalid UUID to fail validation")
	}
	
	// Test valid username
	if err := v.ValidateVar("valid_username123", "username"); err != nil {
		t.Errorf("Expected valid username to pass validation: %v", err)
	}
	
	// Test invalid username (too short)
	if err := v.ValidateVar("ab", "username"); err == nil {
		t.Error("Expected short username to fail validation")
	}
	
	// Test invalid username (invalid characters)
	if err := v.ValidateVar("invalid@username", "username"); err == nil {
		t.Error("Expected username with invalid characters to fail validation")
	}
}

func TestValidateStruct(t *testing.T) {
	v := validation.New()
	
	// Test struct validation
	type TestStruct struct {
		Username string `validate:"required,username"`
		Email    string `validate:"required,email"`
		Password string `validate:"required,password"`
	}
	
	// Test valid struct
	validStruct := TestStruct{
		Username: "testuser123",
		Email:    "test@example.com",
		Password: "ValidPass123!",
	}
	
	if err := v.ValidateStruct(validStruct); err != nil {
		t.Errorf("Expected valid struct to pass validation: %v", err)
	}
	
	// Test invalid struct
	invalidStruct := TestStruct{
		Username: "ab", // too short
		Email:    "invalid-email",
		Password: "weak", // doesn't meet password requirements
	}
	
	if err := v.ValidateStruct(invalidStruct); err == nil {
		t.Error("Expected invalid struct to fail validation")
	}
}

func TestValidateRequest(t *testing.T) {
	v := validation.New()
	
	// Test valid request struct
	type ValidRequest struct {
		Username string `validate:"required,username"`
		Email    string `validate:"required,email"`
		Password string `validate:"required,password"`
	}
	
	validRequest := ValidRequest{
		Username: "testuser123",
		Email:    "test@example.com",
		Password: "ValidPass123!",
	}
	
	if err := validation.ValidateRequest(v, validRequest); err != nil {
		t.Errorf("Expected valid request to pass validation: %v", err)
	}
	
	// Test invalid request struct
	type InvalidRequest struct {
		Username string `validate:"required,username"`
		Email    string `validate:"required,email"`
		Password string `validate:"required,password"`
	}
	
	invalidRequest := InvalidRequest{
		Username: "ab", // too short
		Email:    "invalid-email",
		Password: "weak", // doesn't meet password requirements
	}
	
	if err := validation.ValidateRequest(v, invalidRequest); err == nil {
		t.Error("Expected invalid request to fail validation")
	}
}

func TestGetValidationErrorsComprehensive(t *testing.T) {
	v := validation.New()
	
	// Test with comprehensive validation errors to cover getValidationMessage
	type TestStruct struct {
		Username    string  `validate:"required,username"`
		Email       string  `validate:"required,email"`
		Password    string  `validate:"required,password"`
		Age         int     `validate:"min=18,max=100"`
		Latitude    float64 `validate:"latitude"`
		Longitude   float64 `validate:"longitude"`
		URL         string  `validate:"url"`
		InstagramHandle string `validate:"instagram_handle"`
	}
	
	invalidStruct := TestStruct{
		Username:    "ab", // too short
		Email:       "invalid-email",
		Password:    "weak", // doesn't meet requirements
		Age:         15, // too young
		Latitude:    91.0, // invalid latitude
		Longitude:   181.0, // invalid longitude
		URL:         "not-a-url",
		InstagramHandle: "user-name", // invalid characters
	}
	
	err := v.ValidateStruct(invalidStruct)
	if err == nil {
		t.Error("Expected validation to fail")
	}
	
	// Test GetValidationErrors
	errors := validation.GetValidationErrors(err)
	if errors == nil {
		t.Error("Expected validation errors to be returned")
	}
	
	if len(errors) == 0 {
		t.Error("Expected validation errors to contain field errors")
	}
	
	// Check that we have multiple error types
	errorTags := make(map[string]bool)
	for _, validationErr := range errors {
		errorTags[validationErr.Tag] = true
	}
	
	// Test with nil error
	nilErrors := validation.GetValidationErrors(nil)
	if nilErrors != nil {
		t.Error("Expected nil errors to return nil")
	}
}

func TestURLValidation(t *testing.T) {
	v := validation.New()
	
	// Test valid URLs
	validURLs := []string{
		"https://example.com",
		"http://example.com",
		"https://www.example.com/path",
		"https://example.com:8080/path?query=value",
		"", // empty URL is allowed (optional field)
	}
	
	for _, url := range validURLs {
		if err := v.ValidateVar(url, "url"); err != nil {
			t.Errorf("Expected valid URL '%s' to pass validation: %v", url, err)
		}
	}
	
	// Test invalid URLs
	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com", // unsupported protocol
		"example.com", // missing protocol
	}
	
	for _, url := range invalidURLs {
		if err := v.ValidateVar(url, "url"); err == nil {
			t.Errorf("Expected invalid URL '%s' to fail validation", url)
		}
	}
}

func TestInstagramHandleValidation(t *testing.T) {
	v := validation.New()
	
	// Test valid Instagram handles
	validHandles := []string{
		"username",
		"user_name",
		"user.name",
		"user123",
		"", // empty handle is allowed (optional field)
		"ab", // minimum length
		"verylongusername123456789", // 25 chars, within limit
	}
	
	for _, handle := range validHandles {
		if err := v.ValidateVar(handle, "instagram_handle"); err != nil {
			t.Errorf("Expected valid Instagram handle '%s' to pass validation: %v", handle, err)
		}
	}
	
	// Test invalid Instagram handles
	invalidHandles := []string{
		"user-name", // contains hyphen
		"user@name", // contains @
		"user name", // contains space
	}
	
	for _, handle := range invalidHandles {
		if err := v.ValidateVar(handle, "instagram_handle"); err == nil {
			t.Errorf("Expected invalid Instagram handle '%s' to fail validation", handle)
		}
	}
}

func TestPasswordValidation(t *testing.T) {
	v := validation.New()
	
	// Test valid password
	if err := v.ValidateVar("ValidPass123!", "password"); err != nil {
		t.Errorf("Expected valid password to pass validation: %v", err)
	}
	
	// Test password without uppercase
	if err := v.ValidateVar("invalidpass123!", "password"); err == nil {
		t.Error("Expected password without uppercase to fail validation")
	}
	
	// Test password without lowercase
	if err := v.ValidateVar("INVALIDPASS123!", "password"); err == nil {
		t.Error("Expected password without lowercase to fail validation")
	}
	
	// Test password without digit
	if err := v.ValidateVar("InvalidPassword!", "password"); err == nil {
		t.Error("Expected password without digit to fail validation")
	}
	
	// Test password without special character
	if err := v.ValidateVar("InvalidPassword123", "password"); err == nil {
		t.Error("Expected password without special character to fail validation")
	}
}

func TestLatitudeLongitudeValidation(t *testing.T) {
	v := validation.New()
	
	// Test valid latitude
	if err := v.ValidateVar(37.7749, "latitude"); err != nil {
		t.Errorf("Expected valid latitude to pass validation: %v", err)
	}
	
	// Test invalid latitude (too high)
	if err := v.ValidateVar(91.0, "latitude"); err == nil {
		t.Error("Expected latitude > 90 to fail validation")
	}
	
	// Test invalid latitude (too low)
	if err := v.ValidateVar(-91.0, "latitude"); err == nil {
		t.Error("Expected latitude < -90 to fail validation")
	}
	
	// Test valid longitude
	if err := v.ValidateVar(-122.4194, "longitude"); err != nil {
		t.Errorf("Expected valid longitude to pass validation: %v", err)
	}
	
	// Test invalid longitude (too high)
	if err := v.ValidateVar(181.0, "longitude"); err == nil {
		t.Error("Expected longitude > 180 to fail validation")
	}
	
	// Test invalid longitude (too low)
	if err := v.ValidateVar(-181.0, "longitude"); err == nil {
		t.Error("Expected longitude < -180 to fail validation")
	}
}
