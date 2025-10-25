package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"musicapp/internal/models"
	"musicapp/internal/storage"
	"musicapp/pkg/utils"

	"github.com/google/uuid"
)

// Extended MockUserRepository with additional methods for UserService
type ExtendedMockUserRepository struct {
	*MockUserRepository
	updateError   error
	getByIDError  error
	getNearbyError error // Added for GetNearby errors
	getFollowersError error // Added for GetFollowers errors
	getFollowingError error // Added for GetFollowing errors
	nearbyUsers   []*models.User
	followers     []*models.User
	following     []*models.User
	allUsers      []*models.User
}

func NewExtendedMockUserRepository() *ExtendedMockUserRepository {
	return &ExtendedMockUserRepository{
		MockUserRepository: NewMockUserRepository(),
	}
}

// Override GetByID to use getByIDError field
func (m *ExtendedMockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	if user, exists := m.usersByID[id.String()]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *ExtendedMockUserRepository) Update(ctx context.Context, user *models.User) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.usersByID[user.ID.String()] = user
	return nil
}

func (m *ExtendedMockUserRepository) GetNearby(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.User, error) {
	if m.getNearbyError != nil {
		return nil, m.getNearbyError
	}
	return m.nearbyUsers, nil
}

func (m *ExtendedMockUserRepository) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	if m.getFollowersError != nil {
		return nil, m.getFollowersError
	}
	return m.followers, nil
}

func (m *ExtendedMockUserRepository) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	if m.getFollowingError != nil {
		return nil, m.getFollowingError
	}
	return m.following, nil
}

func (m *ExtendedMockUserRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return m.allUsers, nil
}

// TestableUserService allows dependency injection for testing
type TestableUserService struct {
	userRepo UserRepositoryExtended
	cache    Cache
	s3Client S3Client
}

func NewTestableUserService(userRepo UserRepositoryExtended, cache Cache, s3Client S3Client) *TestableUserService {
	return &TestableUserService{
		userRepo: userRepo,
		cache:    cache,
		s3Client: s3Client,
	}
}

// CreateUser implements the same business logic as UserService.CreateUser
func (s *TestableUserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
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

// GetUser implements the same business logic as UserService.GetUser
func (s *TestableUserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

type MockS3Client struct {
	validateError error
	uploadError   error
	uploadResult  *storage.UploadResult
}

func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		uploadResult: &storage.UploadResult{
			URL: "https://example.com/profile.jpg",
		},
	}
}

func (m *MockS3Client) ValidateImageFile(filename string, size int64) error {
	return m.validateError
}

func (m *MockS3Client) UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error) {
	if m.uploadError != nil {
		return nil, m.uploadError
	}
	return m.uploadResult, nil
}

// Test CreateUser business logic with the REAL UserService using mocks
func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		req            *models.CreateUserRequest
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectUser     bool
	}{
		{
			name: "successful user creation",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Location: &models.Location{Latitude: 40.7128, Longitude: -74.0060},
				City:     stringPtr("Test City"),
				Country:  stringPtr("Test Country"),
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// No existing users, no errors
			},
			expectError: false,
			expectUser:  true,
		},
		{
			name: "user already exists by email",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add existing user by email
				existingUser := &models.User{
					ID:    uuid.New(),
					Email: "existing@example.com",
				}
				userRepo.usersByEmail["existing@example.com"] = existingUser
			},
			expectError:   true,
			errorContains: "user with email existing@example.com already exists",
		},
		{
			name: "username already taken",
			req: &models.CreateUserRequest{
				Username: "existinguser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add existing user by username
				existingUser := &models.User{
					ID:       uuid.New(),
					Username: "existinguser",
				}
				userRepo.usersByUsername["existinguser"] = existingUser
			},
			expectError:   true,
			errorContains: "username existinguser already taken",
		},
		{
			name: "user creation fails",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.createError = fmt.Errorf("database error")
			},
			expectError:   true,
			errorContains: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// Create REAL UserService with mocks - this is the key difference!
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test CreateUser
			user, err := userService.CreateUser(context.Background(), tt.req)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if user != nil {
					t.Error("Expected user to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUser && user == nil {
					t.Error("Expected user but got nil")
				}
				
				// Verify user was created with correct data
				if user != nil {
					if user.Username != tt.req.Username {
						t.Errorf("Expected username %s, got %s", tt.req.Username, user.Username)
					}
					if user.Email != tt.req.Email {
						t.Errorf("Expected email %s, got %s", tt.req.Email, user.Email)
					}
					if user.PasswordHash == "" {
						t.Error("Expected password to be hashed")
					}
				}
			}
		})
	}
}

