package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client     *s3.Client
	uploader   *manager.Uploader
	bucketName string
	cdnURL     string
}

type UploadResult struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

type AudioMetadata struct {
	Duration   float64 `json:"duration,omitempty"`
	Bitrate    int     `json:"bitrate,omitempty"`
	SampleRate int     `json:"sample_rate,omitempty"`
	Channels   int     `json:"channels,omitempty"`
	Format     string  `json:"format,omitempty"`
}

func NewS3Client(region, bucketName, cdnURL string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &S3Client{
		client:     s3.NewFromConfig(cfg),
		uploader:   manager.NewUploader(s3.NewFromConfig(cfg)),
		bucketName: bucketName,
		cdnURL:     cdnURL,
	}, nil
}

// UploadFile uploads a file to S3 and returns the public URL
func (s *S3Client) UploadFile(ctx context.Context, key string, file io.Reader, contentType string, size int64) (*UploadResult, error) {
	// Determine content type if not provided
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// Upload to S3
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		// ACL removed - bucket has ACLs disabled
		// Add metadata
		Metadata: map[string]string{
			"uploaded-at": time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate CDN URL
	url := s.generateURL(key)

	return &UploadResult{
		URL:      url,
		Key:      key,
		Size:     size,
		MimeType: contentType,
	}, nil
}

// UploadAudio uploads an audio file with optimized settings
func (s *S3Client) UploadAudio(ctx context.Context, userID, filename string, file io.Reader, size int64) (*UploadResult, error) {
	// Generate unique key for audio file
	ext := strings.ToLower(filepath.Ext(filename))
	timestamp := time.Now().Unix()
	key := fmt.Sprintf("audio/%s/%d_%s", userID, timestamp, filename)

	// Determine content type for audio
	contentType := s.getAudioContentType(ext)

	// Upload with audio-specific settings
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		// ACL removed - bucket has ACLs disabled
		// Audio-specific metadata
		Metadata: map[string]string{
			"uploaded-at": time.Now().UTC().Format(time.RFC3339),
			"file-type":   "audio",
			"user-id":     userID,
		},
		// Cache control for audio files (1 year)
		CacheControl: aws.String("public, max-age=31536000"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio file: %w", err)
	}

	url := s.generateURL(key)

	return &UploadResult{
		URL:      url,
		Key:      key,
		Size:     size,
		MimeType: contentType,
	}, nil
}

// UploadImage uploads an image file with optimized settings
func (s *S3Client) UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*UploadResult, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	timestamp := time.Now().Unix()
	key := fmt.Sprintf("images/%s/%d_%s", userID, timestamp, filename)

	contentType := s.getImageContentType(ext)

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		// ACL removed - bucket has ACLs disabled
		Metadata: map[string]string{
			"uploaded-at": time.Now().UTC().Format(time.RFC3339),
			"file-type":   "image",
			"user-id":     userID,
		},
		// Cache control for images (1 month)
		CacheControl: aws.String("public, max-age=2592000"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload image file: %w", err)
	}

	url := s.generateURL(key)

	return &UploadResult{
		URL:      url,
		Key:      key,
		Size:     size,
		MimeType: contentType,
	}, nil
}

// DeleteFile deletes a file from S3
func (s *S3Client) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// GetFileInfo gets metadata about a file
func (s *S3Client) GetFileInfo(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	return s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
}

// Generate presigned URL for secure uploads (optional)
func (s *S3Client) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		// ACL removed - bucket has ACLs disabled
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", err
	}

	return request.URL, nil
}

// Helper methods
func (s *S3Client) generateURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", s.cdnURL, key)
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, key)
}

func (s *S3Client) getAudioContentType(ext string) string {
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	default:
		return "audio/mpeg"
	}
}

func (s *S3Client) getImageContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

// ValidateAudioFile validates audio file before upload
func (s *S3Client) ValidateAudioFile(filename string, size int64) error {
	ext := strings.ToLower(filepath.Ext(filename))

	// Check file extension
	validExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"}
	isValid := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid audio file format: %s", ext)
	}

	// Check file size (100MB max for audio)
	maxSize := int64(100 * 1024 * 1024)
	if size > maxSize {
		return fmt.Errorf("audio file too large: %d bytes (max: %d bytes)", size, maxSize)
	}

	return nil
}

// ValidateImageFile validates image file before upload
func (s *S3Client) ValidateImageFile(filename string, size int64) error {
	ext := strings.ToLower(filepath.Ext(filename))

	// Check file extension
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isValid := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid image file format: %s", ext)
	}

	// Check file size (10MB max for images)
	maxSize := int64(10 * 1024 * 1024)
	if size > maxSize {
		return fmt.Errorf("image file too large: %d bytes (max: %d bytes)", size, maxSize)
	}

	return nil
}
