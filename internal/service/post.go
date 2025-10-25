package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"musicapp/internal/interfaces"
	"musicapp/internal/models"
	"musicapp/internal/repository"
	"musicapp/internal/storage"

	"github.com/google/uuid"
)

// PostRepository interface for post data operations
type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error)
	GetByBandID(ctx context.Context, bandID uuid.UUID, limit, offset int) ([]*models.Post, error)
	GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	LikePost(ctx context.Context, userID, postID uuid.UUID) error
	UnlikePost(ctx context.Context, userID, postID uuid.UUID) error
	Repost(ctx context.Context, userID, postID uuid.UUID) error
	IsLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	IsReposted(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Post, error)
}

// UserRepositoryForPost interface for user operations needed by PostService
type UserRepositoryForPost interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// BandRepositoryForPost interface for band operations needed by PostService
type BandRepositoryForPost interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error)
}

// S3ClientForPost interface for S3 operations needed by PostService
type S3ClientForPost interface {
	ValidateImageFile(filename string, size int64) error
	ValidateAudioFile(filename string, size int64) error
	UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error)
	UploadAudio(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error)
}

type PostService struct {
	postRepo PostRepository
	userRepo UserRepositoryForPost
	bandRepo BandRepositoryForPost
	cache    interfaces.Cache
	s3Client S3ClientForPost
}

func NewPostService(postRepo PostRepository, userRepo UserRepositoryForPost, bandRepo BandRepositoryForPost, cache interfaces.Cache, s3Client S3ClientForPost) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
		bandRepo: bandRepo,
		cache:    cache,
		s3Client: s3Client,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(ctx context.Context, userID uuid.UUID, req *models.CreatePostRequest) (*models.Post, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("post content is required")
	}

	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("post content too long (max 2000 characters)")
	}

	// Create post
	post := &models.Post{
		ID:         uuid.New(),
		AuthorID:   &userID,
		AuthorType: "user",
		UserID:     &userID,
		Content:    req.Content,
		MediaURLs:  req.MediaURLs,
		MediaTypes: req.MediaTypes,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Get the created post with counts
	createdPost, err := s.postRepo.GetByID(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created post: %w", err)
	}

	return createdPost, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, postID uuid.UUID, currentUserID *uuid.UUID) (*models.Post, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// Check if current user has liked/reposted this post
	if currentUserID != nil {
		isLiked, _ := s.postRepo.IsLiked(ctx, *currentUserID, postID)
		isReposted, _ := s.postRepo.IsReposted(ctx, *currentUserID, postID)

		post.IsLiked = isLiked
		post.IsReposted = isReposted
	}

	return post, nil
}

// UpdatePost updates a post
func (s *PostService) UpdatePost(ctx context.Context, postID, userID uuid.UUID, req *models.UpdatePostRequest) (*models.Post, error) {
	// Get existing post
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// Check if user is the author
	if post.UserID == nil || *post.UserID != userID {
		return nil, fmt.Errorf("you can only update your own posts")
	}

	// Update fields
	if req.Content != nil {
		if *req.Content == "" {
			return nil, fmt.Errorf("post content cannot be empty")
		}
		if len(*req.Content) > 2000 {
			return nil, fmt.Errorf("post content too long (max 2000 characters)")
		}
		post.Content = *req.Content
	}
	if req.MediaURLs != nil {
		post.MediaURLs = req.MediaURLs
	}
	if req.MediaTypes != nil {
		post.MediaTypes = req.MediaTypes
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return post, nil
}

// DeletePost deletes a post
func (s *PostService) DeletePost(ctx context.Context, postID, userID uuid.UUID) error {
	// Get existing post
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	// Check if user is the author
	if post.UserID == nil || *post.UserID != userID {
		return fmt.Errorf("you can only delete your own posts")
	}

	if err := s.postRepo.Delete(ctx, postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

// LikePost likes a post
func (s *PostService) LikePost(ctx context.Context, userID, postID uuid.UUID) error {
	// Check if post exists
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	if err := s.postRepo.LikePost(ctx, userID, postID); err != nil {
		return fmt.Errorf("failed to like post: %w", err)
	}

	return nil
}

// UnlikePost unlikes a post
func (s *PostService) UnlikePost(ctx context.Context, userID, postID uuid.UUID) error {
	if err := s.postRepo.UnlikePost(ctx, userID, postID); err != nil {
		return fmt.Errorf("failed to unlike post: %w", err)
	}

	return nil
}

// Repost reposts a post
func (s *PostService) Repost(ctx context.Context, userID, postID uuid.UUID) error {
	// Check if post exists
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	if err := s.postRepo.Repost(ctx, userID, postID); err != nil {
		return fmt.Errorf("failed to repost: %w", err)
	}

	return nil
}

// GetFeed retrieves personalized feed for a user
func (s *PostService) GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d (must be >= 0)", offset)
	}

	posts, err := s.postRepo.GetFeed(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve feed: %w", err)
	}

	// Add like/repost status for current user
	for _, post := range posts {
		isLiked, _ := s.postRepo.IsLiked(ctx, userID, post.ID)
		isReposted, _ := s.postRepo.IsReposted(ctx, userID, post.ID)

		post.IsLiked = isLiked
		post.IsReposted = isReposted
	}

	return posts, nil
}

// GetExploreFeed retrieves explore/trending posts
func (s *PostService) GetExploreFeed(ctx context.Context, limit, offset int) ([]*models.Post, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d (must be >= 0)", offset)
	}

	// Use empty UUID for explore feed
	posts, err := s.postRepo.GetFeed(ctx, uuid.Nil, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve explore feed: %w", err)
	}

	return posts, nil
}

