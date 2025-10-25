package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig *Config
	}{
		{
			name: "load with default values",
			envVars: map[string]string{
				// No environment variables set
			},
			expectedConfig: &Config{
				DatabaseURL:        "postgres://dev:devpassword@localhost:5432/musicapp?sslmode=disable",
				RedisURL:           "redis://localhost:6379",
				JWTSecret:          "your-super-secret-jwt-key-change-in-production",
				AWSRegion:          "us-east-1",
				AWSAccessKeyID:     "",
				AWSSecretAccessKey: "",
				S3BucketName:       "",
				S3CDNURL:           "",
				Port:               8080,
				Environment:        "development",
			},
		},
		{
			name: "load with custom environment variables",
			envVars: map[string]string{
				"DATABASE_URL":           "postgres://custom:password@localhost:5432/testdb",
				"REDIS_URL":              "redis://custom:6379",
				"JWT_SECRET":             "custom-jwt-secret",
				"AWS_REGION":             "us-west-2",
				"AWS_ACCESS_KEY_ID":      "custom-access-key",
				"AWS_SECRET_ACCESS_KEY":  "custom-secret-key",
				"S3_BUCKET_NAME":         "custom-bucket",
				"S3_CDN_URL":             "https://custom-cdn.com",
				"PORT":                   "3000",
				"ENVIRONMENT":            "production",
			},
			expectedConfig: &Config{
				DatabaseURL:        "postgres://custom:password@localhost:5432/testdb",
				RedisURL:           "redis://custom:6379",
				JWTSecret:          "custom-jwt-secret",
				AWSRegion:          "us-west-2",
				AWSAccessKeyID:     "custom-access-key",
				AWSSecretAccessKey: "custom-secret-key",
				S3BucketName:       "custom-bucket",
				S3CDNURL:           "https://custom-cdn.com",
				Port:               3000,
				Environment:        "production",
			},
		},
		{
			name: "load with partial environment variables",
			envVars: map[string]string{
				"DATABASE_URL": "postgres://partial:password@localhost:5432/partialdb",
				"PORT":         "9000",
				"ENVIRONMENT":  "staging",
			},
			expectedConfig: &Config{
				DatabaseURL:        "postgres://partial:password@localhost:5432/partialdb",
				RedisURL:           "redis://localhost:6379", // Default
				JWTSecret:          "your-super-secret-jwt-key-change-in-production", // Default
				AWSRegion:          "us-east-1", // Default
				AWSAccessKeyID:     "",
				AWSSecretAccessKey: "",
				S3BucketName:       "",
				S3CDNURL:           "",
				Port:               9000,
				Environment:        "staging",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing environment variables
			clearEnvVars()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			config := Load()

			// Verify configuration
			if config.DatabaseURL != tt.expectedConfig.DatabaseURL {
				t.Errorf("Expected DatabaseURL %s, got %s", tt.expectedConfig.DatabaseURL, config.DatabaseURL)
			}
			if config.RedisURL != tt.expectedConfig.RedisURL {
				t.Errorf("Expected RedisURL %s, got %s", tt.expectedConfig.RedisURL, config.RedisURL)
			}
			if config.JWTSecret != tt.expectedConfig.JWTSecret {
				t.Errorf("Expected JWTSecret %s, got %s", tt.expectedConfig.JWTSecret, config.JWTSecret)
			}
			if config.AWSRegion != tt.expectedConfig.AWSRegion {
				t.Errorf("Expected AWSRegion %s, got %s", tt.expectedConfig.AWSRegion, config.AWSRegion)
			}
			if config.AWSAccessKeyID != tt.expectedConfig.AWSAccessKeyID {
				t.Errorf("Expected AWSAccessKeyID %s, got %s", tt.expectedConfig.AWSAccessKeyID, config.AWSAccessKeyID)
			}
			if config.AWSSecretAccessKey != tt.expectedConfig.AWSSecretAccessKey {
				t.Errorf("Expected AWSSecretAccessKey %s, got %s", tt.expectedConfig.AWSSecretAccessKey, config.AWSSecretAccessKey)
			}
			if config.S3BucketName != tt.expectedConfig.S3BucketName {
				t.Errorf("Expected S3BucketName %s, got %s", tt.expectedConfig.S3BucketName, config.S3BucketName)
			}
			if config.S3CDNURL != tt.expectedConfig.S3CDNURL {
				t.Errorf("Expected S3CDNURL %s, got %s", tt.expectedConfig.S3CDNURL, config.S3CDNURL)
			}
			if config.Port != tt.expectedConfig.Port {
				t.Errorf("Expected Port %d, got %d", tt.expectedConfig.Port, config.Port)
			}
			if config.Environment != tt.expectedConfig.Environment {
				t.Errorf("Expected Environment %s, got %s", tt.expectedConfig.Environment, config.Environment)
			}

			// Clean up environment variables
			clearEnvVars()
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "environment variable is empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "environment variable with spaces",
			key:          "SPACE_VAR",
			defaultValue: "default",
			envValue:     "  spaced value  ",
			expected:     "  spaced value  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing environment variable
			os.Unsetenv(tt.key)

			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			// Test getEnv function
			result := getEnv(tt.key, tt.defaultValue)

			// Verify result
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "valid integer environment variable",
			key:          "TEST_INT",
			defaultValue: 8080,
			envValue:     "3000",
			expected:     3000,
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_INT",
			defaultValue: 8080,
			envValue:     "",
			expected:     8080,
		},
		{
			name:         "invalid integer environment variable",
			key:          "INVALID_INT",
			defaultValue: 8080,
			envValue:     "not_a_number",
			expected:     8080,
		},
		{
			name:         "zero integer environment variable",
			key:          "ZERO_INT",
			defaultValue: 8080,
			envValue:     "0",
			expected:     0,
		},
		{
			name:         "negative integer environment variable",
			key:          "NEGATIVE_INT",
			defaultValue: 8080,
			envValue:     "-1",
			expected:     -1,
		},
		{
			name:         "large integer environment variable",
			key:          "LARGE_INT",
			defaultValue: 8080,
			envValue:     "65535",
			expected:     65535,
		},
		{
			name:         "empty string environment variable",
			key:          "EMPTY_INT",
			defaultValue: 8080,
			envValue:     "",
			expected:     8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing environment variable
			os.Unsetenv(tt.key)

			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			// Test getEnvAsInt function
			result := getEnvAsInt(tt.key, tt.defaultValue)

			// Verify result
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
}

// Helper function to clear environment variables
func clearEnvVars() {
	envVars := []string{
		"DATABASE_URL",
		"REDIS_URL",
		"JWT_SECRET",
		"AWS_REGION",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"S3_BUCKET_NAME",
		"S3_CDN_URL",
		"PORT",
		"ENVIRONMENT",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