// Test GetUser business logic with the REAL UserService using mocks
func TestUserService_GetUser(t *testing.T) {
	t.Run("successful user retrieval", func(t *testing.T) {
		// Setup mocks
		userRepo := NewExtendedMockUserRepository()
		cache := NewMockCache()
		s3Client := NewMockS3Client()
		
		// Add user to repository
		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
		}
		userRepo.usersByID[userID.String()] = user
		
		// Create REAL UserService with mocks
		userService := NewUserService(userRepo, cache, s3Client)
		
		// Test GetUser
		resultUser, err := userService.GetUser(context.Background(), userID)
		
		// Verify results
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
		if resultUser == nil {
			t.Error("Expected user but got nil")
		}
		if resultUser != nil && resultUser.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID.String(), resultUser.ID.String())
		}
	})
	
	t.Run("user not found", func(t *testing.T) {
		// Setup mocks
		userRepo := NewExtendedMockUserRepository()
		cache := NewMockCache()
		s3Client := NewMockS3Client()
		
		// No user in repository
		
		// Create REAL UserService with mocks
		userService := NewUserService(userRepo, cache, s3Client)
		
		// Test GetUser
		userID := uuid.New()
		user, err := userService.GetUser(context.Background(), userID)
		
		// Verify results
		if err == nil {
			t.Error("Expected error but got none")
		}
		if !strings.Contains(err.Error(), "user not found") {
			t.Errorf("Expected error to contain 'user not found', got '%s'", err.Error())
		}
		if user != nil {
			t.Error("Expected user to be nil on error")
		}
	})
	
	t.Run("database error", func(t *testing.T) {
		// Setup mocks
		userRepo := NewExtendedMockUserRepository()
		cache := NewMockCache()
		s3Client := NewMockS3Client()
		
		userRepo.getByIDError = fmt.Errorf("database connection error")
		
		// Create REAL UserService with mocks
		userService := NewUserService(userRepo, cache, s3Client)
		
		// Test GetUser
		userID := uuid.New()
		user, err := userService.GetUser(context.Background(), userID)
		
		// Verify results
		if err == nil {
			t.Error("Expected error but got none")
		}
		if !strings.Contains(err.Error(), "user not found") {
			t.Errorf("Expected error to contain 'user not found', got '%s'", err.Error())
		}
		if user != nil {
			t.Error("Expected user to be nil on error")
		}
	})
}

