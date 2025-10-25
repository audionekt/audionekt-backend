package service

import (
	"context"
	"fmt"
	"time"

	"musicapp/internal/cache"
	"musicapp/internal/errors"
	"musicapp/internal/middleware"
	"musicapp/internal/models"
	"musicapp/internal/repository"
	"musicapp/pkg/utils"

	"github.com/google/uuid"
)

// Interfaces define contracts for dependencies
// This follows Go's "Accept interfaces, return structs" principle
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

type Cache interface {
	SetSession(ctx context.Context, userID string, data interface{}, expiration time.Duration) error
	DeleteSession(ctx context.Context, userID string) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	AddToBlacklist(ctx context.Context, jti string, expiration time.Duration) error
}

type AuthMiddleware interface {
	GenerateToken(userID, username string) (string, error)
	ValidateToken(tokenString string) (*middleware.Claims, error)
}

// AuthService now depends on interfaces, not concrete types
// This makes it much easier to test and more flexible
type AuthService struct {
	userRepo       UserRepository
	cache          Cache
	authMiddleware AuthMiddleware
}

// NewAuthService creates a new AuthService with dependency injection
// This follows the dependency injection pattern for better testability
func NewAuthService(userRepo UserRepository, cache Cache, authMiddleware AuthMiddleware) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		cache:          cache,
		authMiddleware: authMiddleware,
	}
}

// RegisterUser registers a new user and returns JWT token
func (s *AuthService) RegisterUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, string, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, "", errors.NewUserAlreadyExists("email", req.Email)
	}

	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, "", errors.NewUserAlreadyExists("username", req.Username)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", errors.Wrap(err, errors.ErrCodeInternalError, "Failed to hash password")
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
		return nil, "", errors.Wrap(err, errors.ErrCodeDatabaseError, "Failed to create user")
	}

	// Generate JWT token
	token, err := s.authMiddleware.GenerateToken(user.ID.String(), user.Username)
	if err != nil {
		return nil, "", errors.Wrap(err, errors.ErrCodeInternalError, "Failed to generate token")
	}

	// Store session in Redis
	sessionData := map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	}
	if err := s.cache.SetSession(ctx, user.ID.String(), sessionData, 24*time.Hour); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return user, token, nil
}

// LoginUser authenticates user and returns JWT token
func (s *AuthService) LoginUser(ctx context.Context, email, password string) (*models.User, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, "", errors.ErrInvalidCredentials
	}

	// Check password
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", errors.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.authMiddleware.GenerateToken(user.ID.String(), user.Username)
	if err != nil {
		return nil, "", errors.Wrap(err, errors.ErrCodeInternalError, "Failed to generate token")
	}

	// Store session in Redis
	sessionData := map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	}
	if err := s.cache.SetSession(ctx, user.ID.String(), sessionData, 24*time.Hour); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return user, token, nil
}

// LogoutUser invalidates JWT token
func (s *AuthService) LogoutUser(ctx context.Context, jti, userID string) error {
	// Add token to blacklist (24 hours from now)
	if err := s.cache.AddToBlacklist(ctx, jti, 24*time.Hour); err != nil {
		return errors.Wrap(err, errors.ErrCodeRedisError, "Failed to blacklist token")
	}

	// Clear session
	if err := s.cache.DeleteSession(ctx, userID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return nil
}

// ValidateToken validates JWT token and returns user info
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	// Parse and validate the token
	claims, err := s.authMiddleware.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token is blacklisted
	isBlacklisted, err := s.cache.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Get user from database
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// Adapter structs implement interfaces for existing concrete types
// This allows us to use the existing repository/cache/middleware with the new interface-based service

// UserRepositoryAdapter adapts repository.UserRepository to UserRepository interface
type UserRepositoryAdapter struct {
	repo *repository.UserRepository
}

func NewUserRepositoryAdapter(repo *repository.UserRepository) UserRepository {
	return &UserRepositoryAdapter{repo: repo}
}

func (a *UserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return a.repo.GetByEmail(ctx, email)
}

func (a *UserRepositoryAdapter) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return a.repo.GetByUsername(ctx, username)
}

func (a *UserRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *UserRepositoryAdapter) Create(ctx context.Context, user *models.User) error {
	return a.repo.Create(ctx, user)
}

// CacheAdapter adapts cache.Cache to Cache interface
type CacheAdapter struct {
	cache *cache.Cache
}

func NewCacheAdapter(cache *cache.Cache) Cache {
	return &CacheAdapter{cache: cache}
}

func (a *CacheAdapter) SetSession(ctx context.Context, userID string, data interface{}, expiration time.Duration) error {
	return a.cache.SetSession(ctx, userID, data, expiration)
}

func (a *CacheAdapter) DeleteSession(ctx context.Context, userID string) error {
	return a.cache.DeleteSession(ctx, userID)
}

func (a *CacheAdapter) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return a.cache.IsBlacklisted(ctx, jti)
}

func (a *CacheAdapter) AddToBlacklist(ctx context.Context, jti string, expiration time.Duration) error {
	return a.cache.AddToBlacklist(ctx, jti, expiration)
}

// AuthMiddlewareAdapter adapts middleware.AuthMiddleware to AuthMiddleware interface
type AuthMiddlewareAdapter struct {
	middleware *middleware.AuthMiddleware
}

func NewAuthMiddlewareAdapter(middleware *middleware.AuthMiddleware) AuthMiddleware {
	return &AuthMiddlewareAdapter{middleware: middleware}
}

func (a *AuthMiddlewareAdapter) GenerateToken(userID, username string) (string, error) {
	return a.middleware.GenerateToken(userID, username)
}

func (a *AuthMiddlewareAdapter) ValidateToken(tokenString string) (*middleware.Claims, error) {
	return a.middleware.ValidateToken(tokenString)
}
