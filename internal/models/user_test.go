package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUser_ToResponse(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected *UserResponse
	}{
		{
			name: "complete user with all fields",
			user: &User{
				ID:                uuid.New(),
				Username:          "testuser",
				Email:             "test@example.com",
				PasswordHash:      "hashedpassword",
				DisplayName:       stringPtr("Test User"),
				Bio:               stringPtr("Test bio"),
				ProfilePictureURL: stringPtr("https://example.com/pic.jpg"),
				Location: &Location{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				City:            stringPtr("New York"),
				Country:         stringPtr("USA"),
				Genres:          []string{"Electronic", "Hip Hop"},
				Skills:          []string{"Producer", "DJ"},
				SpotifyURL:      stringPtr("https://spotify.com/user"),
				SoundcloudURL:   stringPtr("https://soundcloud.com/user"),
				InstagramHandle: stringPtr("@testuser"),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			expected: &UserResponse{
				ID:                uuid.UUID{}, // Will be set in test
				Username:          "testuser",
				Email:             "test@example.com",
				DisplayName:       stringPtr("Test User"),
				Bio:               stringPtr("Test bio"),
				ProfilePictureURL: stringPtr("https://example.com/pic.jpg"),
				Location: &Location{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				City:            stringPtr("New York"),
				Country:         stringPtr("USA"),
				Genres:          []string{"Electronic", "Hip Hop"},
				Skills:          []string{"Producer", "DJ"},
				SpotifyURL:      stringPtr("https://spotify.com/user"),
				SoundcloudURL:   stringPtr("https://soundcloud.com/user"),
				InstagramHandle: stringPtr("@testuser"),
				CreatedAt:       time.Time{}, // Will be set in test
				UpdatedAt:       time.Time{}, // Will be set in test
			},
		},
		{
			name: "minimal user with only required fields",
			user: &User{
				ID:           uuid.New(),
				Username:     "minimal",
				Email:        "minimal@example.com",
				PasswordHash: "hashedpassword",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			expected: &UserResponse{
				ID:                uuid.UUID{}, // Will be set in test
				Username:          "minimal",
				Email:             "minimal@example.com",
				DisplayName:       nil,
				Bio:               nil,
				ProfilePictureURL: nil,
				Location:          nil,
				City:              nil,
				Country:           nil,
				Genres:            nil,
				Skills:            nil,
				SpotifyURL:        nil,
				SoundcloudURL:     nil,
				InstagramHandle:   nil,
				CreatedAt:         time.Time{}, // Will be set in test
				UpdatedAt:         time.Time{}, // Will be set in test
			},
		},
		{
			name: "user with partial optional fields",
			user: &User{
				ID:           uuid.New(),
				Username:     "partial",
				Email:        "partial@example.com",
				PasswordHash: "hashedpassword",
				DisplayName:  stringPtr("Partial User"),
				Bio:          nil,
				Location: &Location{
					Latitude:  51.5074,
					Longitude: -0.1278,
				},
				City:    stringPtr("London"),
				Country: nil,
				Genres:  []string{"Rock"},
				Skills:  nil,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: &UserResponse{
				ID:                uuid.UUID{}, // Will be set in test
				Username:          "partial",
				Email:             "partial@example.com",
				DisplayName:       stringPtr("Partial User"),
				Bio:               nil,
				ProfilePictureURL: nil,
				Location: &Location{
					Latitude:  51.5074,
					Longitude: -0.1278,
				},
				City:            stringPtr("London"),
				Country:         nil,
				Genres:          []string{"Rock"},
				Skills:          nil,
				SpotifyURL:      nil,
				SoundcloudURL:   nil,
				InstagramHandle: nil,
				CreatedAt:       time.Time{}, // Will be set in test
				UpdatedAt:       time.Time{}, // Will be set in test
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set expected values that should match the input
			tt.expected.ID = tt.user.ID
			tt.expected.CreatedAt = tt.user.CreatedAt
			tt.expected.UpdatedAt = tt.user.UpdatedAt

			// Test ToResponse method
			result := tt.user.ToResponse()

			// Verify all fields match
			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, result.ID)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("Expected Username %s, got %s", tt.expected.Username, result.Username)
			}
			if result.Email != tt.expected.Email {
				t.Errorf("Expected Email %s, got %s", tt.expected.Email, result.Email)
			}
			if !compareStringPtrs(result.DisplayName, tt.expected.DisplayName) {
				t.Errorf("Expected DisplayName %v, got %v", tt.expected.DisplayName, result.DisplayName)
			}
			if !compareStringPtrs(result.Bio, tt.expected.Bio) {
				t.Errorf("Expected Bio %v, got %v", tt.expected.Bio, result.Bio)
			}
			if !compareStringPtrs(result.ProfilePictureURL, tt.expected.ProfilePictureURL) {
				t.Errorf("Expected ProfilePictureURL %v, got %v", tt.expected.ProfilePictureURL, result.ProfilePictureURL)
			}
			if !compareLocations(result.Location, tt.expected.Location) {
				t.Errorf("Expected Location %v, got %v", tt.expected.Location, result.Location)
			}
			if !compareStringPtrs(result.City, tt.expected.City) {
				t.Errorf("Expected City %v, got %v", tt.expected.City, result.City)
			}
			if !compareStringPtrs(result.Country, tt.expected.Country) {
				t.Errorf("Expected Country %v, got %v", tt.expected.Country, result.Country)
			}
			if !compareStringSlices(result.Genres, tt.expected.Genres) {
				t.Errorf("Expected Genres %v, got %v", tt.expected.Genres, result.Genres)
			}
			if !compareStringSlices(result.Skills, tt.expected.Skills) {
				t.Errorf("Expected Skills %v, got %v", tt.expected.Skills, result.Skills)
			}
			if !compareStringPtrs(result.SpotifyURL, tt.expected.SpotifyURL) {
				t.Errorf("Expected SpotifyURL %v, got %v", tt.expected.SpotifyURL, result.SpotifyURL)
			}
			if !compareStringPtrs(result.SoundcloudURL, tt.expected.SoundcloudURL) {
				t.Errorf("Expected SoundcloudURL %v, got %v", tt.expected.SoundcloudURL, result.SoundcloudURL)
			}
			if !compareStringPtrs(result.InstagramHandle, tt.expected.InstagramHandle) {
				t.Errorf("Expected InstagramHandle %v, got %v", tt.expected.InstagramHandle, result.InstagramHandle)
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("Expected CreatedAt %v, got %v", tt.expected.CreatedAt, result.CreatedAt)
			}
			if !result.UpdatedAt.Equal(tt.expected.UpdatedAt) {
				t.Errorf("Expected UpdatedAt %v, got %v", tt.expected.UpdatedAt, result.UpdatedAt)
			}
		})
	}
}

// Helper functions for comparison
func stringPtr(s string) *string {
	return &s
}

func compareStringPtrs(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func compareLocations(a, b *Location) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Latitude == b.Latitude && a.Longitude == b.Longitude
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