// Test UploadProfilePicture business logic with the REAL UserService using mocks
func TestUserService_UploadProfilePicture(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		filename       string
		fileData       []byte
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectURL      bool
		expectedURL    string
	}{
		{
			name:     "successful profile picture upload",
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add user
				userID := uuid.New()
				user := &models.User{
					ID:       userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[userID.String()] = user
				
				// Setup S3 mock to return success
				s3Client.uploadResult = &storage.UploadResult{
					URL:      "https://example.com/profile.jpg",
					Key:      "test-key",
					Size:     1024,
					MimeType: "image/jpeg",
				}
			},
			expectError: false,
			expectURL:   true,
			expectedURL: "https://example.com/profile.jpg",
		},
		{
			name:     "S3 client not configured",
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Don't set up S3Client - leave it nil
			},
			expectError:   true,
			errorContains: "S3 client not configured",
		},
		{
			name:     "invalid image file",
			userID:   uuid.New(),
			filename: "profile.txt", // Wrong file type
			fileData: []byte("not an image"),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Setup S3 mock to return validation error
				s3Client.validateError = fmt.Errorf("invalid file type")
			},
			expectError:   true,
			errorContains: "invalid image file",
		},
		{
			name:     "S3 upload failure",
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add user
				userID := uuid.New()
				user := &models.User{
					ID:       userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[userID.String()] = user
				
				// Setup S3 mock to return upload error
				s3Client.uploadError = fmt.Errorf("S3 upload failed")
			},
			expectError:   true,
			errorContains: "failed to upload profile picture",
		},
		{
			name:     "user not found",
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Don't add user - will cause "user not found" error
				
				// Setup S3 mock to return success
				s3Client.uploadResult = &storage.UploadResult{
					URL:      "https://example.com/profile.jpg",
					Key:      "test-key",
					Size:     1024,
					MimeType: "image/jpeg",
				}
			},
			expectError:   true,
			errorContains: "user not found",
		},
		{
			name:     "database update failure",
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add user
				userID := uuid.New()
				user := &models.User{
					ID:       userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[userID.String()] = user
				
				// Setup S3 mock to return success
				s3Client.uploadResult = &storage.UploadResult{
					URL:      "https://example.com/profile.jpg",
					Key:      "test-key",
					Size:     1024,
					MimeType: "image/jpeg",
				}
				
				// Setup repository to return update error
				userRepo.updateError = fmt.Errorf("database update failed")
			},
			expectError:   true,
			errorContains: "failed to update profile picture URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// For successful tests, add the user with the correct ID
			if !tt.expectError && tt.expectURL {
				user := &models.User{
					ID:       tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[tt.userID.String()] = user
			}
			
			// For database update failure test, add user but set update error
			if tt.name == "database update failure" {
				user := &models.User{
					ID:       tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[tt.userID.String()] = user
			}
			
			// Create REAL UserService with mocks
			var userService *UserService
			if tt.name == "S3 client not configured" {
				userService = NewUserService(userRepo, cache, nil) // Pass nil for S3Client
			} else {
				userService = NewUserService(userRepo, cache, s3Client)
			}
			
			// Test UploadProfilePicture
			url, err := userService.UploadProfilePicture(context.Background(), tt.userID, tt.filename, tt.fileData)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if url != "" {
					t.Error("Expected URL to be empty on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectURL && url == "" {
					t.Error("Expected URL but got empty string")
				}
				if tt.expectedURL != "" && url != tt.expectedURL {
					t.Errorf("Expected URL %s, got %s", tt.expectedURL, url)
				}
			}
		})
	}
}

// Test GetUserBands business logic with the REAL UserService using mocks
func TestUserService_GetUserBands(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectBands    bool
		expectedCount  int
	}{
		{
			name:   "not implemented error",
			userID: uuid.New(),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// No setup needed - method always returns error
			},
			expectError:   true,
			errorContains: "not implemented",
		},
		{
			name:   "different user ID still returns not implemented",
			userID: uuid.New(),
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// No setup needed - method always returns error
			},
			expectError:   true,
			errorContains: "not implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// Create REAL UserService with mocks
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test GetUserBands
			bands, err := userService.GetUserBands(context.Background(), tt.userID)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if bands != nil {
					t.Error("Expected bands to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectBands && bands == nil {
					t.Error("Expected bands but got nil")
				}
				if tt.expectBands && len(bands) != tt.expectedCount {
					t.Errorf("Expected %d bands, got %d", tt.expectedCount, len(bands))
				}
			}
		})
	}
}

