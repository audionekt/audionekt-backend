package service

import (
	"context"
	"fmt"

	"musicapp/internal/models"
	"musicapp/internal/repository"

	"github.com/google/uuid"
)

// FollowRepository interface for follow operations
type FollowRepositoryForFollow interface {
	Create(ctx context.Context, follow *models.Follow) error
	Delete(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) error
	IsFollowing(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) (bool, error)
	GetFollowers(ctx context.Context, followingType string, followingUserID, followingBandID *uuid.UUID, limit, offset int) ([]*models.Follow, error)
	GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*models.Follow, error)
}

// UserRepositoryForFollow interface for user operations
type UserRepositoryForFollow interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// BandRepositoryForFollow interface for band operations
type BandRepositoryForFollow interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error)
}

type FollowService struct {
	followRepo FollowRepositoryForFollow
	userRepo   UserRepositoryForFollow
	bandRepo   BandRepositoryForFollow
	cache      Cache
}

func NewFollowService(followRepo FollowRepositoryForFollow, userRepo UserRepositoryForFollow, bandRepo BandRepositoryForFollow, cache Cache) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
		bandRepo:   bandRepo,
		cache:      cache,
	}
}

// FollowUser follows a user
func (s *FollowService) FollowUser(ctx context.Context, followerID, followingUserID uuid.UUID) error {
	// Check if user is trying to follow themselves
	if followerID == followingUserID {
		return fmt.Errorf("you cannot follow yourself")
	}

	// Check if following user exists
	_, err := s.userRepo.GetByID(ctx, followingUserID)
	if err != nil {
		return fmt.Errorf("user to follow not found: %w", err)
	}

	// Check if already following
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, "user", &followingUserID, nil)
	if err != nil {
		return fmt.Errorf("failed to check follow status: %w", err)
	}

	if isFollowing {
		return fmt.Errorf("already following this user")
	}

	// Create follow relationship
	follow := &models.Follow{
		FollowerID:      followerID,
		FollowingType:   "user",
		FollowingUserID: &followingUserID,
		FollowingBandID: nil,
	}

	if err := s.followRepo.Create(ctx, follow); err != nil {
		return fmt.Errorf("failed to create follow relationship: %w", err)
	}

	return nil
}

// FollowBand follows a band
func (s *FollowService) FollowBand(ctx context.Context, followerID, followingBandID uuid.UUID) error {
	// Check if band exists
	_, err := s.bandRepo.GetByID(ctx, followingBandID)
	if err != nil {
		return fmt.Errorf("band to follow not found: %w", err)
	}

	// Check if already following
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, "band", nil, &followingBandID)
	if err != nil {
		return fmt.Errorf("failed to check follow status: %w", err)
	}

	if isFollowing {
		return fmt.Errorf("already following this band")
	}

	// Create follow relationship
	follow := &models.Follow{
		FollowerID:      followerID,
		FollowingType:   "band",
		FollowingUserID: nil,
		FollowingBandID: &followingBandID,
	}

	if err := s.followRepo.Create(ctx, follow); err != nil {
		return fmt.Errorf("failed to create follow relationship: %w", err)
	}

	return nil
}

// UnfollowUser unfollows a user
func (s *FollowService) UnfollowUser(ctx context.Context, followerID, followingUserID uuid.UUID) error {
	// Check if currently following
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, "user", &followingUserID, nil)
	if err != nil {
		return fmt.Errorf("failed to check follow status: %w", err)
	}

	if !isFollowing {
		return fmt.Errorf("not following this user")
	}

	// Delete follow relationship
	if err := s.followRepo.Delete(ctx, followerID, "user", &followingUserID, nil); err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	return nil
}

// UnfollowBand unfollows a band
func (s *FollowService) UnfollowBand(ctx context.Context, followerID, followingBandID uuid.UUID) error {
	// Check if currently following
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, "band", nil, &followingBandID)
	if err != nil {
		return fmt.Errorf("failed to check follow status: %w", err)
	}

	if !isFollowing {
		return fmt.Errorf("not following this band")
	}

	// Delete follow relationship
	if err := s.followRepo.Delete(ctx, followerID, "band", nil, &followingBandID); err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	return nil
}

// GetFollowers retrieves users who follow the specified user/band
func (s *FollowService) GetFollowers(ctx context.Context, followingType string, followingUserID, followingBandID *uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d (must be >= 0)", offset)
	}

	followers, err := s.followRepo.GetFollowers(ctx, followingType, followingUserID, followingBandID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve followers: %w", err)
	}

	return followers, nil
}

// GetFollowing retrieves users/bands that the specified user is following
func (s *FollowService) GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d (must be >= 0)", offset)
	}

	following, err := s.followRepo.GetFollowing(ctx, followerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve following: %w", err)
	}

	return following, nil
}

// IsFollowing checks if a user is following another user or band
func (s *FollowService) IsFollowing(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) (bool, error) {
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, followingType, followingUserID, followingBandID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %w", err)
	}

	return isFollowing, nil
}

// Adapter structs to bridge existing concrete types with interfaces

// FollowRepositoryAdapter adapts repository.FollowRepository to FollowRepositoryForFollow
type FollowRepositoryAdapter struct {
	repo *repository.FollowRepository
}

func NewFollowRepositoryAdapter(repo *repository.FollowRepository) FollowRepositoryForFollow {
	return &FollowRepositoryAdapter{repo: repo}
}

func (a *FollowRepositoryAdapter) Create(ctx context.Context, follow *models.Follow) error {
	return a.repo.Create(ctx, follow)
}

func (a *FollowRepositoryAdapter) Delete(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) error {
	return a.repo.Delete(ctx, followerID, followingType, followingUserID, followingBandID)
}

func (a *FollowRepositoryAdapter) IsFollowing(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) (bool, error) {
	return a.repo.IsFollowing(ctx, followerID, followingType, followingUserID, followingBandID)
}

func (a *FollowRepositoryAdapter) GetFollowers(ctx context.Context, followingType string, followingUserID, followingBandID *uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	return a.repo.GetFollowers(ctx, followingType, followingUserID, followingBandID, limit, offset)
}

func (a *FollowRepositoryAdapter) GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	return a.repo.GetFollowing(ctx, followerID, limit, offset)
}

// UserRepositoryForFollowAdapter adapts repository.UserRepository to UserRepositoryForFollow
type UserRepositoryForFollowAdapter struct {
	repo *repository.UserRepository
}

func NewUserRepositoryForFollowAdapter(repo *repository.UserRepository) UserRepositoryForFollow {
	return &UserRepositoryForFollowAdapter{repo: repo}
}

func (a *UserRepositoryForFollowAdapter) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return a.repo.GetByID(ctx, id)
}

// BandRepositoryForFollowAdapter adapts repository.BandRepository to BandRepositoryForFollow
type BandRepositoryForFollowAdapter struct {
	repo *repository.BandRepository
}

func NewBandRepositoryForFollowAdapter(repo *repository.BandRepository) BandRepositoryForFollow {
	return &BandRepositoryForFollowAdapter{repo: repo}
}

func (a *BandRepositoryForFollowAdapter) GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error) {
	return a.repo.GetByID(ctx, id)
}
