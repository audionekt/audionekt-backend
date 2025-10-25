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

// Extended interfaces for BandService (building on existing interfaces from auth.go)
type BandRepository interface {
	Create(ctx context.Context, band *models.Band) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error)
	Update(ctx context.Context, band *models.Band) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetNearby(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.Band, error)
	AddMember(ctx context.Context, bandID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, bandID, userID uuid.UUID) error
	GetMembers(ctx context.Context, bandID uuid.UUID) ([]*models.BandMember, error)
	IsMember(ctx context.Context, bandID, userID uuid.UUID) (bool, error)
	IsAdmin(ctx context.Context, bandID, userID uuid.UUID) (bool, error)
	GetUserBands(ctx context.Context, userID uuid.UUID) ([]*models.BandMember, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Band, error)
}

type UserRepositoryForBand interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type S3ClientForBand interface {
	ValidateImageFile(filename string, size int64) error
	UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error)
}

// BandService now depends on interfaces, not concrete types
type BandService struct {
	bandRepo BandRepository
	userRepo UserRepositoryForBand
	cache    interfaces.Cache
	s3Client S3ClientForBand
}

func NewBandService(bandRepo BandRepository, userRepo UserRepositoryForBand, cache interfaces.Cache, s3Client S3ClientForBand) *BandService {
	return &BandService{
		bandRepo: bandRepo,
		userRepo: userRepo,
		cache:    cache,
		s3Client: s3Client,
	}
}

// CreateBand creates a new band
func (s *BandService) CreateBand(ctx context.Context, userID uuid.UUID, req *models.CreateBandRequest) (*models.Band, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("band name is required")
	}

	if len(req.Name) > 100 {
		return nil, fmt.Errorf("band name too long (max 100 characters)")
	}

	// Create band
	band := &models.Band{
		ID:         uuid.New(),
		Name:       req.Name,
		Bio:        req.Bio,
		Location:   req.Location,
		City:       req.City,
		Country:    req.Country,
		Genres:     req.Genres,
		LookingFor: req.LookingFor,
	}

	if err := s.bandRepo.Create(ctx, band); err != nil {
		return nil, fmt.Errorf("failed to create band: %w", err)
	}

	// Add creator as admin member
	if err := s.bandRepo.AddMember(ctx, band.ID, userID, "Admin"); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	return band, nil
}

// GetBand retrieves a band by ID
func (s *BandService) GetBand(ctx context.Context, bandID uuid.UUID) (*models.Band, error) {
	band, err := s.bandRepo.GetByID(ctx, bandID)
	if err != nil {
		return nil, fmt.Errorf("band not found: %w", err)
	}

	// Get members
	members, err := s.bandRepo.GetMembers(ctx, bandID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve band members: %w", err)
	}

	// Add members to response (this would need to be handled in the response model)
	_ = members // For now, just ignore members in service layer

	return band, nil
}

// UpdateBand updates band details
func (s *BandService) UpdateBand(ctx context.Context, bandID, userID uuid.UUID, req *models.UpdateBandRequest) (*models.Band, error) {
	// Check if user is admin of the band
	isAdmin, err := s.bandRepo.IsAdmin(ctx, bandID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !isAdmin {
		return nil, fmt.Errorf("only band admins can update band details")
	}

	// Get existing band
	band, err := s.bandRepo.GetByID(ctx, bandID)
	if err != nil {
		return nil, fmt.Errorf("band not found: %w", err)
	}

	// Update fields
	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("band name cannot be empty")
		}
		if len(*req.Name) > 100 {
			return nil, fmt.Errorf("band name too long (max 100 characters)")
		}
		band.Name = *req.Name
	}
	if req.Bio != nil {
		band.Bio = req.Bio
	}
	if req.Location != nil {
		band.Location = req.Location
	}
	if req.City != nil {
		band.City = req.City
	}
	if req.Country != nil {
		band.Country = req.Country
	}
	if req.Genres != nil {
		band.Genres = req.Genres
	}
	if req.LookingFor != nil {
		band.LookingFor = req.LookingFor
	}

	if err := s.bandRepo.Update(ctx, band); err != nil {
		return nil, fmt.Errorf("failed to update band: %w", err)
	}

	return band, nil
}

