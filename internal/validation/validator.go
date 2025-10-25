package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Validator wraps the go-playground validator with custom rules
type Validator struct {
	validator *validator.Validate
}

// New creates a new validator with custom validation rules
func New() *Validator {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("uuid", validateUUID)
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("latitude", validateLatitude)
	v.RegisterValidation("longitude", validateLongitude)
	v.RegisterValidation("url", validateURL)
	v.RegisterValidation("instagram_handle", validateInstagramHandle)
	
	return &Validator{
		validator: v,
	}
}

// ValidateStruct validates a struct using the registered validation rules
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validator.Struct(s)
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validator.Var(field, tag)
}

// Custom validation functions

// validateUUID validates UUID format
func validateUUID(fl validator.FieldLevel) bool {
	uuidStr := fl.Field().String()
	if uuidStr == "" {
		return true // Allow empty UUIDs (optional fields)
	}
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	
	// Username must be 3-50 characters, alphanumeric and underscores only
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	
	// Check for valid characters (alphanumeric, underscore, hyphen)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if !matched {
		return false
	}
	
	// Cannot start or end with underscore or hyphen
	if strings.HasPrefix(username, "_") || strings.HasPrefix(username, "-") ||
		strings.HasSuffix(username, "_") || strings.HasSuffix(username, "-") {
		return false
	}
	
	return true
}

// validatePassword validates password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	// Password must be at least 8 characters
	if len(password) < 8 {
		return false
	}
	
	// Password must contain at least one uppercase letter
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return false
	}
	
	// Password must contain at least one lowercase letter
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	if !hasLower {
		return false
	}
	
	// Password must contain at least one digit
	hasDigit, _ := regexp.MatchString(`[0-9]`, password)
	if !hasDigit {
		return false
	}
	
	// Password must contain at least one special character
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
	if !hasSpecial {
		return false
	}
	
	return true
}

// validateLatitude validates latitude range
func validateLatitude(fl validator.FieldLevel) bool {
	lat := fl.Field().Float()
	return lat >= -90 && lat <= 90
}

// validateLongitude validates longitude range
func validateLongitude(fl validator.FieldLevel) bool {
	lng := fl.Field().Float()
	return lng >= -180 && lng <= 180
}

// validateURL validates URL format
func validateURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true // Allow empty URLs (optional fields)
	}
	
	// Basic URL validation
	matched, _ := regexp.MatchString(`^https?:\/\/[^\s/$.?#].[^\s]*$`, url)
	return matched
}

// validateInstagramHandle validates Instagram handle format
func validateInstagramHandle(fl validator.FieldLevel) bool {
	handle := fl.Field().String()
	if handle == "" {
		return true // Allow empty handles (optional fields)
	}
	
	// Remove @ if present
	handle = strings.TrimPrefix(handle, "@")
	
	// Instagram handle must be 1-30 characters, alphanumeric, periods, underscores only
	if len(handle) < 1 || len(handle) > 30 {
		return false
	}
	
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._]+$`, handle)
	return matched
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// GetValidationErrors extracts validation errors from validator.ValidationErrors
func GetValidationErrors(err error) []ValidationError {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make([]ValidationError, len(validationErrors))
		for i, err := range validationErrors {
			errors[i] = ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: getValidationMessage(err),
			}
		}
		return errors
	}
	return nil
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", err.Field(), err.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "username":
		return fmt.Sprintf("%s must be 3-50 characters, alphanumeric with underscores and hyphens only", err.Field())
	case "password":
		return fmt.Sprintf("%s must be at least 8 characters with uppercase, lowercase, digit, and special character", err.Field())
	case "latitude":
		return fmt.Sprintf("%s must be between -90 and 90", err.Field())
	case "longitude":
		return fmt.Sprintf("%s must be between -180 and 180", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "instagram_handle":
		return fmt.Sprintf("%s must be a valid Instagram handle", err.Field())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}

// ValidateRequest validates a request struct and returns formatted errors
func ValidateRequest(v *Validator, req interface{}) error {
	if err := v.ValidateStruct(req); err != nil {
		return err
	}
	return nil
}
