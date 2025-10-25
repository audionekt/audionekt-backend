package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBand_ToResponse(t *testing.T) {
	tests := []struct {
		name     string
		band     *Band
		expected *BandResponse
	}{
		{
			name: "complete band with all fields",
			band: &Band{
				ID:                uuid.New(),
				Name:              "Test Band",
				Bio:               stringPtr("Test band bio"),
				ProfilePictureURL: stringPtr("https://example.com/band.jpg"),
				Location: &Location{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				City:       stringPtr("New York"),
				Country:    stringPtr("USA"),
				Genres:     []string{"Rock", "Alternative"},
				LookingFor: []string{"Guitarist", "Drummer"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expected: &BandResponse{
				ID:                uuid.UUID{}, // Will be set in test
				Name:              "Test Band",
				Bio:               stringPtr("Test band bio"),
				ProfilePictureURL: stringPtr("https://example.com/band.jpg"),
				Location: &Location{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				City:       stringPtr("New York"),
				Country:    stringPtr("USA"),
				Genres:     []string{"Rock", "Alternative"},
				LookingFor: []string{"Guitarist", "Drummer"},
				CreatedAt:  time.Time{}, // Will be set in test
				UpdatedAt:  time.Time{}, // Will be set in test
			},
		},
		{
			name: "minimal band with only required fields",
			band: &Band{
				ID:         uuid.New(),
				Name:       "Minimal Band",
				Bio:        nil,
				Location:   nil,
				City:       nil,
				Country:    nil,
				Genres:     nil,
				LookingFor: nil,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expected: &BandResponse{
				ID:                uuid.UUID{}, // Will be set in test
				Name:              "Minimal Band",
				Bio:               nil,
				ProfilePictureURL: nil,
				Location:          nil,
				City:              nil,
				Country:           nil,
				Genres:            nil,
				LookingFor:        nil,
				CreatedAt:         time.Time{}, // Will be set in test
				UpdatedAt:         time.Time{}, // Will be set in test
			},
		},
		{
			name: "band with partial optional fields",
			band: &Band{
				ID:         uuid.New(),
				Name:       "Partial Band",
				Bio:        stringPtr("Partial band bio"),
				Location:   nil,
				City:       stringPtr("London"),
				Country:    stringPtr("UK"),
				Genres:     []string{"Electronic"},
				LookingFor: nil,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expected: &BandResponse{
				ID:                uuid.UUID{}, // Will be set in test
				Name:              "Partial Band",
				Bio:               stringPtr("Partial band bio"),
				ProfilePictureURL: nil,
				Location:          nil,
				City:              stringPtr("London"),
				Country:           stringPtr("UK"),
				Genres:            []string{"Electronic"},
				LookingFor:        nil,
				CreatedAt:         time.Time{}, // Will be set in test
				UpdatedAt:         time.Time{}, // Will be set in test
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set expected values that should match the input
			tt.expected.ID = tt.band.ID
			tt.expected.CreatedAt = tt.band.CreatedAt
			tt.expected.UpdatedAt = tt.band.UpdatedAt

			// Test ToResponse method
			result := tt.band.ToResponse()

			// Verify all fields match
			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, result.ID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Expected Name %s, got %s", tt.expected.Name, result.Name)
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
			if !compareStringSlices(result.LookingFor, tt.expected.LookingFor) {
				t.Errorf("Expected LookingFor %v, got %v", tt.expected.LookingFor, result.LookingFor)
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