// Test GetUserPosts business logic with the REAL UserService using mocks
func TestUserService_GetUserPosts(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		limit          int
		offset         int
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectPosts    bool
		expectedCount  int
	}{
		{
			name:   "successful user posts retrieval",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// No setup needed - method returns empty slice
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 0, // Method returns empty slice
		},
		{
			name:          "invalid limit too low",
			userID:        uuid.New(),
			limit:         0,
			offset:        0,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid limit too high",
			userID:        uuid.New(),
			limit:         101,
			offset:        0,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid offset negative",
			userID:        uuid.New(),
			limit:         20,
			offset:        -1,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid offset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// Create REAL UserService with mocks
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test GetUserPosts
			posts, err := userService.GetUserPosts(context.Background(), tt.userID, tt.limit, tt.offset)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if posts != nil {
					t.Error("Expected posts to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectPosts && posts == nil {
					t.Error("Expected posts but got nil")
				}
				if tt.expectPosts && len(posts) != tt.expectedCount {
					t.Errorf("Expected %d posts, got %d", tt.expectedCount, len(posts))
				}
			}
		})
	}
}

// Test GetAllUsers business logic with the REAL UserService using mocks
func TestUserService_GetAllUsers(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		offset         int
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectUsers    bool
		expectedCount  int
	}{
		{
			name:   "successful all users retrieval",
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add users
				userRepo.allUsers = []*models.User{
					{ID: uuid.New(), Username: "user1", Email: "user1@example.com"},
					{ID: uuid.New(), Username: "user2", Email: "user2@example.com"},
					{ID: uuid.New(), Username: "user3", Email: "user3@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 3,
		},
		{
			name:   "default limit when zero",
			limit:  0,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.allUsers = []*models.User{
					{ID: uuid.New(), Username: "user1", Email: "user1@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 1,
		},
		{
			name:   "limit capped at 100",
			limit:  150,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.allUsers = []*models.User{
					{ID: uuid.New(), Username: "user1", Email: "user1@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 1,
		},
		{
			name:   "negative offset becomes zero",
			limit:  20,
			offset: -5,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.allUsers = []*models.User{
					{ID: uuid.New(), Username: "user1", Email: "user1@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 1,
		},
		{
			name:   "no users found",
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.allUsers = []*models.User{} // Empty slice
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// Create REAL UserService with mocks
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test GetAllUsers
			users, err := userService.GetAllUsers(context.Background(), tt.limit, tt.offset)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if users != nil {
					t.Error("Expected users to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUsers && users == nil {
					t.Error("Expected users but got nil")
				}
				if tt.expectUsers && len(users) != tt.expectedCount {
					t.Errorf("Expected %d users, got %d", tt.expectedCount, len(users))
				}
			}
		})
	}
}

// Test GetFollowing business logic with the REAL UserService using mocks
func TestUserService_GetFollowing(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		limit          int
		offset         int
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectUsers    bool
		expectedCount  int
	}{
		{
			name:   "successful following retrieval",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add following users
				userRepo.following = []*models.User{
					{ID: uuid.New(), Username: "following1", Email: "following1@example.com"},
					{ID: uuid.New(), Username: "following2", Email: "following2@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 2,
		},
		{
			name:          "invalid limit too low",
			userID:        uuid.New(),
			limit:         0,
			offset:        0,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid limit too high",
			userID:        uuid.New(),
			limit:         101,
			offset:        0,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid offset negative",
			userID:        uuid.New(),
			limit:         20,
			offset:        -1,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid offset",
		},
		{
			name:   "database error",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.getFollowingError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to get following",
		},
		{
			name:   "no following found",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.following = []*models.User{} // Empty slice
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// Create REAL UserService with mocks
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test GetFollowing
			users, err := userService.GetFollowing(context.Background(), tt.userID, tt.limit, tt.offset)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if users != nil {
					t.Error("Expected users to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUsers && users == nil {
					t.Error("Expected users but got nil")
				}
				if tt.expectUsers && len(users) != tt.expectedCount {
					t.Errorf("Expected %d users, got %d", tt.expectedCount, len(users))
				}
			}
		})
	}
}

// Test GetFollowers business logic with the REAL UserService using mocks
func TestUserService_GetFollowers(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		limit          int
		offset         int
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectUsers    bool
		expectedCount  int
	}{
		{
			name:   "successful followers retrieval",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add followers
				userRepo.followers = []*models.User{
					{ID: uuid.New(), Username: "follower1", Email: "follower1@example.com"},
					{ID: uuid.New(), Username: "follower2", Email: "follower2@example.com"},
					{ID: uuid.New(), Username: "follower3", Email: "follower3@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 3,
		},
		{
			name:          "invalid limit too low",
			userID:        uuid.New(),
			limit:         0,
			offset:        0,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid limit too high",
			userID:        uuid.New(),
			limit:         101,
			offset:        0,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid offset negative",
			userID:        uuid.New(),
			limit:         20,
			offset:        -1,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid offset",
		},
		{
			name:   "database error",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.getFollowersError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to get followers",
		},
		{
			name:   "no followers found",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.followers = []*models.User{} // Empty slice
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 0,
		},
		{
			name:   "pagination test",
			userID: uuid.New(),
			limit:  10,
			offset: 5,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add followers
				userRepo.followers = []*models.User{
					{ID: uuid.New(), Username: "follower1", Email: "follower1@example.com"},
					{ID: uuid.New(), Username: "follower2", Email: "follower2@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// Create REAL UserService with mocks
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test GetFollowers
			users, err := userService.GetFollowers(context.Background(), tt.userID, tt.limit, tt.offset)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if users != nil {
					t.Error("Expected users to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUsers && users == nil {
					t.Error("Expected users but got nil")
				}
				if tt.expectUsers && len(users) != tt.expectedCount {
					t.Errorf("Expected %d users, got %d", tt.expectedCount, len(users))
				}
			}
		})
	}
}

// Test GetNearbyUsers business logic with the REAL UserService using mocks
func TestUserService_GetNearbyUsers(t *testing.T) {
	tests := []struct {
		name           string
		lat            float64
		lng            float64
		radiusKm       int
		limit          int
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectUsers    bool
		expectedCount  int
	}{
		{
			name:      "successful nearby users retrieval",
			lat:       40.7128,
			lng:       -74.0060,
			radiusKm:  10,
			limit:     20,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// Add nearby users
				userRepo.nearbyUsers = []*models.User{
					{ID: uuid.New(), Username: "user1", Email: "user1@example.com"},
					{ID: uuid.New(), Username: "user2", Email: "user2@example.com"},
				}
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 2,
		},
		{
			name:          "invalid latitude too low",
			lat:           -91.0,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid latitude",
		},
		{
			name:          "invalid latitude too high",
			lat:           91.0,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid latitude",
		},
		{
			name:          "invalid longitude too low",
			lat:           40.7128,
			lng:           -181.0,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid longitude",
		},
		{
			name:          "invalid longitude too high",
			lat:           40.7128,
			lng:           181.0,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid longitude",
		},
		{
			name:          "invalid radius too low",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      0,
			limit:         20,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid radius",
		},
		{
			name:          "invalid radius too high",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      501,
			limit:         20,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid radius",
		},
		{
			name:          "invalid limit too low",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         0,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid limit too high",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         101,
			setupMocks:    func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:      "database error",
			lat:       40.7128,
			lng:       -74.0060,
			radiusKm:  10,
			limit:     20,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.getNearbyError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to get nearby users",
		},
		{
			name:      "no nearby users found",
			lat:       40.7128,
			lng:       -74.0060,
			radiusKm:  10,
			limit:     20,
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.nearbyUsers = []*models.User{} // Empty slice
			},
			expectError:   false,
			expectUsers:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// Create REAL UserService with mocks
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test GetNearbyUsers
			users, err := userService.GetNearbyUsers(context.Background(), tt.lat, tt.lng, tt.radiusKm, tt.limit)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if users != nil {
					t.Error("Expected users to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUsers && users == nil {
					t.Error("Expected users but got nil")
				}
				if tt.expectUsers && len(users) != tt.expectedCount {
					t.Errorf("Expected %d users, got %d", tt.expectedCount, len(users))
				}
			}
		})
	}
}

// Test UpdateUser business logic with the REAL UserService using mocks
func TestUserService_UpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		req            *models.UpdateUserRequest
		setupMocks     func(*ExtendedMockUserRepository, *MockCache, *MockS3Client)
		expectError    bool
		errorContains  string
		expectUser     bool
	}{
		{
			name:   "successful user update",
			userID: uuid.New(),
			req: &models.UpdateUserRequest{
				DisplayName: stringPtr("Updated Name"),
				Bio:         stringPtr("Updated bio"),
				City:        stringPtr("Updated City"),
				Country:     stringPtr("Updated Country"),
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// No setup needed - we'll add the user in the test
			},
			expectError: false,
			expectUser:  true,
		},
		{
			name:   "user not found",
			userID: uuid.New(),
			req: &models.UpdateUserRequest{
				DisplayName: stringPtr("Updated Name"),
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// No user in repository
			},
			expectError:   true,
			errorContains: "user not found",
		},
		{
			name:   "database error on get",
			userID: uuid.New(),
			req: &models.UpdateUserRequest{
				DisplayName: stringPtr("Updated Name"),
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.getByIDError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "user not found",
		},
		{
			name:   "database error on update",
			userID: uuid.New(),
			req: &models.UpdateUserRequest{
				DisplayName: stringPtr("Updated Name"),
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				userRepo.updateError = fmt.Errorf("database update error")
			},
			expectError:   true,
			errorContains: "failed to update user",
		},
		{
			name:   "update all fields",
			userID: uuid.New(),
			req: &models.UpdateUserRequest{
				DisplayName:     stringPtr("New Display Name"),
				Bio:             stringPtr("New bio"),
				Location:        &models.Location{Latitude: 40.7128, Longitude: -74.0060},
				City:            stringPtr("New York"),
				Country:         stringPtr("USA"),
				Genres:          []string{"Electronic", "Hip-Hop"},
				Skills:          []string{"Producer", "DJ"},
				SpotifyURL:      stringPtr("https://spotify.com/user123"),
				SoundcloudURL:   stringPtr("https://soundcloud.com/user123"),
				InstagramHandle: stringPtr("@user123"),
			},
			setupMocks: func(userRepo *ExtendedMockUserRepository, cache *MockCache, s3Client *MockS3Client) {
				// No setup needed - we'll add the user in the test
			},
			expectError: false,
			expectUser:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewExtendedMockUserRepository()
			cache := NewMockCache()
			s3Client := NewMockS3Client()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, s3Client)
			}
			
			// For successful tests, add the user with the correct ID
			if !tt.expectError && tt.expectUser {
				user := &models.User{
					ID:       tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[tt.userID.String()] = user
			}
			
			// For database error on update test, add user but set update error
			if tt.name == "database error on update" {
				user := &models.User{
					ID:       tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[tt.userID.String()] = user
			}
			
			// Create REAL UserService with mocks
			userService := NewUserService(userRepo, cache, s3Client)
			
			// Test UpdateUser
			user, err := userService.UpdateUser(context.Background(), tt.userID, tt.req)
			
			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
				if user != nil {
					t.Error("Expected user to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUser && user == nil {
					t.Error("Expected user but got nil")
				}
				
				// Verify user was updated with correct data
				if user != nil && tt.req != nil {
					if tt.req.DisplayName != nil && (user.DisplayName == nil || *user.DisplayName != *tt.req.DisplayName) {
						t.Errorf("Expected DisplayName %s, got %v", *tt.req.DisplayName, user.DisplayName)
					}
					if tt.req.Bio != nil && (user.Bio == nil || *user.Bio != *tt.req.Bio) {
						t.Errorf("Expected Bio %s, got %v", *tt.req.Bio, user.Bio)
					}
					if tt.req.City != nil && (user.City == nil || *user.City != *tt.req.City) {
						t.Errorf("Expected City %s, got %v", *tt.req.City, user.City)
					}
					if tt.req.Country != nil && (user.Country == nil || *user.Country != *tt.req.Country) {
						t.Errorf("Expected Country %s, got %v", *tt.req.Country, user.Country)
					}
				}
			}
		})
	}
}
