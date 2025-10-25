package interfaces

import (
	"context"
	"io"
	"time"

	"musicapp/internal/middleware"
	"musicapp/internal/models"
	"musicapp/internal/storage"

	"github.com/google/uuid"
)

// Cache defines the interface for caching operations
type Cache interface {
	SetSession(ctx context.Context, userID string, data interface{}, expiration time.Duration) error
	DeleteSession(ctx context.Context, userID string) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	AddToBlacklist(ctx context.Context, jti string, expiration time.Duration) error
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

// UserRepositoryExtended extends UserRepository with additional operations
type UserRepositoryExtended interface {
	UserRepository
	Update(ctx context.Context, user *models.User) error
	GetNearby(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.User, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.User, error)
}

// AuthMiddleware defines the interface for authentication operations
type AuthMiddleware interface {
	GenerateToken(userID, username string) (string, error)
	ValidateToken(tokenString string) (*middleware.Claims, error)
}

// S3Client defines the interface for S3 operations
type S3Client interface {
	ValidateImageFile(filename string, size int64) error
	UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error)
}
