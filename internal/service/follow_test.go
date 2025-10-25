package service

import (
	"context"
	"errors"
	"testing"

	"musicapp/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for FollowService testing

type MockFollowRepositoryForFollow struct {
	createError      error
	deleteError      error
	isFollowingError error
	isFollowingResult bool
	getFollowersError error
	getFollowersResult []*models.Follow
	getFollowingError error
	getFollowingResult []*models.Follow
}

func (m *MockFollowRepositoryForFollow) Create(ctx context.Context, follow *models.Follow) error {
	return m.createError
}

func (m *MockFollowRepositoryForFollow) Delete(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) error {
	return m.deleteError
}

func (m *MockFollowRepositoryForFollow) IsFollowing(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) (bool, error) {
	return m.isFollowingResult, m.isFollowingError
}

func (m *MockFollowRepositoryForFollow) GetFollowers(ctx context.Context, followingType string, followingUserID, followingBandID *uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	return m.getFollowersResult, m.getFollowersError
}

func (m *MockFollowRepositoryForFollow) GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	return m.getFollowingResult, m.getFollowingError
}

type MockUserRepositoryForFollow struct {
	getByIDError error
	user         *models.User
}

func (m *MockUserRepositoryForFollow) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.user, m.getByIDError
}

type MockBandRepositoryForFollow struct {
	getByIDError error
	band         *models.Band
}

func (m *MockBandRepositoryForFollow) GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error) {
	return m.band, m.getByIDError
}


func TestNewFollowService(t *testing.T) {
	followRepo := &MockFollowRepositoryForFollow{}
	userRepo := &MockUserRepositoryForFollow{}
	bandRepo := &MockBandRepositoryForFollow{}
	cache := &MockCache{}

	service := NewFollowService(followRepo, userRepo, bandRepo, cache)

	assert.NotNil(t, service)
	assert.Equal(t, followRepo, service.followRepo)
	assert.Equal(t, userRepo, service.userRepo)
	assert.Equal(t, bandRepo, service.bandRepo)
	assert.Equal(t, cache, service.cache)
}

