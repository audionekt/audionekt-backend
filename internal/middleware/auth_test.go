package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"musicapp/internal/cache"
)

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectStatus   int
		expectUserID   string
		expectUsername string
	}{
		{
			name:           "valid token",
			authHeader:     "", // Will be generated
			expectStatus:   http.StatusOK, // Should work now with nil cache check
			expectUserID:   "user123",
			expectUsername: "testuser",
		},
		{
			name:         "missing authorization header",
			authHeader:   "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid authorization header format",
			authHeader:   "InvalidFormat token123",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer invalid.token.here",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "empty token",
			authHeader:   "Bearer ",
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAuthMiddleware([]byte("test-secret-key"), nil)

			// Generate token if needed
			if tt.name == "valid token" {
				token, err := middleware.GenerateToken(tt.expectUserID, tt.expectUsername)
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				tt.authHeader = "Bearer " + token
			}

			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if user context is set
				if tt.expectUserID != "" {
					userID, ok := GetUserIDFromContext(r.Context())
					if !ok {
						t.Error("Expected userID in context")
					}
					if userID != tt.expectUserID {
						t.Errorf("Expected userID '%s', got '%s'", tt.expectUserID, userID)
					}

					username, ok := GetUsernameFromContext(r.Context())
					if !ok {
						t.Error("Expected username in context")
					}
					if username != tt.expectUsername {
						t.Errorf("Expected username '%s', got '%s'", tt.expectUsername, username)
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Apply middleware
			middleware.RequireAuth(handler).ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, rr.Code)
			}
		})
	}
}

func TestNewAuthMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		jwtSecret []byte
		cache    *cache.Cache
	}{
		{
			name:      "create auth middleware with valid parameters",
			jwtSecret: []byte("test-secret-key"),
			cache:     &cache.Cache{},
		},
		{
			name:      "create auth middleware with empty secret",
			jwtSecret: []byte(""),
			cache:     &cache.Cache{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAuthMiddleware(tt.jwtSecret, tt.cache)

			if middleware == nil {
				t.Error("Expected middleware to be created, got nil")
			}
			if middleware.jwtSecret == nil {
				t.Error("Expected jwtSecret to be set, got nil")
			}
			if middleware.cache == nil {
				t.Error("Expected cache to be set, got nil")
			}
		})
	}
}

func TestAuthMiddleware_ValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		jwtSecret   []byte
		tokenString string
		expectError bool
		expectClaims bool
	}{
		{
			name:        "valid token",
			jwtSecret:   []byte("test-secret-key"),
			tokenString: "", // Will be generated
			expectError: false,
			expectClaims: true,
		},
		{
			name:        "invalid token format",
			jwtSecret:   []byte("test-secret-key"),
			tokenString: "invalid.token.format",
			expectError: true,
			expectClaims: false,
		},
		{
			name:        "empty token",
			jwtSecret:   []byte("test-secret-key"),
			tokenString: "",
			expectError: true,
			expectClaims: false,
		},
		{
			name:        "token with wrong secret",
			jwtSecret:   []byte("test-secret-key"),
			tokenString: "", // Will be generated with different secret
			expectError: true,
			expectClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAuthMiddleware(tt.jwtSecret, nil)

			// Generate token if needed
			if tt.name == "valid token" {
				token, err := middleware.GenerateToken("user123", "testuser")
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				tt.tokenString = token
			} else if tt.name == "token with wrong secret" {
				// Generate token with different secret
				wrongMiddleware := NewAuthMiddleware([]byte("wrong-secret"), nil)
				token, err := wrongMiddleware.GenerateToken("user123", "testuser")
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				tt.tokenString = token
			}

			// Test ValidateToken
			claims, err := middleware.ValidateToken(tt.tokenString)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if claims != nil {
					t.Error("Expected claims to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectClaims && claims == nil {
					t.Error("Expected claims but got nil")
				}
				if claims != nil {
					if claims.UserID != "user123" {
						t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
					}
					if claims.Username != "testuser" {
						t.Errorf("Expected Username 'testuser', got '%s'", claims.Username)
					}
				}
			}
		})
	}
}

