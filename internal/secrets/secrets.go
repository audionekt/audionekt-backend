package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

// SecretManager handles secure secret generation and management
type SecretManager struct {
	jwtSecret []byte
}

// NewSecretManager creates a new secret manager
func NewSecretManager() (*SecretManager, error) {
	jwtSecret, err := getOrGenerateJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWT secret: %w", err)
	}

	return &SecretManager{
		jwtSecret: jwtSecret,
	}, nil
}

// GetJWTSecret returns the JWT secret
func (sm *SecretManager) GetJWTSecret() []byte {
	return sm.jwtSecret
}

// getOrGenerateJWTSecret gets JWT secret from environment or generates a new one
func getOrGenerateJWTSecret() ([]byte, error) {
	// First, try to get from environment variable
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		if len(secret) < 32 {
			return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
		}
		return []byte(secret), nil
	}

	// If not in environment, check if we're in development mode
	if os.Getenv("ENVIRONMENT") == "development" {
		// Generate a random secret for development
		secret := generateRandomSecret(64)
		fmt.Printf("WARNING: Generated random JWT secret for development. Set JWT_SECRET environment variable for production.\n")
		return secret, nil
	}

	// In production, JWT_SECRET must be set
	return nil, fmt.Errorf("JWT_SECRET environment variable is required in production")
}

// generateRandomSecret generates a cryptographically secure random secret
func generateRandomSecret(length int) []byte {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random secret: %v", err))
	}
	return []byte(base64.URLEncoding.EncodeToString(bytes))
}

// ValidateSecretStrength checks if a secret meets security requirements
func ValidateSecretStrength(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("secret must be at least 32 characters long")
	}
	
	// Check for common weak secrets
	weakSecrets := []string{
		"your-super-secret-jwt-key-change-in-production",
		"secret",
		"password",
		"1234567890",
		"jwt-secret",
		"change-me",
	}
	
	for _, weak := range weakSecrets {
		if secret == weak {
			return fmt.Errorf("secret is too weak, please use a stronger secret")
		}
	}
	
	return nil
}
