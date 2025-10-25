package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"musicapp/internal/models"
	"musicapp/internal/storage"

	"github.com/google/uuid"
)

// Mock implementations for PostService testing

type MockPostRepository struct {
	postsByID     map[string]*models.Post
	userPosts     map[string][]*models.Post
	bandPosts     map[string][]*models.Post
	feedPosts     []*models.Post
	allPosts      []*models.Post
	createError   error
	getByIDError  error
	updateError   error
	deleteError   error
	getByUserIDError error
	getByBandIDError error
	getFeedError  error
	getAllError   error
	likePostError error
	unlikePostError error
	repostError   error
	isLikedResult bool
	isRepostedResult bool
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{
		postsByID: make(map[string]*models.Post),
		userPosts: make(map[string][]*models.Post),
		bandPosts: make(map[string][]*models.Post),
		feedPosts: []*models.Post{},
		allPosts:  []*models.Post{},
	}
}

func (m *MockPostRepository) Create(ctx context.Context, post *models.Post) error {
	if m.createError != nil {
		return m.createError
	}
	m.postsByID[post.ID.String()] = post
	return nil
}

func (m *MockPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	post, exists := m.postsByID[id.String()]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

func (m *MockPostRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	if m.getByUserIDError != nil {
		return nil, m.getByUserIDError
	}
	posts, exists := m.userPosts[userID.String()]
	if !exists {
		return []*models.Post{}, nil
	}
	return posts, nil
}

func (m *MockPostRepository) GetByBandID(ctx context.Context, bandID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	if m.getByBandIDError != nil {
		return nil, m.getByBandIDError
	}
	posts, exists := m.bandPosts[bandID.String()]
	if !exists {
		return []*models.Post{}, nil
	}
	return posts, nil
}

func (m *MockPostRepository) GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	if m.getFeedError != nil {
		return nil, m.getFeedError
	}
	return m.feedPosts, nil
}

func (m *MockPostRepository) Update(ctx context.Context, post *models.Post) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.postsByID[post.ID.String()] = post
	return nil
}

func (m *MockPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.postsByID, id.String())
	return nil
}

func (m *MockPostRepository) LikePost(ctx context.Context, userID, postID uuid.UUID) error {
	if m.likePostError != nil {
		return m.likePostError
	}
	return nil
}

func (m *MockPostRepository) UnlikePost(ctx context.Context, userID, postID uuid.UUID) error {
	if m.unlikePostError != nil {
		return m.unlikePostError
	}
	return nil
}

func (m *MockPostRepository) Repost(ctx context.Context, userID, postID uuid.UUID) error {
	if m.repostError != nil {
		return m.repostError
	}
	return nil
}

func (m *MockPostRepository) IsLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	return m.isLikedResult, nil
}

func (m *MockPostRepository) IsReposted(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	return m.isRepostedResult, nil
}

func (m *MockPostRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Post, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	return m.allPosts, nil
}

type MockUserRepositoryForPost struct {
	usersByID map[string]*models.User
	getByIDError error
}

func NewMockUserRepositoryForPost() *MockUserRepositoryForPost {
	return &MockUserRepositoryForPost{
		usersByID: make(map[string]*models.User),
	}
}

func (m *MockUserRepositoryForPost) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	user, exists := m.usersByID[id.String()]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

type MockBandRepositoryForPost struct {
	bandsByID map[string]*models.Band
	getByIDError error
}

func NewMockBandRepositoryForPost() *MockBandRepositoryForPost {
	return &MockBandRepositoryForPost{
		bandsByID: make(map[string]*models.Band),
	}
}

func (m *MockBandRepositoryForPost) GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	band, exists := m.bandsByID[id.String()]
	if !exists {
		return nil, fmt.Errorf("band not found")
	}
	return band, nil
}

type MockS3ClientForPost struct {
	uploadResult        *storage.UploadResult
	uploadError         error
	validateError       error
	validateImageError  error
	validateAudioError  error
}

func NewMockS3ClientForPost() *MockS3ClientForPost {
	return &MockS3ClientForPost{}
}

func (m *MockS3ClientForPost) ValidateImageFile(filename string, size int64) error {
	return m.validateImageError
}

func (m *MockS3ClientForPost) ValidateAudioFile(filename string, size int64) error {
	return m.validateAudioError
}

func (m *MockS3ClientForPost) UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error) {
	if m.uploadError != nil {
		return nil, m.uploadError
	}
	return m.uploadResult, nil
}

