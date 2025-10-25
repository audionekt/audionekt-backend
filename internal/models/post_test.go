package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPost_ToResponse(t *testing.T) {
	tests := []struct {
		name     string
		post     *Post
		expected *PostResponse
	}{
		{
			name: "complete post with all fields",
			post: &Post{
				ID:         uuid.New(),
				AuthorID:   uuidPtr(uuid.New()),
				AuthorType: "user",
				BandID:     nil,
				UserID:     uuidPtr(uuid.New()),
				Content:    "Test post content",
				MediaURLs:  []string{"https://example.com/image.jpg", "https://example.com/audio.mp3"},
				MediaTypes: []string{"image", "audio"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				Author:     "Test User",
				LikesCount: 10,
				RepostsCount: 5,
				IsLiked:    true,
				IsReposted: false,
			},
			expected: &PostResponse{
				ID:           uuid.UUID{}, // Will be set in test
				AuthorID:     nil, // Will be set in test
				AuthorType:   "user",
				BandID:       nil,
				UserID:       nil, // Will be set in test
				Content:      "Test post content",
				MediaURLs:    []string{"https://example.com/image.jpg", "https://example.com/audio.mp3"},
				MediaTypes:   []string{"image", "audio"},
				CreatedAt:    time.Time{}, // Will be set in test
				UpdatedAt:    time.Time{}, // Will be set in test
				Author:       "Test User",
				LikesCount:   10,
				RepostsCount: 5,
				IsLiked:      true,
				IsReposted:   false,
			},
		},
		{
			name: "minimal post with only required fields",
			post: &Post{
				ID:         uuid.New(),
				AuthorID:   nil,
				AuthorType: "band",
				BandID:     uuidPtr(uuid.New()),
				UserID:     nil,
				Content:    "Minimal post",
				MediaURLs:  nil,
				MediaTypes: nil,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				Author:     nil,
				LikesCount: 0,
				RepostsCount: 0,
				IsLiked:    false,
				IsReposted: false,
			},
			expected: &PostResponse{
				ID:           uuid.UUID{}, // Will be set in test
				AuthorID:     nil,
				AuthorType:   "band",
				BandID:       nil, // Will be set in test
				UserID:       nil,
				Content:      "Minimal post",
				MediaURLs:    nil,
				MediaTypes:   nil,
				CreatedAt:    time.Time{}, // Will be set in test
				UpdatedAt:    time.Time{}, // Will be set in test
				Author:       nil,
				LikesCount:   0,
				RepostsCount: 0,
				IsLiked:      false,
				IsReposted:   false,
			},
		},
		{
			name: "post with partial optional fields",
			post: &Post{
				ID:         uuid.New(),
				AuthorID:   uuidPtr(uuid.New()),
				AuthorType: "user",
				BandID:     nil,
				UserID:     uuidPtr(uuid.New()),
				Content:    "Partial post",
				MediaURLs:  []string{"https://example.com/video.mp4"},
				MediaTypes: []string{"video"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				Author:     "Partial User",
				LikesCount: 3,
				RepostsCount: 1,
				IsLiked:    false,
				IsReposted: true,
			},
			expected: &PostResponse{
				ID:           uuid.UUID{}, // Will be set in test
				AuthorID:     nil, // Will be set in test
				AuthorType:   "user",
				BandID:       nil,
				UserID:       nil, // Will be set in test
				Content:      "Partial post",
				MediaURLs:    []string{"https://example.com/video.mp4"},
				MediaTypes:   []string{"video"},
				CreatedAt:    time.Time{}, // Will be set in test
				UpdatedAt:    time.Time{}, // Will be set in test
				Author:       "Partial User",
				LikesCount:   3,
				RepostsCount: 1,
				IsLiked:      false,
				IsReposted:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set expected values that should match the input
			tt.expected.ID = tt.post.ID
			tt.expected.CreatedAt = tt.post.CreatedAt
			tt.expected.UpdatedAt = tt.post.UpdatedAt

			// Set UUID fields
			if tt.post.AuthorID != nil {
				tt.expected.AuthorID = tt.post.AuthorID
			}
			if tt.post.BandID != nil {
				tt.expected.BandID = tt.post.BandID
			}
			if tt.post.UserID != nil {
				tt.expected.UserID = tt.post.UserID
			}

			// Test ToResponse method
			result := tt.post.ToResponse()

			// Verify all fields match
			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, result.ID)
			}
			if !compareUUIDPtrs(result.AuthorID, tt.expected.AuthorID) {
				t.Errorf("Expected AuthorID %v, got %v", tt.expected.AuthorID, result.AuthorID)
			}
			if result.AuthorType != tt.expected.AuthorType {
				t.Errorf("Expected AuthorType %s, got %s", tt.expected.AuthorType, result.AuthorType)
			}
			if !compareUUIDPtrs(result.BandID, tt.expected.BandID) {
				t.Errorf("Expected BandID %v, got %v", tt.expected.BandID, result.BandID)
			}
			if !compareUUIDPtrs(result.UserID, tt.expected.UserID) {
				t.Errorf("Expected UserID %v, got %v", tt.expected.UserID, result.UserID)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("Expected Content %s, got %s", tt.expected.Content, result.Content)
			}
			if !compareStringSlices(result.MediaURLs, tt.expected.MediaURLs) {
				t.Errorf("Expected MediaURLs %v, got %v", tt.expected.MediaURLs, result.MediaURLs)
			}
			if !compareStringSlices(result.MediaTypes, tt.expected.MediaTypes) {
				t.Errorf("Expected MediaTypes %v, got %v", tt.expected.MediaTypes, result.MediaTypes)
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("Expected CreatedAt %v, got %v", tt.expected.CreatedAt, result.CreatedAt)
			}
			if !result.UpdatedAt.Equal(tt.expected.UpdatedAt) {
				t.Errorf("Expected UpdatedAt %v, got %v", tt.expected.UpdatedAt, result.UpdatedAt)
			}
			if result.Author != tt.expected.Author {
				t.Errorf("Expected Author %v, got %v", tt.expected.Author, result.Author)
			}
			if result.LikesCount != tt.expected.LikesCount {
				t.Errorf("Expected LikesCount %d, got %d", tt.expected.LikesCount, result.LikesCount)
			}
			if result.RepostsCount != tt.expected.RepostsCount {
				t.Errorf("Expected RepostsCount %d, got %d", tt.expected.RepostsCount, result.RepostsCount)
			}
			if result.IsLiked != tt.expected.IsLiked {
				t.Errorf("Expected IsLiked %v, got %v", tt.expected.IsLiked, result.IsLiked)
			}
			if result.IsReposted != tt.expected.IsReposted {
				t.Errorf("Expected IsReposted %v, got %v", tt.expected.IsReposted, result.IsReposted)
			}
		})
	}
}

// Helper function for UUID pointers
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

func compareUUIDPtrs(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