func TestFollowService_FollowUser(t *testing.T) {
	tests := []struct {
		name             string
		followerID       uuid.UUID
		followingUserID  uuid.UUID
		setupMocks       func(*MockFollowRepositoryForFollow, *MockUserRepositoryForFollow)
		expectedError    string
	}{
		{
			name:            "successful follow",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, userRepo *MockUserRepositoryForFollow) {
				userRepo.user = &models.User{ID: uuid.New()}
				followRepo.isFollowingResult = false
			},
			expectedError: "",
		},
		{
			name:            "cannot follow self",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, userRepo *MockUserRepositoryForFollow) {
				// Set same ID for both
			},
			expectedError: "you cannot follow yourself",
		},
		{
			name:            "user not found",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, userRepo *MockUserRepositoryForFollow) {
				userRepo.getByIDError = errors.New("user not found")
			},
			expectedError: "user to follow not found",
		},
		{
			name:            "already following",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, userRepo *MockUserRepositoryForFollow) {
				userRepo.user = &models.User{ID: uuid.New()}
				followRepo.isFollowingResult = true
			},
			expectedError: "already following this user",
		},
		{
			name:            "database create error",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, userRepo *MockUserRepositoryForFollow) {
				userRepo.user = &models.User{ID: uuid.New()}
				followRepo.isFollowingResult = false
				followRepo.createError = errors.New("database error")
			},
			expectedError: "failed to create follow relationship",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followRepo := &MockFollowRepositoryForFollow{}
			userRepo := &MockUserRepositoryForFollow{}
			bandRepo := &MockBandRepositoryForFollow{}
			cache := &MockCache{}

			// Handle special case for "cannot follow self"
			if tt.name == "cannot follow self" {
				tt.followingUserID = tt.followerID
			}

			tt.setupMocks(followRepo, userRepo)

			service := NewFollowService(followRepo, userRepo, bandRepo, cache)
			err := service.FollowUser(context.Background(), tt.followerID, tt.followingUserID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFollowService_FollowBand(t *testing.T) {
	tests := []struct {
		name             string
		followerID       uuid.UUID
		followingBandID  uuid.UUID
		setupMocks       func(*MockFollowRepositoryForFollow, *MockBandRepositoryForFollow)
		expectedError    string
	}{
		{
			name:            "successful follow",
			followerID:      uuid.New(),
			followingBandID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, bandRepo *MockBandRepositoryForFollow) {
				bandRepo.band = &models.Band{ID: uuid.New()}
				followRepo.isFollowingResult = false
			},
			expectedError: "",
		},
		{
			name:            "band not found",
			followerID:      uuid.New(),
			followingBandID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, bandRepo *MockBandRepositoryForFollow) {
				bandRepo.getByIDError = errors.New("band not found")
			},
			expectedError: "band to follow not found",
		},
		{
			name:            "already following",
			followerID:      uuid.New(),
			followingBandID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, bandRepo *MockBandRepositoryForFollow) {
				bandRepo.band = &models.Band{ID: uuid.New()}
				followRepo.isFollowingResult = true
			},
			expectedError: "already following this band",
		},
		{
			name:            "database create error",
			followerID:      uuid.New(),
			followingBandID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow, bandRepo *MockBandRepositoryForFollow) {
				bandRepo.band = &models.Band{ID: uuid.New()}
				followRepo.isFollowingResult = false
				followRepo.createError = errors.New("database error")
			},
			expectedError: "failed to create follow relationship",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followRepo := &MockFollowRepositoryForFollow{}
			userRepo := &MockUserRepositoryForFollow{}
			bandRepo := &MockBandRepositoryForFollow{}
			cache := &MockCache{}

			tt.setupMocks(followRepo, bandRepo)

			service := NewFollowService(followRepo, userRepo, bandRepo, cache)
			err := service.FollowBand(context.Background(), tt.followerID, tt.followingBandID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFollowService_UnfollowUser(t *testing.T) {
	tests := []struct {
		name             string
		followerID       uuid.UUID
		followingUserID  uuid.UUID
		setupMocks       func(*MockFollowRepositoryForFollow)
		expectedError    string
	}{
		{
			name:            "successful unfollow",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = true
			},
			expectedError: "",
		},
		{
			name:            "not following",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = false
			},
			expectedError: "not following this user",
		},
		{
			name:            "database delete error",
			followerID:      uuid.New(),
			followingUserID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = true
				followRepo.deleteError = errors.New("database error")
			},
			expectedError: "failed to unfollow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followRepo := &MockFollowRepositoryForFollow{}
			userRepo := &MockUserRepositoryForFollow{}
			bandRepo := &MockBandRepositoryForFollow{}
			cache := &MockCache{}

			tt.setupMocks(followRepo)

			service := NewFollowService(followRepo, userRepo, bandRepo, cache)
			err := service.UnfollowUser(context.Background(), tt.followerID, tt.followingUserID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFollowService_UnfollowBand(t *testing.T) {
	tests := []struct {
		name             string
		followerID       uuid.UUID
		followingBandID  uuid.UUID
		setupMocks       func(*MockFollowRepositoryForFollow)
		expectedError    string
	}{
		{
			name:            "successful unfollow",
			followerID:      uuid.New(),
			followingBandID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = true
			},
			expectedError: "",
		},
		{
			name:            "not following",
			followerID:      uuid.New(),
			followingBandID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = false
			},
			expectedError: "not following this band",
		},
		{
			name:            "database delete error",
			followerID:      uuid.New(),
			followingBandID: uuid.New(),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = true
				followRepo.deleteError = errors.New("database error")
			},
			expectedError: "failed to unfollow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followRepo := &MockFollowRepositoryForFollow{}
			userRepo := &MockUserRepositoryForFollow{}
			bandRepo := &MockBandRepositoryForFollow{}
			cache := &MockCache{}

			tt.setupMocks(followRepo)

			service := NewFollowService(followRepo, userRepo, bandRepo, cache)
			err := service.UnfollowBand(context.Background(), tt.followerID, tt.followingBandID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFollowService_GetFollowers(t *testing.T) {
	tests := []struct {
		name             string
		followingType    string
		followingUserID  *uuid.UUID
		followingBandID  *uuid.UUID
		limit            int
		offset           int
		setupMocks       func(*MockFollowRepositoryForFollow)
		expectedError    string
		expectedCount    int
	}{
		{
			name:            "successful retrieval",
			followingType:   "user",
			followingUserID: uuidPtr(uuid.New()),
			limit:           10,
			offset:          0,
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.getFollowersResult = []*models.Follow{
					{ID: uuid.New()},
					{ID: uuid.New()},
				}
			},
			expectedError: "",
			expectedCount:  2,
		},
		{
			name:          "invalid limit - too low",
			followingType: "user",
			limit:         0,
			offset:        0,
			setupMocks:    func(followRepo *MockFollowRepositoryForFollow) {},
			expectedError: "invalid limit",
		},
		{
			name:          "invalid limit - too high",
			followingType: "user",
			limit:         101,
			offset:        0,
			setupMocks:    func(followRepo *MockFollowRepositoryForFollow) {},
			expectedError: "invalid limit",
		},
		{
			name:          "invalid offset",
			followingType: "user",
			limit:         10,
			offset:        -1,
			setupMocks:    func(followRepo *MockFollowRepositoryForFollow) {},
			expectedError: "invalid offset",
		},
		{
			name:            "database error",
			followingType:   "user",
			followingUserID: uuidPtr(uuid.New()),
			limit:           10,
			offset:          0,
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.getFollowersError = errors.New("database error")
			},
			expectedError: "failed to retrieve followers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followRepo := &MockFollowRepositoryForFollow{}
			userRepo := &MockUserRepositoryForFollow{}
			bandRepo := &MockBandRepositoryForFollow{}
			cache := &MockCache{}

			tt.setupMocks(followRepo)

			service := NewFollowService(followRepo, userRepo, bandRepo, cache)
			followers, err := service.GetFollowers(context.Background(), tt.followingType, tt.followingUserID, tt.followingBandID, tt.limit, tt.offset)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, followers)
			} else {
				assert.NoError(t, err)
				assert.Len(t, followers, tt.expectedCount)
			}
		})
	}
}

