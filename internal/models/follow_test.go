package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFollow_ToResponse(t *testing.T) {
	tests := []struct {
		name     string
		follow   *Follow
		expected *FollowResponse
	}{
		{
			name: "complete follow with all fields",
			follow: &Follow{
				ID:              uuid.New(),
				FollowerID:      uuid.New(),
				FollowingType:   "user",
				FollowingUserID: uuidPtr(uuid.New()),
				FollowingBandID: nil,
				CreatedAt:       time.Now(),
				Follower:        &User{ID: uuid.New(), Username: "follower"},
				FollowingUser:   &User{ID: uuid.New(), Username: "following"},
				FollowingBand:   nil,
			},
			expected: &FollowResponse{
				ID:              uuid.UUID{}, // Will be set in test
				FollowerID:      uuid.UUID{}, // Will be set in test
				FollowingType:   "user",
				FollowingUserID: nil, // Will be set in test
				FollowingBandID: nil,
				CreatedAt:       time.Time{}, // Will be set in test
				Follower:        &User{ID: uuid.UUID{}, Username: "follower"}, // Will be set in test
				FollowingUser:   &User{ID: uuid.UUID{}, Username: "following"}, // Will be set in test
				FollowingBand:   nil,
			},
		},
		{
			name: "follow band with minimal fields",
			follow: &Follow{
				ID:              uuid.New(),
				FollowerID:      uuid.New(),
				FollowingType:   "band",
				FollowingUserID: nil,
				FollowingBandID: uuidPtr(uuid.New()),
				CreatedAt:       time.Now(),
				Follower:        nil,
				FollowingUser:   nil,
				FollowingBand:   &Band{ID: uuid.New(), Name: "following band"},
			},
			expected: &FollowResponse{
				ID:              uuid.UUID{}, // Will be set in test
				FollowerID:      uuid.UUID{}, // Will be set in test
				FollowingType:   "band",
				FollowingUserID: nil,
				FollowingBandID: nil, // Will be set in test
				CreatedAt:       time.Time{}, // Will be set in test
				Follower:        nil,
				FollowingUser:   nil,
				FollowingBand:   &Band{ID: uuid.UUID{}, Name: "following band"}, // Will be set in test
			},
		},
		{
			name: "follow with partial optional fields",
			follow: &Follow{
				ID:              uuid.New(),
				FollowerID:      uuid.New(),
				FollowingType:   "user",
				FollowingUserID: uuidPtr(uuid.New()),
				FollowingBandID: nil,
				CreatedAt:       time.Now(),
				Follower:        &User{ID: uuid.New(), Username: "partial follower"},
				FollowingUser:   nil,
				FollowingBand:   nil,
			},
			expected: &FollowResponse{
				ID:              uuid.UUID{}, // Will be set in test
				FollowerID:      uuid.UUID{}, // Will be set in test
				FollowingType:   "user",
				FollowingUserID: nil, // Will be set in test
				FollowingBandID: nil,
				CreatedAt:       time.Time{}, // Will be set in test
				Follower:        &User{ID: uuid.UUID{}, Username: "partial follower"}, // Will be set in test
				FollowingUser:   nil,
				FollowingBand:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set expected values that should match the input
			tt.expected.ID = tt.follow.ID
			tt.expected.FollowerID = tt.follow.FollowerID
			tt.expected.CreatedAt = tt.follow.CreatedAt

			// Set UUID fields
			if tt.follow.FollowingUserID != nil {
				tt.expected.FollowingUserID = tt.follow.FollowingUserID
			}
			if tt.follow.FollowingBandID != nil {
				tt.expected.FollowingBandID = tt.follow.FollowingBandID
			}

			// Set joined data
			if tt.follow.Follower != nil {
				tt.expected.Follower = tt.follow.Follower
			}
			if tt.follow.FollowingUser != nil {
				tt.expected.FollowingUser = tt.follow.FollowingUser
			}
			if tt.follow.FollowingBand != nil {
				tt.expected.FollowingBand = tt.follow.FollowingBand
			}

			// Test ToResponse method
			result := tt.follow.ToResponse()

			// Verify all fields match
			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, result.ID)
			}
			if result.FollowerID != tt.expected.FollowerID {
				t.Errorf("Expected FollowerID %v, got %v", tt.expected.FollowerID, result.FollowerID)
			}
			if result.FollowingType != tt.expected.FollowingType {
				t.Errorf("Expected FollowingType %s, got %s", tt.expected.FollowingType, result.FollowingType)
			}
			if !compareUUIDPtrs(result.FollowingUserID, tt.expected.FollowingUserID) {
				t.Errorf("Expected FollowingUserID %v, got %v", tt.expected.FollowingUserID, result.FollowingUserID)
			}
			if !compareUUIDPtrs(result.FollowingBandID, tt.expected.FollowingBandID) {
				t.Errorf("Expected FollowingBandID %v, got %v", tt.expected.FollowingBandID, result.FollowingBandID)
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("Expected CreatedAt %v, got %v", tt.expected.CreatedAt, result.CreatedAt)
			}
			if !compareUserPtrs(result.Follower, tt.expected.Follower) {
				t.Errorf("Expected Follower %v, got %v", tt.expected.Follower, result.Follower)
			}
			if !compareUserPtrs(result.FollowingUser, tt.expected.FollowingUser) {
				t.Errorf("Expected FollowingUser %v, got %v", tt.expected.FollowingUser, result.FollowingUser)
			}
			if !compareBandPtrs(result.FollowingBand, tt.expected.FollowingBand) {
				t.Errorf("Expected FollowingBand %v, got %v", tt.expected.FollowingBand, result.FollowingBand)
			}
		})
	}
}

// Helper functions for comparison
func compareUserPtrs(a, b *User) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.ID == b.ID && a.Username == b.Username
}

func compareBandPtrs(a, b *Band) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.ID == b.ID && a.Name == b.Name
}
