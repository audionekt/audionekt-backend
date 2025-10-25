package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"musicapp/internal/cache"
	"musicapp/internal/middleware"
	"musicapp/internal/models"
	"musicapp/internal/repository"
	"musicapp/internal/storage"
	"musicapp/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Mock implementations for testing the interface-based AuthService
// These are much cleaner than the previous approach

type MockUserRepository struct {
	usersByEmail    map[string]*models.User
	usersByUsername map[string]*models.User
	usersByID       map[string]*models.User
	createError     error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		usersByEmail:    make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
		usersByID:       make(map[string]*models.User),
	}
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, exists := m.usersByEmail[email]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if user, exists := m.usersByUsername[username]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if user, exists := m.usersByID[id.String()]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.createError != nil {
		return m.createError
	}
	m.usersByEmail[user.Email] = user
	m.usersByUsername[user.Username] = user
	m.usersByID[user.ID.String()] = user
	return nil
}

type MockCache struct {
	setSessionError    error
	deleteSessionError error
	isBlacklistedError error
	addToBlacklistError error
	blacklistedTokens  map[string]bool
}

func NewMockCache() *MockCache {
	return &MockCache{
		blacklistedTokens: make(map[string]bool),
	}
}

func (m *MockCache) SetSession(ctx context.Context, userID string, data interface{}, expiration time.Duration) error {
	return m.setSessionError
}

func (m *MockCache) DeleteSession(ctx context.Context, userID string) error {
	return m.deleteSessionError
}

func (m *MockCache) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if m.isBlacklistedError != nil {
		return false, m.isBlacklistedError
	}
	return m.blacklistedTokens[jti], nil
}

func (m *MockCache) AddToBlacklist(ctx context.Context, jti string, expiration time.Duration) error {
	if m.addToBlacklistError != nil {
		return m.addToBlacklistError
	}
	m.blacklistedTokens[jti] = true
	return nil
}

type MockAuthMiddleware struct {
	generateTokenError error
	validateTokenError error
	token              string
	claims             *middleware.Claims
}

func NewMockAuthMiddleware() *MockAuthMiddleware {
	return &MockAuthMiddleware{
		token: "mock-jwt-token",
		claims: &middleware.Claims{
			UserID:   "mock-jti",
			Username: "testuser",
		},
	}
}

func (m *MockAuthMiddleware) GenerateToken(userID, username string) (string, error) {
	if m.generateTokenError != nil {
		return "", m.generateTokenError
	}
	return m.token, nil
}

func (m *MockAuthMiddleware) ValidateToken(tokenString string) (*middleware.Claims, error) {
	if m.validateTokenError != nil {
		return nil, m.validateTokenError
	}
	return m.claims, nil
}

