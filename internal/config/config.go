package config

import (
	"log"
	"os"
	"strconv"

	"musicapp/internal/secrets"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret []byte

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3BucketName       string
	S3CDNURL           string

	// Server
	Port        int
	Environment string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize secret manager
	secretManager, err := secrets.NewSecretManager()
	if err != nil {
		log.Fatalf("Failed to initialize secret manager: %v", err)
	}

	config := &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://dev:devpassword@localhost:5432/musicapp?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          secretManager.GetJWTSecret(),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3BucketName:       getEnv("S3_BUCKET_NAME", ""),
		S3CDNURL:           getEnv("S3_CDN_URL", ""),
		Port:               getEnvAsInt("PORT", 8080),
		Environment:        getEnv("ENVIRONMENT", "development"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
