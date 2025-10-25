package secrets_test

import (
	"os"
	"testing"

	"musicapp/internal/secrets"
)

func TestSecretManager(t *testing.T) {
	// Test with environment variable
	os.Setenv("JWT_SECRET", "test-secret-that-is-long-enough-for-validation")
	defer os.Unsetenv("JWT_SECRET")

	sm, err := secrets.NewSecretManager()
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	secret := sm.GetJWTSecret()
	if len(secret) == 0 {
		t.Error("Expected non-empty JWT secret")
	}
}

func TestSecretValidation(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid secret",
			secret:  "this-is-a-valid-secret-that-is-long-enough",
			wantErr: false,
		},
		{
			name:    "too short",
			secret:  "short",
			wantErr: true,
		},
		{
			name:    "weak secret",
			secret:  "your-super-secret-jwt-key-change-in-production",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := secrets.ValidateSecretStrength(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecretStrength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