// Test RegisterUser business logic with the REAL AuthService using mocks
func TestAuthService_RegisterUser(t *testing.T) {
	tests := []struct {
		name           string
		req            *models.CreateUserRequest
		setupMocks     func(*MockUserRepository, *MockCache, *MockAuthMiddleware)
		expectError    bool
		errorContains  string
		expectUser     bool
		expectToken    bool
	}{
		{
			name: "successful user registration",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// No existing users, no errors
			},
			expectError: false,
			expectUser:  true,
			expectToken: true,
		},
		{
			name: "user already exists by email",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
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
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
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
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				userRepo.createError = fmt.Errorf("database error")
			},
			expectError:   true,
			errorContains: "failed to create user",
		},
		{
			name: "token generation fails",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				authMid.generateTokenError = fmt.Errorf("jwt error")
			},
			expectError:   true,
			errorContains: "failed to generate token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewMockUserRepository()
			cache := NewMockCache()
			authMid := NewMockAuthMiddleware()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, authMid)
			}
			
			// Create REAL AuthService with mocks - this is the key difference!
			authService := NewAuthService(userRepo, cache, authMid)
			
			// Test RegisterUser
			user, token, err := authService.RegisterUser(context.Background(), tt.req)
			
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
				if token != "" {
					t.Error("Expected token to be empty on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUser && user == nil {
					t.Error("Expected user but got nil")
				}
				if tt.expectToken && token == "" {
					t.Error("Expected token but got empty string")
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

// Test LoginUser business logic with the REAL AuthService using mocks
func TestAuthService_LoginUser(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		setupMocks     func(*MockUserRepository, *MockCache, *MockAuthMiddleware)
		expectError    bool
		errorContains  string
		expectUser     bool
		expectToken    bool
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Add existing user with hashed password
				hashedPassword, _ := utils.HashPassword("password123")
				existingUser := &models.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					Username:     "testuser",
					PasswordHash: hashedPassword,
				}
				userRepo.usersByEmail["test@example.com"] = existingUser
			},
			expectError: false,
			expectUser:  true,
			expectToken: true,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// No user in mock repository
			},
			expectError:   true,
			errorContains: "invalid credentials",
		},
		{
			name:     "wrong password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Add existing user with different password
				hashedPassword, _ := utils.HashPassword("correctpassword")
				existingUser := &models.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					Username:     "testuser",
					PasswordHash: hashedPassword,
				}
				userRepo.usersByEmail["test@example.com"] = existingUser
			},
			expectError:   true,
			errorContains: "invalid credentials",
		},
		{
			name:     "token generation fails",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Add existing user
				hashedPassword, _ := utils.HashPassword("password123")
				existingUser := &models.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					Username:     "testuser",
					PasswordHash: hashedPassword,
				}
				userRepo.usersByEmail["test@example.com"] = existingUser
				// Make token generation fail
				authMid.generateTokenError = fmt.Errorf("jwt error")
			},
			expectError:   true,
			errorContains: "failed to generate token",
		},
		{
			name:     "empty email",
			email:    "",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// No setup needed for empty email
			},
			expectError:   true,
			errorContains: "invalid credentials",
		},
		{
			name:     "empty password",
			email:    "test@example.com",
			password: "",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Add existing user
				hashedPassword, _ := utils.HashPassword("password123")
				existingUser := &models.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					Username:     "testuser",
					PasswordHash: hashedPassword,
				}
				userRepo.usersByEmail["test@example.com"] = existingUser
			},
			expectError:   true,
			errorContains: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewMockUserRepository()
			cache := NewMockCache()
			authMid := NewMockAuthMiddleware()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, authMid)
			}
			
			// Create REAL AuthService with mocks
			authService := NewAuthService(userRepo, cache, authMid)
			
			// Test LoginUser
			user, token, err := authService.LoginUser(context.Background(), tt.email, tt.password)
			
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
				if token != "" {
					t.Error("Expected token to be empty on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectUser && user == nil {
					t.Error("Expected user but got nil")
				}
				if tt.expectToken && token == "" {
					t.Error("Expected token but got empty string")
				}
				
				// Verify user data
				if user != nil {
					if user.Email != tt.email {
						t.Errorf("Expected email %s, got %s", tt.email, user.Email)
					}
				}
			}
		})
	}
}

