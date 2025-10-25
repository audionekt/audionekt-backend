package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"musicapp/internal/interfaces"
	"musicapp/internal/logging"
	"musicapp/internal/models"
	"musicapp/internal/repository"
	"musicapp/internal/storage"
	"musicapp/pkg/utils"

	"github.com/google/uuid"
)

// UserService now depends on interfaces, not concrete types
// This makes it much easier to test and more flexible
type UserService struct {
	userRepo interfaces.UserRepositoryExtended
	cache    interfaces.Cache
	s3Client interfaces.S3Client
	logger   *logging.Logger
}

// NewUserService creates a new UserService with dependency injection
// This follows the dependency injection pattern for better testability
func NewUserService(userRepo interfaces.UserRepositoryExtended, cache interfaces.Cache, s3Client interfaces.S3Client, logger *logging.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		cache:    cache,
		s3Client: s3Client,
		logger:   logger,
	}
}

// CreateUser creates a new user with hashed password
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username %s already taken", req.Username)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Location:     req.Location,
		City:         req.City,
		Country:      req.Country,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// UpdateUser updates user profile
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.Location != nil {
		user.Location = req.Location
	}
	if req.City != nil {
		user.City = req.City
	}
	if req.Country != nil {
		user.Country = req.Country
	}
	if req.Genres != nil {
		user.Genres = req.Genres
	}
	if req.Skills != nil {
		user.Skills = req.Skills
	}
	if req.SpotifyURL != nil {
		user.SpotifyURL = req.SpotifyURL
	}
	if req.SoundcloudURL != nil {
		user.SoundcloudURL = req.SoundcloudURL
	}
	if req.InstagramHandle != nil {
		user.InstagramHandle = req.InstagramHandle
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// GetNearbyUsers finds users within a specified radius
func (s *UserService) GetNearbyUsers(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.User, error) {
	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("invalid latitude: %f", lat)
	}
	if lng < -180 || lng > 180 {
		return nil, fmt.Errorf("invalid longitude: %f", lng)
	}
	if radiusKm <= 0 || radiusKm > 500 {
		return nil, fmt.Errorf("invalid radius: %d km (must be 1-500)", radiusKm)
	}
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}

	users, err := s.userRepo.GetNearby(ctx, lat, lng, radiusKm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby users: %w", err)
	}

	return users, nil
}

// GetUserPosts retrieves posts by a specific user
func (s *UserService) GetUserPosts(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d (must be >= 0)", offset)
	}

	// This would need to be implemented in the repository
	// For now, return empty slice
	return []*models.Post{}, nil
}

// GetFollowers retrieves users who follow the specified user
func (s *UserService) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d (must be >= 0)", offset)
	}

	followers, err := s.userRepo.GetFollowers(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}

	return followers, nil
}

// GetFollowing retrieves users that the specified user is following
func (s *UserService) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d (must be >= 0)", offset)
	}

	following, err := s.userRepo.GetFollowing(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}

	return following, nil
}

// UploadProfilePicture uploads a profile picture to S3
func (s *UserService) UploadProfilePicture(ctx context.Context, userID uuid.UUID, filename string, fileData []byte) (string, error) {
	start := time.Now()
	logger := s.logger.WithUserID(userID.String()).WithOperation("upload_profile_picture")
	
	if s.s3Client == nil {
		logger.Error("S3 client not configured")
		return "", fmt.Errorf("S3 client not configured")
	}

	logger.WithField("filename", filename).WithField("file_size", len(fileData)).Info("Starting profile picture upload")

	// Validate image file
	if err := s.s3Client.ValidateImageFile(filename, int64(len(fileData))); err != nil {
		logger.WithError(err).Error("Image validation failed")
		return "", fmt.Errorf("invalid image file: %w", err)
	}

	logger.Debug("Image validation passed, attempting S3 upload")

	// Upload to S3
	uploadResult, err := s.s3Client.UploadImage(ctx, userID.String(), filename, bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		logger.WithError(err).Error("S3 upload failed")
		return "", fmt.Errorf("failed to upload profile picture: %w", err)
	}

	logger.WithField("upload_url", uploadResult.URL).WithDuration(time.Since(start)).Info("S3 upload successful")

	// Update user profile picture URL
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		logger.WithError(err).Error("User not found during profile picture update")
		return "", fmt.Errorf("user not found: %w", err)
	}

	user.ProfilePictureURL = &uploadResult.URL
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.WithError(err).Error("Failed to update user profile picture URL")
		return "", fmt.Errorf("failed to update profile picture URL: %w", err)
	}

	logger.WithDuration(time.Since(start)).Info("Profile picture upload completed successfully")
	return uploadResult.URL, nil
}

// GetUserBands gets all bands that a user is a member of
func (s *UserService) GetUserBands(ctx context.Context, userID uuid.UUID) ([]*models.BandMember, error) {
	// This would need to be implemented in the user service
	// For now, we'll delegate to the band service
	// In a real implementation, you might want to add this to the user repository
	return nil, fmt.Errorf("not implemented - use band service instead")
}

// GetAllUsers gets all users with pagination
func (s *UserService) GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	// Default pagination values
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit to prevent abuse
	}
	if offset < 0 {
		offset = 0
	}

	return s.userRepo.GetAll(ctx, limit, offset)
}

// Extended adapter for UserService (building on existing adapters from auth.go)
type UserRepositoryExtendedAdapter struct {
	*UserRepositoryAdapter
}

func NewUserRepositoryExtendedAdapter(repo *repository.UserRepository) interfaces.UserRepositoryExtended {
	baseAdapter := NewUserRepositoryAdapter(repo)
	return &UserRepositoryExtendedAdapter{
		UserRepositoryAdapter: baseAdapter.(*UserRepositoryAdapter),
	}
}

func (a *UserRepositoryExtendedAdapter) Update(ctx context.Context, user *models.User) error {
	return a.repo.Update(ctx, user)
}

func (a *UserRepositoryExtendedAdapter) GetNearby(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.User, error) {
	return a.repo.GetNearby(ctx, lat, lng, radiusKm, limit)
}

func (a *UserRepositoryExtendedAdapter) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	return a.repo.GetFollowers(ctx, userID, limit, offset)
}

func (a *UserRepositoryExtendedAdapter) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	return a.repo.GetFollowing(ctx, userID, limit, offset)
}

func (a *UserRepositoryExtendedAdapter) GetAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return a.repo.GetAll(ctx, limit, offset)
}

// S3ClientAdapter adapts storage.S3Client to S3Client interface
type S3ClientAdapter struct {
	s3Client *storage.S3Client
}

func NewS3ClientAdapter(s3Client *storage.S3Client) interfaces.S3Client {
	return &S3ClientAdapter{s3Client: s3Client}
}

func (a *S3ClientAdapter) ValidateImageFile(filename string, size int64) error {
	return a.s3Client.ValidateImageFile(filename, size)
}

func (a *S3ClientAdapter) UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error) {
	return a.s3Client.UploadImage(ctx, userID, filename, file, size)
}
