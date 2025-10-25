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