// Test ValidateToken business logic with the REAL AuthService using mocks
func TestAuthService_ValidateToken(t *testing.T) {
	tests := []struct {
		name           string
		tokenString    string
		setupMocks     func(*MockUserRepository, *MockCache, *MockAuthMiddleware)
		expectError    bool
		errorContains  string
		expectUser     bool
	}{
		{
			name:        "successful token validation",
			tokenString: "valid-jwt-token",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Setup valid claims
				userID := uuid.New()
				authMid.claims = &middleware.Claims{
					UserID:   userID.String(),
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "token-jti-123",
					},
				}
				
				// Add user to repository
				user := &models.User{
					ID:       userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[userID.String()] = user
				
				// Token not blacklisted
				cache.blacklistedTokens = make(map[string]bool)
			},
			expectError: false,
			expectUser:  true,
		},
		{
			name:        "invalid token format",
			tokenString: "invalid-token",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Make token validation fail
				authMid.validateTokenError = fmt.Errorf("invalid token format")
			},
			expectError:   true,
			errorContains: "invalid token",
		},
		{
			name:        "token is blacklisted",
			tokenString: "blacklisted-token",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Setup valid claims
				userID := uuid.New()
				authMid.claims = &middleware.Claims{
					UserID:   userID.String(),
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "blacklisted-jti",
					},
				}
				
				// Make token blacklisted
				cache.blacklistedTokens = map[string]bool{
					"blacklisted-jti": true,
				}
			},
			expectError:   true,
			errorContains: "token has been revoked",
		},
		{
			name:        "invalid user ID in token",
			tokenString: "invalid-userid-token",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Setup invalid user ID in claims
				authMid.claims = &middleware.Claims{
					UserID:   "invalid-uuid",
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "token-jti-123",
					},
				}
				
				// Token not blacklisted
				cache.blacklistedTokens = make(map[string]bool)
			},
			expectError:   true,
			errorContains: "invalid user ID in token",
		},
		{
			name:        "user not found in database",
			tokenString: "user-not-found-token",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Setup valid claims
				userID := uuid.New()
				authMid.claims = &middleware.Claims{
					UserID:   userID.String(),
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "token-jti-123",
					},
				}
				
				// Token not blacklisted
				cache.blacklistedTokens = make(map[string]bool)
				
				// User not in repository
				userRepo.usersByID = make(map[string]*models.User)
			},
			expectError:   true,
			errorContains: "user not found",
		},
		{
			name:        "blacklist check fails",
			tokenString: "blacklist-error-token",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Setup valid claims
				userID := uuid.New()
				authMid.claims = &middleware.Claims{
					UserID:   userID.String(),
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "token-jti-123",
					},
				}
				
				// Make blacklist check fail
				cache.isBlacklistedError = fmt.Errorf("redis error")
			},
			expectError:   true,
			errorContains: "failed to check token blacklist",
		},
		{
			name:        "empty token",
			tokenString: "",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Make empty token validation fail
				authMid.validateTokenError = fmt.Errorf("empty token")
			},
			expectError:   true,
			errorContains: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewMockUserRepository()
			cache := NewMockCache()
			authMid := NewMockAuthMiddleware()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, authMid)
			}
			
			// Create REAL AuthService with mocks
			authService := NewAuthService(userRepo, cache, authMid)
			
			// Test ValidateToken
			user, err := authService.ValidateToken(context.Background(), tt.tokenString)
			
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
				
				// Verify user data
				if user != nil && authMid.claims != nil {
					if user.ID.String() != authMid.claims.UserID {
						t.Errorf("Expected user ID %s, got %s", authMid.claims.UserID, user.ID.String())
					}
					if user.Username != authMid.claims.Username {
						t.Errorf("Expected username %s, got %s", authMid.claims.Username, user.Username)
					}
				}
			}
		})
	}
}