func (m *MockS3ClientForPost) UploadAudio(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error) {
	if m.uploadError != nil {
		return nil, m.uploadError
	}
	return m.uploadResult, nil
}

// Test NewPostService constructor
func TestNewPostService(t *testing.T) {
	postRepo := NewMockPostRepository()
	userRepo := NewMockUserRepositoryForPost()
	bandRepo := NewMockBandRepositoryForPost()
	cache := NewMockCache()
	s3Client := NewMockS3ClientForPost()

	service := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)

	if service == nil {
		t.Error("Expected PostService to be created, got nil")
	}

	if service.postRepo != postRepo {
		t.Error("Expected postRepo to be set correctly")
	}

	if service.userRepo != userRepo {
		t.Error("Expected userRepo to be set correctly")
	}

	if service.bandRepo != bandRepo {
		t.Error("Expected bandRepo to be set correctly")
	}

	if service.cache != cache {
		t.Error("Expected cache to be set correctly")
	}

	if service.s3Client == nil {
		t.Error("Expected s3Client to be set correctly")
	}
}

// Test CreatePost business logic with the REAL PostService using mocks
func TestPostService_CreatePost(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		req            *models.CreatePostRequest
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
		expectPost     bool
	}{
		{
			name:   "successful post creation",
			userID: uuid.New(),
			req: &models.CreatePostRequest{
				Content: "This is a test post",
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No special setup needed for successful creation
			},
			expectError: false,
			expectPost:  true,
		},
		{
			name:   "empty content",
			userID: uuid.New(),
			req: &models.CreatePostRequest{
				Content: "",
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "post content is required",
		},
		{
			name:   "content too long",
			userID: uuid.New(),
			req: &models.CreatePostRequest{
				Content: strings.Repeat("a", 2001), // 2001 characters
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "post content too long",
		},
		{
			name:   "database create error",
			userID: uuid.New(),
			req: &models.CreatePostRequest{
				Content: "This is a test post",
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Setup repository to return create error
				postRepo.createError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to create post",
		},
		{
			name:   "database get error after create",
			userID: uuid.New(),
			req: &models.CreatePostRequest{
				Content: "This is a test post",
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Setup repository to return get error after successful create
				postRepo.getByIDError = fmt.Errorf("database query error")
			},
			expectError:   true,
			errorContains: "failed to retrieve created post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test CreatePost
			post, err := postService.CreatePost(context.Background(), tt.userID, tt.req)
			
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
				if post != nil {
					t.Error("Expected post to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectPost && post == nil {
					t.Error("Expected post but got nil")
				}
				if post != nil && post.Content != tt.req.Content {
					t.Errorf("Expected post content '%s', got '%s'", tt.req.Content, post.Content)
				}
			}
		})
	}
}

// Test GetPost business logic with the REAL PostService using mocks
func TestPostService_GetPost(t *testing.T) {
	tests := []struct {
		name           string
		postID         uuid.UUID
		currentUserID  *uuid.UUID
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
		expectPost     bool
		expectLiked    bool
		expectReposted bool
	}{
		{
			name:          "successful post retrieval without current user",
			postID:        uuid.New(),
			currentUserID: nil,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				post := &models.Post{
					ID:      postID,
					Content: "Test post content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError: false,
			expectPost:  true,
		},
		{
			name:          "successful post retrieval with current user",
			postID:        uuid.New(),
			currentUserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				post := &models.Post{
					ID:      postID,
					Content: "Test post content",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Mock like/repost status
				postRepo.isLikedResult = true
				postRepo.isRepostedResult = false
			},
			expectError:    false,
			expectPost:     true,
			expectLiked:    true,
			expectReposted: false,
		},
		{
			name:          "post not found",
			postID:        uuid.New(),
			currentUserID: nil,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Don't add post - will cause "post not found" error
			},
			expectError:   true,
			errorContains: "post not found",
		},
		{
			name:          "database error",
			postID:        uuid.New(),
			currentUserID: nil,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Setup repository to return error
				postRepo.getByIDError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// For successful tests, add the post with the correct ID
			if !tt.expectError && tt.expectPost {
				post := &models.Post{
					ID:      tt.postID,
					Content: "Test post content",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test GetPost
			post, err := postService.GetPost(context.Background(), tt.postID, tt.currentUserID)
			
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
				if post != nil {
					t.Error("Expected post to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectPost && post == nil {
					t.Error("Expected post but got nil")
				}
				if post != nil {
					if tt.currentUserID != nil {
						if tt.expectLiked && !post.IsLiked {
							t.Error("Expected post to be liked")
						}
						if tt.expectReposted && !post.IsReposted {
							t.Error("Expected post to be reposted")
						}
					}
				}
			}
		})
	}
}
// Test UpdatePost business logic with the REAL PostService using mocks
func TestPostService_UpdatePost(t *testing.T) {
	tests := []struct {
		name           string
		postID         uuid.UUID
		userID         uuid.UUID
		req            *models.UpdatePostRequest
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
		expectPost     bool
	}{
		{
			name:   "successful post update",
			postID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdatePostRequest{
				Content: stringPtr("Updated post content"),
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Original content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError: false,
			expectPost:  true,
		},
		{
			name:   "post not found",
			postID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdatePostRequest{
				Content: stringPtr("Updated content"),
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Don't add post - will cause "post not found" error
			},
			expectError:   true,
			errorContains: "post not found",
		},
		{
			name:   "user not author",
			postID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdatePostRequest{
				Content: stringPtr("Updated content"),
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post with different user ID
				postID := uuid.New()
				differentUserID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &differentUserID,
					Content: "Original content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError:   true,
			errorContains: "you can only update your own posts",
		},
		{
			name:   "empty content",
			postID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdatePostRequest{
				Content: stringPtr(""),
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Original content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError:   true,
			errorContains: "post content cannot be empty",
		},
		{
			name:   "content too long",
			postID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdatePostRequest{
				Content: stringPtr(strings.Repeat("a", 2001)), // 2001 characters
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Original content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError:   true,
			errorContains: "post content too long",
		},
		{
			name:   "database update error",
			postID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdatePostRequest{
				Content: stringPtr("Updated content"),
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Original content",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Setup repository to return update error
				postRepo.updateError = fmt.Errorf("database update failed")
			},
			expectError:   true,
			errorContains: "failed to update post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// For tests that need a post, add the post with the correct IDs
			if tt.name == "successful post update" || tt.name == "empty content" || tt.name == "content too long" || tt.name == "database update error" {
				post := &models.Post{
					ID:      tt.postID,
					UserID:  &tt.userID,
					Content: "Original content",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// For "user not author" test, add post with different user ID
			if tt.name == "user not author" {
				differentUserID := uuid.New()
				post := &models.Post{
					ID:      tt.postID,
					UserID:  &differentUserID,
					Content: "Original content",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test UpdatePost
			post, err := postService.UpdatePost(context.Background(), tt.postID, tt.userID, tt.req)
			
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
				if post != nil {
					t.Error("Expected post to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectPost && post == nil {
					t.Error("Expected post but got nil")
				}
				if post != nil && tt.req.Content != nil && post.Content != *tt.req.Content {
					t.Errorf("Expected post content '%s', got '%s'", *tt.req.Content, post.Content)
				}
			}
		})
	}
}
// Test DeletePost business logic with the REAL PostService using mocks
func TestPostService_DeletePost(t *testing.T) {
	tests := []struct {
		name           string
		postID         uuid.UUID
		userID         uuid.UUID
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful post deletion",
			postID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No special setup needed for successful deletion
			},
			expectError: false,
		},
		{
			name:   "post not found",
			postID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Don't add post - will cause "post not found" error
			},
			expectError:   true,
			errorContains: "post not found",
		},
		{
			name:   "user not author",
			postID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post with different user ID
				postID := uuid.New()
				differentUserID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &differentUserID,
					Content: "Test content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError:   true,
			errorContains: "you can only delete your own posts",
		},
		{
			name:   "database delete error",
			postID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test content",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Setup repository to return delete error
				postRepo.deleteError = fmt.Errorf("database delete failed")
			},
			expectError:   true,
			errorContains: "failed to delete post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// For tests that need a post, add the post with the correct IDs
			if tt.name == "successful post deletion" || tt.name == "database delete error" {
				post := &models.Post{
					ID:      tt.postID,
					UserID:  &tt.userID,
					Content: "Test content",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// For "user not author" test, add post with different user ID
			if tt.name == "user not author" {
				differentUserID := uuid.New()
				post := &models.Post{
					ID:      tt.postID,
					UserID:  &differentUserID,
					Content: "Test content",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test DeletePost
			err := postService.DeletePost(context.Background(), tt.postID, tt.userID)
			
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
		})
	}
}// Test LikePost business logic with the REAL PostService using mocks
func TestPostService_LikePost(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		postID         uuid.UUID
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful post like",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				post := &models.Post{
					ID:      postID,
					Content: "Test content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError: false,
		},
		{
			name:   "post not found",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Don't add post - will cause "post not found" error
			},
			expectError:   true,
			errorContains: "post not found",
		},
		{
			name:   "database like error",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				post := &models.Post{
					ID:      postID,
					Content: "Test content",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Setup repository to return like error
				postRepo.likePostError = fmt.Errorf("database like failed")
			},
			expectError:   true,
			errorContains: "failed to like post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// For tests that need a post, add the post with the correct ID
			if tt.name == "successful post like" || tt.name == "database like error" {
				post := &models.Post{
					ID:      tt.postID,
					Content: "Test content",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test LikePost
			err := postService.LikePost(context.Background(), tt.userID, tt.postID)
			
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
		})
	}
}// Test UnlikePost business logic with the REAL PostService using mocks
func TestPostService_UnlikePost(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		postID         uuid.UUID
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful post unlike",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No special setup needed for successful unlike
			},
			expectError: false,
		},
		{
			name:   "database unlike error",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Setup repository to return unlike error
				postRepo.unlikePostError = fmt.Errorf("database unlike failed")
			},
			expectError:   true,
			errorContains: "failed to unlike post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test UnlikePost
			err := postService.UnlikePost(context.Background(), tt.userID, tt.postID)
			
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
		})
	}
}// Test Repost business logic with the REAL PostService using mocks
func TestPostService_Repost(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		postID         uuid.UUID
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful post repost",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				post := &models.Post{
					ID:      postID,
					Content: "Test content",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError: false,
		},
		{
			name:   "post not found",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Don't add post - will cause "post not found" error
			},
			expectError:   true,
			errorContains: "post not found",
		},
		{
			name:   "database repost error",
			userID: uuid.New(),
			postID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				post := &models.Post{
					ID:      postID,
					Content: "Test content",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Setup repository to return repost error
				postRepo.repostError = fmt.Errorf("database repost failed")
			},
			expectError:   true,
			errorContains: "failed to repost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// For tests that need a post, add the post with the correct ID
			if tt.name == "successful post repost" || tt.name == "database repost error" {
				post := &models.Post{
					ID:      tt.postID,
					Content: "Test content",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test Repost
			err := postService.Repost(context.Background(), tt.userID, tt.postID)
			
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
		})
	}
}// Test GetAllPosts business logic with the REAL PostService using mocks
func TestPostService_GetAllPosts(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		offset         int
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
		expectPosts    bool
		expectedCount  int
	}{
		{
			name:   "successful all posts retrieval with default pagination",
			limit:  0, // Should default to 20
			offset: -1, // Should default to 0
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetAll to return some posts
				postRepo.allPosts = []*models.Post{
					{
						ID:      uuid.New(),
						Content: "Post 1",
					},
					{
						ID:      uuid.New(),
						Content: "Post 2",
					},
				}
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 2,
		},
		{
			name:   "successful all posts retrieval with custom pagination",
			limit:  10,
			offset: 5,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetAll to return some posts
				postRepo.allPosts = []*models.Post{
					{
						ID:      uuid.New(),
						Content: "Post 1",
					},
				}
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 1,
		},
		{
			name:   "limit too high gets capped",
			limit:  150, // Should be capped to 100
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetAll to return some posts
				postRepo.allPosts = []*models.Post{
					{
						ID:      uuid.New(),
						Content: "Post 1",
					},
				}
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 1,
		},
		{
			name:   "no posts found",
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetAll to return empty list
				postRepo.allPosts = []*models.Post{}
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 0,
		},
		{
			name:   "database error",
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Setup repository to return error
				postRepo.getAllError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test GetAllPosts
			posts, err := postService.GetAllPosts(context.Background(), tt.limit, tt.offset)
			
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
				if tt.expectedCount >= 0 && len(posts) != tt.expectedCount {
					t.Errorf("Expected %d posts, got %d", tt.expectedCount, len(posts))
				}
			}
		})
	}
}// Test GetFeed business logic with the REAL PostService using mocks
func TestPostService_GetFeed(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		limit          int
		offset         int
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
		expectPosts    bool
		expectedCount  int
	}{
		{
			name:   "successful feed retrieval",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetFeed to return some posts
				postRepo.feedPosts = []*models.Post{
					{
						ID:      uuid.New(),
						Content: "Feed Post 1",
					},
					{
						ID:      uuid.New(),
						Content: "Feed Post 2",
					},
				}
				
				// Mock like/repost status
				postRepo.isLikedResult = true
				postRepo.isRepostedResult = false
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 2,
		},
		{
			name:   "invalid limit too low",
			userID: uuid.New(),
			limit:  0,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:   "invalid limit too high",
			userID: uuid.New(),
			limit:  101,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:   "invalid offset negative",
			userID: uuid.New(),
			limit:  20,
			offset: -1,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "invalid offset",
		},
		{
			name:   "database feed error",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Setup repository to return feed error
				postRepo.getFeedError = fmt.Errorf("database feed query failed")
			},
			expectError:   true,
			errorContains: "failed to retrieve feed",
		},
		{
			name:   "empty feed",
			userID: uuid.New(),
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetFeed to return empty list
				postRepo.feedPosts = []*models.Post{}
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test GetFeed
			posts, err := postService.GetFeed(context.Background(), tt.userID, tt.limit, tt.offset)
			
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
				if tt.expectedCount >= 0 && len(posts) != tt.expectedCount {
					t.Errorf("Expected %d posts, got %d", tt.expectedCount, len(posts))
				}
				// Verify like/repost status is set for posts
				if len(posts) > 0 {
					for _, post := range posts {
						if !post.IsLiked {
							t.Error("Expected post to be liked")
						}
						if post.IsReposted {
							t.Error("Expected post to not be reposted")
						}
					}
				}
			}
		})
	}
}// Test GetExploreFeed business logic with the REAL PostService using mocks
func TestPostService_GetExploreFeed(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		offset         int
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
		expectPosts    bool
		expectedCount  int
	}{
		{
			name:   "successful explore feed retrieval",
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetFeed to return some posts (using uuid.Nil)
				postRepo.feedPosts = []*models.Post{
					{
						ID:      uuid.New(),
						Content: "Explore Post 1",
					},
					{
						ID:      uuid.New(),
						Content: "Explore Post 2",
					},
				}
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 2,
		},
		{
			name:   "invalid limit too low",
			limit:  0,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:   "invalid limit too high",
			limit:  101,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:   "invalid offset negative",
			limit:  20,
			offset: -1,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// No setup needed for validation error
			},
			expectError:   true,
			errorContains: "invalid offset",
		},
		{
			name:   "database explore feed error",
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Setup repository to return feed error
				postRepo.getFeedError = fmt.Errorf("database explore feed query failed")
			},
			expectError:   true,
			errorContains: "failed to retrieve explore feed",
		},
		{
			name:   "empty explore feed",
			limit:  20,
			offset: 0,
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Mock GetFeed to return empty list
				postRepo.feedPosts = []*models.Post{}
			},
			expectError:   false,
			expectPosts:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// Create REAL PostService with mocks
			postService := NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			
			// Test GetExploreFeed
			posts, err := postService.GetExploreFeed(context.Background(), tt.limit, tt.offset)
			
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
				if tt.expectedCount >= 0 && len(posts) != tt.expectedCount {
					t.Errorf("Expected %d posts, got %d", tt.expectedCount, len(posts))
				}
				// Verify like/repost status is NOT set for explore feed
				if len(posts) > 0 {
					for _, post := range posts {
						if post.IsLiked {
							t.Error("Expected post to not be liked in explore feed")
						}
						if post.IsReposted {
							t.Error("Expected post to not be reposted in explore feed")
						}
					}
				}
			}
		})
	}
}// Test UploadMedia business logic with the REAL PostService using mocks
func TestPostService_UploadMedia(t *testing.T) {
	tests := []struct {
		name           string
		postID         uuid.UUID
		userID         uuid.UUID
		filename       string
		fileData       []byte
		contentType    string
		setupMocks     func(*MockPostRepository, *MockUserRepositoryForPost, *MockBandRepositoryForPost, *MockCache, *MockS3ClientForPost)
		expectError    bool
		errorContains  string
		expectURL      bool
		expectMediaType string
	}{
		{
			name:        "successful image upload",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.jpg",
			fileData:    []byte("fake image data"),
			contentType: "image",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Mock S3 upload success
				s3Client.uploadResult = &storage.UploadResult{
					URL: "https://example.com/image.jpg",
				}
			},
			expectError:     false,
			expectURL:       true,
			expectMediaType: "image",
		},
		{
			name:        "successful audio upload",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.mp3",
			fileData:    []byte("fake audio data"),
			contentType: "audio",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Mock S3 upload success
				s3Client.uploadResult = &storage.UploadResult{
					URL: "https://example.com/audio.mp3",
				}
			},
			expectError:     false,
			expectURL:       true,
			expectMediaType: "audio",
		},
		{
			name:        "S3 client not configured",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.jpg",
			fileData:    []byte("fake image data"),
			contentType: "image",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
				// Don't set s3Client - will be nil
			},
			expectError:   true,
			errorContains: "S3 client not configured",
		},
		{
			name:        "post not found",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.jpg",
			fileData:    []byte("fake image data"),
			contentType: "image",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Don't add post - will cause "post not found" error
			},
			expectError:   true,
			errorContains: "post not found",
		},
		{
			name:        "user not author",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.jpg",
			fileData:    []byte("fake image data"),
			contentType: "image",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post with different user ID
				postID := uuid.New()
				differentUserID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &differentUserID, // Different from test userID
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError:   true,
			errorContains: "you can only add media to your own posts",
		},
		{
			name:        "unsupported media type",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.txt",
			fileData:    []byte("fake text data"),
			contentType: "text",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
			},
			expectError:   true,
			errorContains: "unsupported media type",
		},
		{
			name:        "image validation error",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.jpg",
			fileData:    []byte("fake image data"),
			contentType: "image",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Mock S3 validation error
				s3Client.validateImageError = fmt.Errorf("invalid image format")
			},
			expectError:   true,
			errorContains: "invalid image file",
		},
		{
			name:        "S3 upload error",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.jpg",
			fileData:    []byte("fake image data"),
			contentType: "image",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Mock S3 upload error
				s3Client.uploadError = fmt.Errorf("S3 upload failed")
			},
			expectError:   true,
			errorContains: "failed to upload media",
		},
		{
			name:        "database update error",
			postID:      uuid.New(),
			userID:      uuid.New(),
			filename:    "test.jpg",
			fileData:    []byte("fake image data"),
			contentType: "image",
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockUserRepositoryForPost, bandRepo *MockBandRepositoryForPost, cache *MockCache, s3Client *MockS3ClientForPost) {
				// Add post to mock
				postID := uuid.New()
				userID := uuid.New()
				post := &models.Post{
					ID:      postID,
					UserID:  &userID,
					Content: "Test post",
				}
				postRepo.postsByID[postID.String()] = post
				
				// Mock S3 upload success
				s3Client.uploadResult = &storage.UploadResult{
					URL: "https://example.com/image.jpg",
				}
				
				// Mock database update error
				postRepo.updateError = fmt.Errorf("database update failed")
			},
			expectError:   true,
			errorContains: "failed to update post with media",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			postRepo := NewMockPostRepository()
			userRepo := NewMockUserRepositoryForPost()
			bandRepo := NewMockBandRepositoryForPost()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForPost()
			
			if tt.setupMocks != nil {
				tt.setupMocks(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// For tests that need a post, add the post with the correct IDs
			if tt.name == "successful image upload" || tt.name == "successful audio upload" || tt.name == "image validation error" || tt.name == "S3 upload error" || tt.name == "database update error" || tt.name == "unsupported media type" {
				post := &models.Post{
					ID:      tt.postID,
					UserID:  &tt.userID,
					Content: "Test post",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// For "user not author" test, add post with different user ID
			if tt.name == "user not author" {
				differentUserID := uuid.New()
				post := &models.Post{
					ID:      tt.postID,
					UserID:  &differentUserID,
					Content: "Test post",
				}
				postRepo.postsByID[tt.postID.String()] = post
			}
			
			// Create REAL PostService with mocks
			var postService *PostService
			if tt.name == "S3 client not configured" {
				// For this test, pass nil s3Client
				postService = NewPostService(postRepo, userRepo, bandRepo, cache, nil)
			} else {
				postService = NewPostService(postRepo, userRepo, bandRepo, cache, s3Client)
			}
			
			// Test UploadMedia
			url, mediaType, err := postService.UploadMedia(context.Background(), tt.postID, tt.userID, tt.filename, tt.fileData, tt.contentType)
			
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
				if mediaType != "" {
					t.Error("Expected mediaType to be empty on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectURL && url == "" {
					t.Error("Expected URL but got empty")
				}
				if tt.expectMediaType != "" && mediaType != tt.expectMediaType {
					t.Errorf("Expected mediaType '%s', got '%s'", tt.expectMediaType, mediaType)
				}
			}
		})
	}
}