// UploadMedia uploads media files to a post
func (s *PostService) UploadMedia(ctx context.Context, postID, userID uuid.UUID, filename string, fileData []byte, contentType string) (string, string, error) {
	if s.s3Client == nil {
		return "", "", fmt.Errorf("S3 client not configured")
	}

	// Get existing post
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return "", "", fmt.Errorf("post not found: %w", err)
	}

	// Check if user is the author
	if post.UserID == nil || *post.UserID != userID {
		return "", "", fmt.Errorf("you can only add media to your own posts")
	}

	var mediaType string
	var uploadResult *storage.UploadResult

	// Determine media type and upload
	if contentType == "image" {
		mediaType = "image"
		if err := s.s3Client.ValidateImageFile(filename, int64(len(fileData))); err != nil {
			return "", "", fmt.Errorf("invalid image file: %w", err)
		}
		uploadResult, err = s.s3Client.UploadImage(ctx, userID.String(), filename, bytes.NewReader(fileData), int64(len(fileData)))
	} else if contentType == "audio" {
		mediaType = "audio"
		if err := s.s3Client.ValidateAudioFile(filename, int64(len(fileData))); err != nil {
			return "", "", fmt.Errorf("invalid audio file: %w", err)
		}
		uploadResult, err = s.s3Client.UploadAudio(ctx, userID.String(), filename, bytes.NewReader(fileData), int64(len(fileData)))
	} else {
		return "", "", fmt.Errorf("unsupported media type: %s", contentType)
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to upload media: %w", err)
	}

	// Update post with new media
	post.MediaURLs = append(post.MediaURLs, uploadResult.URL)
	post.MediaTypes = append(post.MediaTypes, mediaType)

	if err := s.postRepo.Update(ctx, post); err != nil {
		return "", "", fmt.Errorf("failed to update post with media: %w", err)
	}

	return uploadResult.URL, mediaType, nil
}

// GetAllPosts gets all posts with pagination
func (s *PostService) GetAllPosts(ctx context.Context, limit, offset int) ([]*models.Post, error) {
	// Default pagination values
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit to prevent abuse
	}
	if offset < 0 {
		offset = 0
	}

	return s.postRepo.GetAll(ctx, limit, offset)
}

// Adapter structs to bridge existing concrete types with new interfaces

// PostRepositoryAdapter adapts *repository.PostRepository to PostRepository interface
type PostRepositoryAdapter struct {
	*repository.PostRepository
}

func NewPostRepositoryAdapter(repo *repository.PostRepository) PostRepository {
	return &PostRepositoryAdapter{repo}
}

// UserRepositoryForPostAdapter adapts *repository.UserRepository to UserRepositoryForPost interface
type UserRepositoryForPostAdapter struct {
	*repository.UserRepository
}

func NewUserRepositoryForPostAdapter(repo *repository.UserRepository) UserRepositoryForPost {
	return &UserRepositoryForPostAdapter{repo}
}

// BandRepositoryForPostAdapter adapts *repository.BandRepository to BandRepositoryForPost interface
type BandRepositoryForPostAdapter struct {
	*repository.BandRepository
}

func NewBandRepositoryForPostAdapter(repo *repository.BandRepository) BandRepositoryForPost {
	return &BandRepositoryForPostAdapter{repo}
}

// S3ClientForPostAdapter adapts *storage.S3Client to S3ClientForPost interface
type S3ClientForPostAdapter struct {
	*storage.S3Client
}

func NewS3ClientForPostAdapter(client *storage.S3Client) S3ClientForPost {
	return &S3ClientForPostAdapter{client}
}