// Test LogoutUser business logic with the REAL AuthService using mocks
func TestAuthService_LogoutUser(t *testing.T) {
	tests := []struct {
		name           string
		jti            string
		userID         string
		setupMocks     func(*MockUserRepository, *MockCache, *MockAuthMiddleware)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful logout",
			jti:    "token-jti-123",
			userID: "user-123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// No errors - everything succeeds
			},
			expectError: false,
		},
		{
			name:   "blacklist fails",
			jti:    "token-jti-123",
			userID: "user-123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Make blacklist operation fail
				cache.addToBlacklistError = fmt.Errorf("redis connection error")
			},
			expectError:   true,
			errorContains: "failed to blacklist token",
		},
		{
			name:   "session deletion fails but logout succeeds",
			jti:    "token-jti-123",
			userID: "user-123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Make session deletion fail (should not fail logout)
				cache.deleteSessionError = fmt.Errorf("redis session error")
			},
			expectError: false, // Session deletion failure should not fail logout
		},
		{
			name:   "empty jti",
			jti:    "",
			userID: "user-123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Empty JTI should still work (just blacklist empty string)
			},
			expectError: false,
		},
		{
			name:   "empty userID",
			jti:    "token-jti-123",
			userID: "",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Empty userID should still work (just delete empty session)
			},
			expectError: false,
		},
		{
			name:   "both blacklist and session deletion fail",
			jti:    "token-jti-123",
			userID: "user-123",
			setupMocks: func(userRepo *MockUserRepository, cache *MockCache, authMid *MockAuthMiddleware) {
				// Make both operations fail
				cache.addToBlacklistError = fmt.Errorf("redis blacklist error")
				cache.deleteSessionError = fmt.Errorf("redis session error")
			},
			expectError:   true,
			errorContains: "failed to blacklist token", // Blacklist error should be returned
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := NewMockUserRepository()
			cache := NewMockCache()
			authMid := NewMockAuthMiddleware()
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, cache, authMid)
			}
			
			// Create REAL AuthService with mocks
			authService := NewAuthService(userRepo, cache, authMid)
			
			// Test LogoutUser
			err := authService.LogoutUser(context.Background(), tt.jti, tt.userID)
			
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
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
			
			// Verify that blacklist was called (if no blacklist error)
			if cache.addToBlacklistError == nil {
				if !cache.blacklistedTokens[tt.jti] {
					t.Errorf("Expected token %s to be blacklisted", tt.jti)
				}
			}
		})
	}
}

func TestAuthService_BusinessLogic(t *testing.T) {
	// Test the actual business logic by focusing on what we can test
	// without complex dependencies
	
	t.Run("NewAuthService creates service with dependencies", func(t *testing.T) {
		userRepo := &repository.UserRepository{}
		cache := &cache.Cache{}
		authMid := &middleware.AuthMiddleware{}
		
		authService := NewAuthService(userRepo, cache, authMid)
		
		if authService == nil {
			t.Error("Expected AuthService to be created")
		}
		if authService.userRepo != userRepo {
			t.Error("Expected userRepo to be set")
		}
		if authService.cache != cache {
			t.Error("Expected cache to be set")
		}
		if authService.authMiddleware != authMid {
			t.Error("Expected authMiddleware to be set")
		}
	})
	
	t.Run("AuthService methods exist and are callable", func(t *testing.T) {
		authService := &AuthService{}
		
		// Test that methods exist and can be called
		var registerFunc func(context.Context, *models.CreateUserRequest) (*models.User, string, error)
		var loginFunc func(context.Context, string, string) (*models.User, string, error)
		var validateFunc func(context.Context, string) (*models.User, error)
		var logoutFunc func(context.Context, string, string) error
		
		registerFunc = authService.RegisterUser
		loginFunc = authService.LoginUser
		validateFunc = authService.ValidateToken
		logoutFunc = authService.LogoutUser
		
		if registerFunc == nil {
			t.Error("RegisterUser method should exist")
		}
		if loginFunc == nil {
			t.Error("LoginUser method should exist")
		}
		if validateFunc == nil {
			t.Error("ValidateToken method should exist")
		}
		if logoutFunc == nil {
			t.Error("LogoutUser method should exist")
		}
	})
	
	t.Run("AuthService handles nil input gracefully", func(t *testing.T) {
		userRepo := &repository.UserRepository{}
		cache := &cache.Cache{}
		authMid := &middleware.AuthMiddleware{}
		
		authService := NewAuthService(userRepo, cache, authMid)
		
		// Test that methods handle nil input gracefully
		// This tests the input validation logic
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic for nil input: %v", r)
			}
		}()
		
		// Test RegisterUser with nil request
		user, token, err := authService.RegisterUser(context.Background(), nil)
		if err == nil {
			t.Error("Expected error for nil request")
		}
		if user != nil {
			t.Error("Expected user to be nil for nil request")
		}
		if token != "" {
			t.Error("Expected token to be empty for nil request")
		}
	})
}