func TestAuthMiddleware_GenerateToken(t *testing.T) {
	tests := []struct {
		name      string
		jwtSecret []byte
		userID    string
		username  string
		expectError bool
	}{
		{
			name:      "generate valid token",
			jwtSecret: []byte("test-secret-key"),
			userID:    "user123",
			username:  "testuser",
			expectError: false,
		},
		{
			name:      "generate token with empty userID",
			jwtSecret: []byte("test-secret-key"),
			userID:    "",
			username:  "testuser",
			expectError: false,
		},
		{
			name:      "generate token with empty username",
			jwtSecret: []byte("test-secret-key"),
			userID:    "user123",
			username:  "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAuthMiddleware(tt.jwtSecret, nil)

			// Test GenerateToken
			token, err := middleware.GenerateToken(tt.userID, tt.username)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if token != "" {
					t.Error("Expected empty token on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if token == "" {
					t.Error("Expected token but got empty string")
				}

				// Verify token can be validated
				claims, err := middleware.ValidateToken(token)
				if err != nil {
					t.Errorf("Generated token should be valid, got error: %v", err)
				}
				if claims != nil {
					if claims.UserID != tt.userID {
						t.Errorf("Expected UserID '%s', got '%s'", tt.userID, claims.UserID)
					}
					if claims.Username != tt.username {
						t.Errorf("Expected Username '%s', got '%s'", tt.username, claims.Username)
					}
				}
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
		ok       bool
	}{
		{
			name:     "context with user_id",
			ctx:      context.WithValue(context.Background(), "user_id", "user123"),
			expected: "user123",
			ok:       true,
		},
		{
			name:     "context without user_id",
			ctx:      context.Background(),
			expected: "",
			ok:       false,
		},
		{
			name:     "context with wrong type",
			ctx:      context.WithValue(context.Background(), "user_id", 123),
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, ok := GetUserIDFromContext(tt.ctx)

			if ok != tt.ok {
				t.Errorf("Expected ok %v, got %v", tt.ok, ok)
			}
			if userID != tt.expected {
				t.Errorf("Expected userID '%s', got '%s'", tt.expected, userID)
			}
		})
	}
}

func TestGetUsernameFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
		ok       bool
	}{
		{
			name:     "context with username",
			ctx:      context.WithValue(context.Background(), "username", "testuser"),
			expected: "testuser",
			ok:       true,
		},
		{
			name:     "context without username",
			ctx:      context.Background(),
			expected: "",
			ok:       false,
		},
		{
			name:     "context with wrong type",
			ctx:      context.WithValue(context.Background(), "username", 123),
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, ok := GetUsernameFromContext(tt.ctx)

			if ok != tt.ok {
				t.Errorf("Expected ok %v, got %v", tt.ok, ok)
			}
			if username != tt.expected {
				t.Errorf("Expected username '%s', got '%s'", tt.expected, username)
			}
		})
	}
}

func TestGetJTIFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
		ok       bool
	}{
		{
			name:     "context with jti",
			ctx:      context.WithValue(context.Background(), "jti", "jti123"),
			expected: "jti123",
			ok:       true,
		},
		{
			name:     "context without jti",
			ctx:      context.Background(),
			expected: "",
			ok:       false,
		},
		{
			name:     "context with wrong type",
			ctx:      context.WithValue(context.Background(), "jti", 123),
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jti, ok := GetJTIFromContext(tt.ctx)

			if ok != tt.ok {
				t.Errorf("Expected ok %v, got %v", tt.ok, ok)
			}
			if jti != tt.expected {
				t.Errorf("Expected jti '%s', got '%s'", tt.expected, jti)
			}
		})
	}
}