func TestFollowService_GetFollowing(t *testing.T) {
	tests := []struct {
		name             string
		followerID       uuid.UUID
		limit            int
		offset           int
		setupMocks       func(*MockFollowRepositoryForFollow)
		expectedError    string
		expectedCount    int
	}{
		{
			name:       "successful retrieval",
			followerID: uuid.New(),
			limit:      10,
			offset:     0,
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.getFollowingResult = []*models.Follow{
					{ID: uuid.New()},
					{ID: uuid.New()},
					{ID: uuid.New()},
				}
			},
			expectedError: "",
			expectedCount:  3,
		},
		{
			name:       "invalid limit - too low",
			followerID: uuid.New(),
			limit:      0,
			offset:     0,
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {},
			expectedError: "invalid limit",
		},
		{
			name:       "invalid limit - too high",
			followerID: uuid.New(),
			limit:      101,
			offset:     0,
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {},
			expectedError: "invalid limit",
		},
		{
			name:       "invalid offset",
			followerID: uuid.New(),
			limit:      10,
			offset:     -1,
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {},
			expectedError: "invalid offset",
		},
		{
			name:       "database error",
			followerID: uuid.New(),
			limit:      10,
			offset:     0,
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.getFollowingError = errors.New("database error")
			},
			expectedError: "failed to retrieve following",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followRepo := &MockFollowRepositoryForFollow{}
			userRepo := &MockUserRepositoryForFollow{}
			bandRepo := &MockBandRepositoryForFollow{}
			cache := &MockCache{}

			tt.setupMocks(followRepo)

			service := NewFollowService(followRepo, userRepo, bandRepo, cache)
			following, err := service.GetFollowing(context.Background(), tt.followerID, tt.limit, tt.offset)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, following)
			} else {
				assert.NoError(t, err)
				assert.Len(t, following, tt.expectedCount)
			}
		})
	}
}

func TestFollowService_IsFollowing(t *testing.T) {
	tests := []struct {
		name             string
		followerID       uuid.UUID
		followingType    string
		followingUserID  *uuid.UUID
		followingBandID  *uuid.UUID
		setupMocks       func(*MockFollowRepositoryForFollow)
		expectedResult   bool
		expectedError    string
	}{
		{
			name:            "is following user",
			followerID:      uuid.New(),
			followingType:   "user",
			followingUserID: uuidPtr(uuid.New()),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = true
			},
			expectedResult: true,
			expectedError:  "",
		},
		{
			name:            "is not following user",
			followerID:      uuid.New(),
			followingType:   "user",
			followingUserID: uuidPtr(uuid.New()),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = false
			},
			expectedResult: false,
			expectedError:  "",
		},
		{
			name:            "is following band",
			followerID:      uuid.New(),
			followingType:   "band",
			followingBandID: uuidPtr(uuid.New()),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingResult = true
			},
			expectedResult: true,
			expectedError:  "",
		},
		{
			name:            "database error",
			followerID:      uuid.New(),
			followingType:   "user",
			followingUserID: uuidPtr(uuid.New()),
			setupMocks: func(followRepo *MockFollowRepositoryForFollow) {
				followRepo.isFollowingError = errors.New("database error")
			},
			expectedResult: false,
			expectedError:  "failed to check follow status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followRepo := &MockFollowRepositoryForFollow{}
			userRepo := &MockUserRepositoryForFollow{}
			bandRepo := &MockBandRepositoryForFollow{}
			cache := &MockCache{}

			tt.setupMocks(followRepo)

			service := NewFollowService(followRepo, userRepo, bandRepo, cache)
			result, err := service.IsFollowing(context.Background(), tt.followerID, tt.followingType, tt.followingUserID, tt.followingBandID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, tt.expectedResult, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

// Helper function for UUID pointers
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}