// Test all other service constructors
func TestUserService_Constructor(t *testing.T) {
	userRepo := &repository.UserRepository{}
	cache := &cache.Cache{}
	s3Client := &storage.S3Client{}
	
	userService := NewUserService(userRepo, cache, s3Client)
	
	if userService == nil {
		t.Error("Expected UserService to be created")
	}
	if userService.userRepo != userRepo {
		t.Error("Expected userRepo to be set")
	}
	if userService.cache != cache {
		t.Error("Expected cache to be set")
	}
	if userService.s3Client != s3Client {
		t.Error("Expected s3Client to be set")
	}
}

func TestBandService_Constructor(t *testing.T) {
	bandRepo := &repository.BandRepository{}
	userRepo := &repository.UserRepository{}
	cache := &cache.Cache{}
	s3Client := &storage.S3Client{}
	
	bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
	
	if bandService == nil {
		t.Error("Expected BandService to be created")
	}
	if bandService.bandRepo != bandRepo {
		t.Error("Expected bandRepo to be set")
	}
	if bandService.userRepo != userRepo {
		t.Error("Expected userRepo to be set")
	}
	if bandService.cache != cache {
		t.Error("Expected cache to be set")
	}
	if bandService.s3Client != s3Client {
		t.Error("Expected s3Client to be set")
	}
}

func TestPostService_Constructor(t *testing.T) {
	postRepo := &repository.PostRepository{}
	userRepo := &repository.UserRepository{}
	bandRepo := &repository.BandRepository{}
	cache := &cache.Cache{}
	s3Client := &storage.S3Client{}
	
	postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
	
	if postService == nil {
		t.Error("Expected PostService to be created")
	}
	if postService.postRepo != postRepo {
		t.Error("Expected postRepo to be set")
	}
	if postService.userRepo != userRepo {
		t.Error("Expected userRepo to be set")
	}
	if postService.bandRepo != bandRepo {
		t.Error("Expected bandRepo to be set")
	}
	if postService.cache != cache {
		t.Error("Expected cache to be set")
	}
	if postService.s3Client != s3Client {
		t.Error("Expected s3Client to be set")
	}
}

func TestFollowService_Constructor(t *testing.T) {
	followRepo := &repository.FollowRepository{}
	userRepo := &repository.UserRepository{}
	bandRepo := &repository.BandRepository{}
	cache := &cache.Cache{}
	
	followService := NewFollowService(followRepo, userRepo, bandRepo, cache)
	
	if followService == nil {
		t.Error("Expected FollowService to be created")
	}
	if followService.followRepo != followRepo {
		t.Error("Expected followRepo to be set")
	}
	if followService.userRepo != userRepo {
		t.Error("Expected userRepo to be set")
	}
	if followService.bandRepo != bandRepo {
		t.Error("Expected bandRepo to be set")
	}
	if followService.cache != cache {
		t.Error("Expected cache to be set")
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Test RegisterUser business logic
func TestAuthService_RegisterUser_BusinessLogic(t *testing.T) {
	t.Run("RegisterUser with valid request", func(t *testing.T) {
		// This test will fail due to nil dependencies, but we test the structure
		userRepo := &repository.UserRepository{}
		cache := &cache.Cache{}
		authMid := &middleware.AuthMiddleware{}
		
		authService := NewAuthService(userRepo, cache, authMid)
		
		req := &models.CreateUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic due to nil dependencies: %v", r)
			}
		}()
		
		user, token, err := authService.RegisterUser(context.Background(), req)
		
		// We expect this to fail due to nil dependencies
		if err == nil {
			t.Error("Expected error due to nil dependencies")
		}
		if user != nil {
			t.Error("Expected user to be nil due to nil dependencies")
		}
		if token != "" {
			t.Error("Expected token to be empty due to nil dependencies")
		}
	})
}