// DeleteBand deletes a band
func (s *BandService) DeleteBand(ctx context.Context, bandID, userID uuid.UUID) error {
	// Check if user is admin of the band
	isAdmin, err := s.bandRepo.IsAdmin(ctx, bandID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if !isAdmin {
		return fmt.Errorf("only band admins can delete the band")
	}

	if err := s.bandRepo.Delete(ctx, bandID); err != nil {
		return fmt.Errorf("failed to delete band: %w", err)
	}

	return nil
}

// JoinBand adds a user to a band
func (s *BandService) JoinBand(ctx context.Context, bandID, userID uuid.UUID) error {
	// Check if band exists
	_, err := s.bandRepo.GetByID(ctx, bandID)
	if err != nil {
		return fmt.Errorf("band not found: %w", err)
	}

	// Check if user is already a member
	isMember, err := s.bandRepo.IsMember(ctx, bandID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	if isMember {
		return fmt.Errorf("user is already a member of this band")
	}

	// Add user as member
	if err := s.bandRepo.AddMember(ctx, bandID, userID, "Member"); err != nil {
		return fmt.Errorf("failed to join band: %w", err)
	}

	return nil
}

// LeaveBand removes a user from a band
func (s *BandService) LeaveBand(ctx context.Context, bandID, userID uuid.UUID) error {
	// Check if user is a member
	isMember, err := s.bandRepo.IsMember(ctx, bandID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	if !isMember {
		return fmt.Errorf("user is not a member of this band")
	}

	// Remove user from band
	if err := s.bandRepo.RemoveMember(ctx, bandID, userID); err != nil {
		return fmt.Errorf("failed to leave band: %w", err)
	}

	return nil
}

// GetBandMembers retrieves all members of a band
func (s *BandService) GetBandMembers(ctx context.Context, bandID uuid.UUID) ([]*models.BandMember, error) {
	members, err := s.bandRepo.GetMembers(ctx, bandID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve band members: %w", err)
	}

	return members, nil
}

// GetNearbyBands finds bands within a specified radius
func (s *BandService) GetNearbyBands(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.Band, error) {
	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("invalid latitude: %f", lat)
	}
	if lng < -180 || lng > 180 {
		return nil, fmt.Errorf("invalid longitude: %f", lng)
	}
	if radiusKm <= 0 || radiusKm > 500 {
		return nil, fmt.Errorf("invalid radius: %d km (must be 1-500)", radiusKm)
	}
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid limit: %d (must be 1-100)", limit)
	}

	bands, err := s.bandRepo.GetNearby(ctx, lat, lng, radiusKm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve nearby bands: %w", err)
	}

	return bands, nil
}

// UploadProfilePicture uploads a band profile picture to S3
func (s *BandService) UploadProfilePicture(ctx context.Context, bandID, userID uuid.UUID, filename string, fileData []byte) (string, error) {
	if s.s3Client == nil {
		return "", fmt.Errorf("S3 client not configured")
	}

	// Check if user is admin of the band
	isAdmin, err := s.bandRepo.IsAdmin(ctx, bandID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to check permissions: %w", err)
	}

	if !isAdmin {
		return "", fmt.Errorf("only band admins can update profile picture")
	}

	// Validate image file
	if err := s.s3Client.ValidateImageFile(filename, int64(len(fileData))); err != nil {
		return "", fmt.Errorf("invalid image file: %w", err)
	}

	// Upload to S3
	uploadResult, err := s.s3Client.UploadImage(ctx, bandID.String(), filename, bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return "", fmt.Errorf("failed to upload profile picture: %w", err)
	}

	// Update band profile picture URL
	band, err := s.bandRepo.GetByID(ctx, bandID)
	if err != nil {
		return "", fmt.Errorf("band not found: %w", err)
	}

	band.ProfilePictureURL = &uploadResult.URL
	if err := s.bandRepo.Update(ctx, band); err != nil {
		return "", fmt.Errorf("failed to update profile picture URL: %w", err)
	}

	return uploadResult.URL, nil
}

// GetUserBands gets all bands that a user is a member of
func (s *BandService) GetUserBands(ctx context.Context, userID uuid.UUID) ([]*models.BandMember, error) {
	return s.bandRepo.GetUserBands(ctx, userID)
}

// GetAllBands gets all bands with pagination
func (s *BandService) GetAllBands(ctx context.Context, limit, offset int) ([]*models.Band, error) {
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

	return s.bandRepo.GetAll(ctx, limit, offset)
}

// Adapter structs implement interfaces for existing concrete types
// This allows us to use the existing repository/cache/s3 with the new interface-based service

// BandRepositoryAdapter adapts repository.BandRepository to BandRepository interface
type BandRepositoryAdapter struct {
	repo *repository.BandRepository
}

func NewBandRepositoryAdapter(repo *repository.BandRepository) BandRepository {
	return &BandRepositoryAdapter{repo: repo}
}

func (a *BandRepositoryAdapter) Create(ctx context.Context, band *models.Band) error {
	return a.repo.Create(ctx, band)
}

func (a *BandRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *BandRepositoryAdapter) Update(ctx context.Context, band *models.Band) error {
	return a.repo.Update(ctx, band)
}

func (a *BandRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.repo.Delete(ctx, id)
}

func (a *BandRepositoryAdapter) GetNearby(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.Band, error) {
	return a.repo.GetNearby(ctx, lat, lng, radiusKm, limit)
}

func (a *BandRepositoryAdapter) GetUserBands(ctx context.Context, userID uuid.UUID) ([]*models.BandMember, error) {
	return a.repo.GetUserBands(ctx, userID)
}

func (a *BandRepositoryAdapter) AddMember(ctx context.Context, bandID, userID uuid.UUID, role string) error {
	return a.repo.AddMember(ctx, bandID, userID, role)
}

func (a *BandRepositoryAdapter) RemoveMember(ctx context.Context, bandID, userID uuid.UUID) error {
	return a.repo.RemoveMember(ctx, bandID, userID)
}

func (a *BandRepositoryAdapter) GetMembers(ctx context.Context, bandID uuid.UUID) ([]*models.BandMember, error) {
	return a.repo.GetMembers(ctx, bandID)
}

func (a *BandRepositoryAdapter) IsMember(ctx context.Context, bandID, userID uuid.UUID) (bool, error) {
	return a.repo.IsMember(ctx, bandID, userID)
}

func (a *BandRepositoryAdapter) IsAdmin(ctx context.Context, bandID, userID uuid.UUID) (bool, error) {
	return a.repo.IsAdmin(ctx, bandID, userID)
}

func (a *BandRepositoryAdapter) GetAll(ctx context.Context, limit, offset int) ([]*models.Band, error) {
	return a.repo.GetAll(ctx, limit, offset)
}

// UserRepositoryForBandAdapter adapts repository.UserRepository to UserRepositoryForBand interface
type UserRepositoryForBandAdapter struct {
	repo *repository.UserRepository
}

func NewUserRepositoryForBandAdapter(repo *repository.UserRepository) UserRepositoryForBand {
	return &UserRepositoryForBandAdapter{repo: repo}
}

func (a *UserRepositoryForBandAdapter) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return a.repo.GetByID(ctx, id)
}

// S3ClientForBandAdapter adapts storage.S3Client to S3ClientForBand interface
type S3ClientForBandAdapter struct {
	s3Client *storage.S3Client
}

func NewS3ClientForBandAdapter(s3Client *storage.S3Client) S3ClientForBand {
	return &S3ClientForBandAdapter{s3Client: s3Client}
}

func (a *S3ClientForBandAdapter) ValidateImageFile(filename string, size int64) error {
	return a.s3Client.ValidateImageFile(filename, size)
}

func (a *S3ClientForBandAdapter) UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error) {
	return a.s3Client.UploadImage(ctx, userID, filename, file, size)
}
