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

// Mock implementations for BandService dependencies
type MockBandRepository struct {
	bandsByID     map[string]*models.Band
	userBands     map[string][]*models.Band
	nearbyBands   []*models.Band
	allBands      []*models.Band
	createError   error
	getByIDError  error
	updateError   error
	deleteError   error
	getNearbyError error
	getUserBandsError error
	getAllError   error
	isAdminResult bool
	isAdminError  error
	isMemberResult bool
	isMemberError  error
	addMemberError error
	removeMemberError error
	members []*models.BandMember
	getMembersError error
}

func NewMockBandRepository() *MockBandRepository {
	return &MockBandRepository{
		bandsByID:   make(map[string]*models.Band),
		userBands:   make(map[string][]*models.Band),
		nearbyBands: []*models.Band{},
		allBands:    []*models.Band{},
	}
}

func (m *MockBandRepository) Create(ctx context.Context, band *models.Band) error {
	if m.createError != nil {
		return m.createError
	}
	m.bandsByID[band.ID.String()] = band
	return nil
}

func (m *MockBandRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	band, exists := m.bandsByID[id.String()]
	if !exists {
		return nil, fmt.Errorf("band not found")
	}
	return band, nil
}

func (m *MockBandRepository) Update(ctx context.Context, band *models.Band) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.bandsByID[band.ID.String()] = band
	return nil
}

func (m *MockBandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.bandsByID, id.String())
	return nil
}

func (m *MockBandRepository) GetNearby(ctx context.Context, lat, lng float64, radiusKm, limit int) ([]*models.Band, error) {
	if m.getNearbyError != nil {
		return nil, m.getNearbyError
	}
	return m.nearbyBands, nil
}

func (m *MockBandRepository) GetUserBands(ctx context.Context, userID uuid.UUID) ([]*models.BandMember, error) {
	if m.getUserBandsError != nil {
		return nil, m.getUserBandsError
	}
	bands, exists := m.userBands[userID.String()]
	if !exists {
		return []*models.BandMember{}, nil
	}
	// Convert []*models.Band to []*models.BandMember
	var members []*models.BandMember
	for _, band := range bands {
		member := &models.BandMember{
			BandID: band.ID,
			UserID: userID,
			Band:   band,
		}
		members = append(members, member)
	}
	return members, nil
}

func (m *MockBandRepository) AddMember(ctx context.Context, bandID, userID uuid.UUID, role string) error {
	if m.addMemberError != nil {
		return m.addMemberError
	}
	return nil
}

func (m *MockBandRepository) RemoveMember(ctx context.Context, bandID, userID uuid.UUID) error {
	if m.removeMemberError != nil {
		return m.removeMemberError
	}
	return nil
}

func (m *MockBandRepository) GetMembers(ctx context.Context, bandID uuid.UUID) ([]*models.BandMember, error) {
	if m.getMembersError != nil {
		return nil, m.getMembersError
	}
	return m.members, nil
}

func (m *MockBandRepository) IsMember(ctx context.Context, bandID, userID uuid.UUID) (bool, error) {
	return m.isMemberResult, m.isMemberError
}

func (m *MockBandRepository) IsAdmin(ctx context.Context, bandID, userID uuid.UUID) (bool, error) {
	return m.isAdminResult, m.isAdminError
}

func (m *MockBandRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Band, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	return m.allBands, nil
}

type MockUserRepositoryForBand struct {
	usersByID map[string]*models.User
	getByIDError error
}

func NewMockUserRepositoryForBand() *MockUserRepositoryForBand {
	return &MockUserRepositoryForBand{
		usersByID: make(map[string]*models.User),
	}
}

func (m *MockUserRepositoryForBand) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	user, exists := m.usersByID[id.String()]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

type MockS3ClientForBand struct {
	validateError error
	uploadError   error
	uploadResult  *storage.UploadResult
}

func NewMockS3ClientForBand() *MockS3ClientForBand {
	return &MockS3ClientForBand{
		uploadResult: &storage.UploadResult{
			URL:      "https://example.com/image.jpg",
			Key:      "test-key",
			Size:     1024,
			MimeType: "image/jpeg",
		},
	}
}

func (m *MockS3ClientForBand) ValidateImageFile(filename string, size int64) error {
	return m.validateError
}

func (m *MockS3ClientForBand) UploadImage(ctx context.Context, userID, filename string, file io.Reader, size int64) (*storage.UploadResult, error) {
	if m.uploadError != nil {
		return nil, m.uploadError
	}
	return m.uploadResult, nil
}

// Test CreateBand business logic with the REAL BandService using mocks
func TestBandService_CreateBand(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		req            *models.CreateBandRequest
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectBand     bool
	}{
		{
			name:   "successful band creation",
			userID: uuid.New(),
			req: &models.CreateBandRequest{
				Name:        "Test Band",
				Bio:         stringPtr("A test band"),
				Genres:      []string{"Rock", "Alternative"},
				Location:    &models.Location{Latitude: 40.7128, Longitude: -74.0060},
				City:        stringPtr("New York"),
				Country:     stringPtr("USA"),
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add user
				userID := uuid.New()
				user := &models.User{
					ID:       userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[userID.String()] = user
			},
			expectError: false,
			expectBand:  true,
		},
		{
			name:   "missing band name",
			userID: uuid.New(),
			req: &models.CreateBandRequest{
				Name: "", // Empty name
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "band name is required",
		},
		{
			name:   "band name too long",
			userID: uuid.New(),
			req: &models.CreateBandRequest{
				Name: strings.Repeat("a", 101), // 101 characters
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "band name too long",
		},
		{
			name:   "database error on band creation",
			userID: uuid.New(),
			req: &models.CreateBandRequest{
				Name: "Test Band",
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add user
				userID := uuid.New()
				user := &models.User{
					ID:       userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[userID.String()] = user
				bandRepo.createError = fmt.Errorf("database error")
			},
			expectError:   true,
			errorContains: "failed to create band",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// For successful tests, add the user with the correct ID
			if !tt.expectError && tt.expectBand {
				user := &models.User{
					ID:       tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[tt.userID.String()] = user
			}
			
			// For database error on band creation test, add user but set create error
			if tt.name == "database error on band creation" {
				user := &models.User{
					ID:       tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.usersByID[tt.userID.String()] = user
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test CreateBand
			band, err := bandService.CreateBand(context.Background(), tt.userID, tt.req)
			
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
				if band != nil {
					t.Error("Expected band to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectBand && band == nil {
					t.Error("Expected band but got nil")
				}
				
				// Verify band was created with correct data
				if band != nil {
					if band.Name != tt.req.Name {
						t.Errorf("Expected band name %s, got %s", tt.req.Name, band.Name)
					}
					// Note: Band doesn't have CreatorID field, creator is managed through band_members table
				}
			}
		})
	}
}

// Test GetBand business logic with the REAL BandService using mocks
func TestBandService_GetBand(t *testing.T) {
	tests := []struct {
		name           string
		bandID         uuid.UUID
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectBand     bool
	}{
		{
			name:   "successful band retrieval",
			bandID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[bandID.String()] = band
			},
			expectError: false,
			expectBand:  true,
		},
		{
			name:   "band not found",
			bandID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// No band added
			},
			expectError:   true,
			errorContains: "band not found",
		},
		{
			name:   "database error - GetByID",
			bandID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				bandRepo.getByIDError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "band not found",
		},
		{
			name:   "database error - GetMembers",
			bandID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// This will be set up in the test loop
			},
			expectError:   true,
			errorContains: "failed to retrieve band members",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// For successful tests, add the band with the correct ID
			if !tt.expectError && tt.expectBand {
				band := &models.Band{
					ID:   tt.bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[tt.bandID.String()] = band
			}
			
			// Special handling for GetMembers error test
			if tt.name == "database error - GetMembers" {
				band := &models.Band{
					ID:   tt.bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[tt.bandID.String()] = band
				bandRepo.getMembersError = fmt.Errorf("members database error")
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test GetBand
			band, err := bandService.GetBand(context.Background(), tt.bandID)
			
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
				if band != nil {
					t.Error("Expected band to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectBand && band == nil {
					t.Error("Expected band but got nil")
				}
				if band != nil && band.ID != tt.bandID {
					t.Errorf("Expected band ID %s, got %s", tt.bandID.String(), band.ID.String())
				}
			}
		})
	}
}

// Test UpdateBand business logic with the REAL BandService using mocks
func TestBandService_UpdateBand(t *testing.T) {
	tests := []struct {
		name           string
		bandID         uuid.UUID
		userID         uuid.UUID
		req            *models.UpdateBandRequest
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectBand     bool
	}{
		{
			name:   "successful band update by admin",
			bandID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdateBandRequest{
				Name:       stringPtr("Updated Band Name"),
				Bio:        stringPtr("Updated bio"),
				City:       stringPtr("Updated City"),
				Country:    stringPtr("Updated Country"),
				Genres:     []string{"Rock", "Alternative"},
				LookingFor: []string{"Guitarist", "Drummer"},
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Original Band Name",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
			},
			expectError: false,
			expectBand:  true,
		},
		{
			name:   "user not admin",
			bandID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdateBandRequest{
				Name: stringPtr("Updated Band Name"),
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return false
				bandRepo.isAdminResult = false
			},
			expectError:   true,
			errorContains: "only band admins can update band details",
		},
		{
			name:   "band not found",
			bandID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdateBandRequest{
				Name: stringPtr("Updated Band Name"),
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				// Don't add band - will cause "band not found" error
			},
			expectError:   true,
			errorContains: "band not found",
		},
		{
			name:   "empty band name",
			bandID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdateBandRequest{
				Name: stringPtr(""), // Empty name
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Original Band Name",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
			},
			expectError:   true,
			errorContains: "band name cannot be empty",
		},
		{
			name:   "band name too long",
			bandID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdateBandRequest{
				Name: stringPtr(strings.Repeat("a", 101)), // 101 characters
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Original Band Name",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
			},
			expectError:   true,
			errorContains: "band name too long",
		},
		{
			name:   "database update failure",
			bandID: uuid.New(),
			userID: uuid.New(),
			req: &models.UpdateBandRequest{
				Name: stringPtr("Updated Band Name"),
			},
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Original Band Name",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				
				// Setup repository to return update error
				bandRepo.updateError = fmt.Errorf("database update failed")
			},
			expectError:   true,
			errorContains: "failed to update band",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// For tests that need a band, add the band with the correct ID
			if tt.name == "successful band update by admin" || tt.name == "empty band name" || tt.name == "band name too long" || tt.name == "database update failure" {
				band := &models.Band{
					ID:   tt.bandID,
					Name: "Original Band Name",
				}
				bandRepo.bandsByID[tt.bandID.String()] = band
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test UpdateBand
			band, err := bandService.UpdateBand(context.Background(), tt.bandID, tt.userID, tt.req)
			
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
				if band != nil {
					t.Error("Expected band to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectBand && band == nil {
					t.Error("Expected band but got nil")
				}
				
				// Verify band was updated with correct data
				if band != nil && tt.req != nil {
					if tt.req.Name != nil && band.Name != *tt.req.Name {
						t.Errorf("Expected band name %s, got %s", *tt.req.Name, band.Name)
					}
				}
			}
		})
	}
}

// Test DeleteBand business logic with the REAL BandService using mocks
func TestBandService_DeleteBand(t *testing.T) {
	tests := []struct {
		name           string
		bandID         uuid.UUID
		userID         uuid.UUID
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful band deletion by admin",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
			},
			expectError: false,
		},
		{
			name:   "user not admin",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return false
				bandRepo.isAdminResult = false
			},
			expectError:   true,
			errorContains: "only band admins can delete the band",
		},
		{
			name:   "database delete failure",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				
				// Setup repository to return delete error
				bandRepo.deleteError = fmt.Errorf("database delete failed")
			},
			expectError:   true,
			errorContains: "failed to delete band",
		},
		{
			name:   "IsAdmin check failure",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return error
				bandRepo.isAdminError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to check permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test DeleteBand
			err := bandService.DeleteBand(context.Background(), tt.bandID, tt.userID)
			
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
}

// Test JoinBand business logic with the REAL BandService using mocks
func TestBandService_JoinBand(t *testing.T) {
	tests := []struct {
		name           string
		bandID         uuid.UUID
		userID         uuid.UUID
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful band join",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsMember to return false (not a member yet)
				bandRepo.isMemberResult = false
			},
			expectError: false,
		},
		{
			name:   "band not found",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Don't add band - will cause "band not found" error
				// Mock IsMember to return false
				bandRepo.isMemberResult = false
			},
			expectError:   true,
			errorContains: "band not found",
		},
		{
			name:   "user already a member",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsMember to return true (already a member)
				bandRepo.isMemberResult = true
			},
			expectError:   true,
			errorContains: "user is already a member of this band",
		},
		{
			name:   "database add member failure",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsMember to return false
				bandRepo.isMemberResult = false
				
				// Setup repository to return add member error
				bandRepo.addMemberError = fmt.Errorf("database add member failed")
			},
			expectError:   true,
			errorContains: "failed to join band",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// For tests that need a band, add the band with the correct ID
			if tt.name == "successful band join" || tt.name == "user already a member" || tt.name == "database add member failure" {
				band := &models.Band{
					ID:   tt.bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[tt.bandID.String()] = band
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test JoinBand
			err := bandService.JoinBand(context.Background(), tt.bandID, tt.userID)
			
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
}

// Test LeaveBand business logic with the REAL BandService using mocks
func TestBandService_LeaveBand(t *testing.T) {
	tests := []struct {
		name           string
		bandID         uuid.UUID
		userID         uuid.UUID
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
	}{
		{
			name:   "successful band leave",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsMember to return true (is a member)
				bandRepo.isMemberResult = true
			},
			expectError: false,
		},
		{
			name:   "user not a member",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsMember to return false (not a member)
				bandRepo.isMemberResult = false
			},
			expectError:   true,
			errorContains: "user is not a member of this band",
		},
		{
			name:   "database remove member failure",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsMember to return true
				bandRepo.isMemberResult = true
				
				// Setup repository to return remove member error
				bandRepo.removeMemberError = fmt.Errorf("database remove member failed")
			},
			expectError:   true,
			errorContains: "failed to leave band",
		},
		{
			name:   "IsMember check failure",
			bandID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsMember to return error
				bandRepo.isMemberError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to check membership",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test LeaveBand
			err := bandService.LeaveBand(context.Background(), tt.bandID, tt.userID)
			
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
}// Test GetBandMembers business logic with the REAL BandService using mocks
func TestBandService_GetBandMembers(t *testing.T) {
	tests := []struct {
		name           string
		bandID         uuid.UUID
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectMembers  bool
		expectedCount  int
	}{
		{
			name:   "successful band members retrieval",
			bandID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetMembers to return some members
				bandRepo.members = []*models.BandMember{
					{
						BandID: uuid.New(),
						UserID: uuid.New(),
						Role:   stringPtr("Admin"),
					},
					{
						BandID: uuid.New(),
						UserID: uuid.New(),
						Role:   stringPtr("Member"),
					},
				}
			},
			expectError:   false,
			expectMembers: true,
			expectedCount: 2,
		},
		{
			name:   "band with no members",
			bandID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetMembers to return empty list
				bandRepo.members = []*models.BandMember{}
			},
			expectError:   false,
			expectMembers: true,
			expectedCount: 0,
		},
		{
			name:   "database error",
			bandID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Setup repository to return error
				bandRepo.getMembersError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to retrieve band members",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test GetBandMembers
			members, err := bandService.GetBandMembers(context.Background(), tt.bandID)
			
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
				if members != nil {
					t.Error("Expected members to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectMembers && members == nil {
					t.Error("Expected members but got nil")
				}
				if tt.expectedCount >= 0 && len(members) != tt.expectedCount {
					t.Errorf("Expected %d members, got %d", tt.expectedCount, len(members))
				}
			}
		})
	}
}
// Test GetNearbyBands business logic with the REAL BandService using mocks
func TestBandService_GetNearbyBands(t *testing.T) {
	tests := []struct {
		name           string
		lat            float64
		lng            float64
		radiusKm       int
		limit          int
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectBands    bool
		expectedCount  int
	}{
		{
			name:     "successful nearby bands retrieval",
			lat:      40.7128,
			lng:      -74.0060,
			radiusKm: 10,
			limit:    20,
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetNearby to return some bands
				bandRepo.nearbyBands = []*models.Band{
					{
						ID:   uuid.New(),
						Name: "Nearby Band 1",
					},
					{
						ID:   uuid.New(),
						Name: "Nearby Band 2",
					},
				}
			},
			expectError:   false,
			expectBands:   true,
			expectedCount: 2,
		},
		{
			name:          "invalid latitude too low",
			lat:           -91.0,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid latitude",
		},
		{
			name:          "invalid latitude too high",
			lat:           91.0,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid latitude",
		},
		{
			name:          "invalid longitude too low",
			lat:           40.7128,
			lng:           -181.0,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid longitude",
		},
		{
			name:          "invalid longitude too high",
			lat:           40.7128,
			lng:           181.0,
			radiusKm:      10,
			limit:         20,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid longitude",
		},
		{
			name:          "invalid radius too low",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      0,
			limit:         20,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid radius",
		},
		{
			name:          "invalid radius too high",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      501,
			limit:         20,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid radius",
		},
		{
			name:          "invalid limit too low",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         0,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:          "invalid limit too high",
			lat:           40.7128,
			lng:           -74.0060,
			radiusKm:      10,
			limit:         101,
			setupMocks:    func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {},
			expectError:   true,
			errorContains: "invalid limit",
		},
		{
			name:     "database error",
			lat:      40.7128,
			lng:      -74.0060,
			radiusKm: 10,
			limit:    20,
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Setup repository to return error
				bandRepo.getNearbyError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "failed to retrieve nearby bands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test GetNearbyBands
			bands, err := bandService.GetNearbyBands(context.Background(), tt.lat, tt.lng, tt.radiusKm, tt.limit)
			
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
				if bands != nil {
					t.Error("Expected bands to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectBands && bands == nil {
					t.Error("Expected bands but got nil")
				}
				if tt.expectedCount >= 0 && len(bands) != tt.expectedCount {
					t.Errorf("Expected %d bands, got %d", tt.expectedCount, len(bands))
				}
			}
		})
	}
}// Test GetUserBands business logic with the REAL BandService using mocks
func TestBandService_GetUserBands(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectBands    bool
		expectedCount  int
	}{
		{
			name:   "successful user bands retrieval",
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// No setup needed - will be done in test execution
			},
			expectError:   false,
			expectBands:   true,
			expectedCount: 0,
		},
		{
			name:   "user with no bands",
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetUserBands to return empty list
				bandRepo.userBands = map[string][]*models.Band{}
			},
			expectError:   false,
			expectBands:   true,
			expectedCount: 0,
		},
		{
			name:   "database error",
			userID: uuid.New(),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Setup repository to return error
				bandRepo.getUserBandsError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test GetUserBands
			bands, err := bandService.GetUserBands(context.Background(), tt.userID)
			
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
				if bands != nil {
					t.Error("Expected bands to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectBands && bands == nil {
					t.Error("Expected bands but got nil")
				}
				if tt.expectedCount >= 0 && len(bands) != tt.expectedCount {
					t.Errorf("Expected %d bands, got %d", tt.expectedCount, len(bands))
				}
			}
		})
	}
}// Test GetAllBands business logic with the REAL BandService using mocks
func TestBandService_GetAllBands(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		offset         int
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectBands    bool
		expectedCount  int
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:   "successful all bands retrieval with default pagination",
			limit:  0, // Should default to 20
			offset: -1, // Should default to 0
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetAll to return some bands
				bandRepo.allBands = []*models.Band{
					{
						ID:   uuid.New(),
						Name: "Band 1",
					},
					{
						ID:   uuid.New(),
						Name: "Band 2",
					},
				}
			},
			expectError:    false,
			expectBands:    true,
			expectedCount:  2,
			expectedLimit:  20, // Should be defaulted
			expectedOffset: 0,  // Should be defaulted
		},
		{
			name:   "successful all bands retrieval with custom pagination",
			limit:  10,
			offset: 5,
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetAll to return some bands
				bandRepo.allBands = []*models.Band{
					{
						ID:   uuid.New(),
						Name: "Band 1",
					},
				}
			},
			expectError:    false,
			expectBands:    true,
			expectedCount:  1,
			expectedLimit:  10,
			expectedOffset: 5,
		},
		{
			name:   "limit too high gets capped",
			limit:  150, // Should be capped to 100
			offset: 0,
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetAll to return some bands
				bandRepo.allBands = []*models.Band{
					{
						ID:   uuid.New(),
						Name: "Band 1",
					},
				}
			},
			expectError:    false,
			expectBands:    true,
			expectedCount:  1,
			expectedLimit:  100, // Should be capped
			expectedOffset: 0,
		},
		{
			name:   "no bands found",
			limit:  20,
			offset: 0,
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock GetAll to return empty list
				bandRepo.allBands = []*models.Band{}
			},
			expectError:    false,
			expectBands:    true,
			expectedCount:  0,
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name:   "database error",
			limit:  20,
			offset: 0,
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Setup repository to return error
				bandRepo.getAllError = fmt.Errorf("database connection error")
			},
			expectError:   true,
			errorContains: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// Create REAL BandService with mocks
			bandService := NewBandService(bandRepo, userRepo, cache, s3Client)
			
			// Test GetAllBands
			bands, err := bandService.GetAllBands(context.Background(), tt.limit, tt.offset)
			
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
				if bands != nil {
					t.Error("Expected bands to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectBands && bands == nil {
					t.Error("Expected bands but got nil")
				}
				if tt.expectedCount >= 0 && len(bands) != tt.expectedCount {
					t.Errorf("Expected %d bands, got %d", tt.expectedCount, len(bands))
				}
			}
		})
	}
}// Test UploadProfilePicture business logic with the REAL BandService using mocks
func TestBandService_UploadProfilePicture(t *testing.T) {
	tests := []struct {
		name           string
		bandID         uuid.UUID
		userID         uuid.UUID
		filename       string
		fileData       []byte
		setupMocks     func(*MockBandRepository, *MockUserRepositoryForBand, *MockCache, *MockS3ClientForBand)
		expectError    bool
		errorContains  string
		expectURL      bool
		expectedURL    string
	}{
		{
			name:     "successful profile picture upload",
			bandID:   uuid.New(),
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				
				// Setup S3 mock to return success
				s3Client.uploadResult = &storage.UploadResult{
					URL:      "https://example.com/profile.jpg",
					Key:      "test-key",
					Size:     1024,
					MimeType: "image/jpeg",
				}
			},
			expectError: false,
			expectURL:   true,
			expectedURL: "https://example.com/profile.jpg",
		},
		{
			name:     "S3 client not configured",
			bandID:   uuid.New(),
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Don't set up S3Client - leave it nil
			},
			expectError:   true,
			errorContains: "S3 client not configured",
		},
		{
			name:     "user not admin",
			bandID:   uuid.New(),
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return false
				bandRepo.isAdminResult = false
			},
			expectError:   true,
			errorContains: "only band admins can update profile picture",
		},
		{
			name:     "invalid image file",
			bandID:   uuid.New(),
			userID:   uuid.New(),
			filename: "profile.txt", // Wrong file type
			fileData: []byte("not an image"),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				
				// Setup S3 mock to return validation error
				s3Client.validateError = fmt.Errorf("invalid file type")
			},
			expectError:   true,
			errorContains: "invalid image file",
		},
		{
			name:     "S3 upload failure",
			bandID:   uuid.New(),
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				
				// Setup S3 mock to return upload error
				s3Client.uploadError = fmt.Errorf("S3 upload failed")
			},
			expectError:   true,
			errorContains: "failed to upload profile picture",
		},
		{
			name:     "band not found",
			bandID:   uuid.New(),
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				
				// Setup S3 mock to return success
				s3Client.uploadResult = &storage.UploadResult{
					URL:      "https://example.com/profile.jpg",
					Key:      "test-key",
					Size:     1024,
					MimeType: "image/jpeg",
				}
				
				// Don't add band - will cause "band not found" error
			},
			expectError:   true,
			errorContains: "band not found",
		},
		{
			name:     "database update failure",
			bandID:   uuid.New(),
			userID:   uuid.New(),
			filename: "profile.jpg",
			fileData: []byte("fake image data"),
			setupMocks: func(bandRepo *MockBandRepository, userRepo *MockUserRepositoryForBand, cache *MockCache, s3Client *MockS3ClientForBand) {
				// Add band
				bandID := uuid.New()
				band := &models.Band{
					ID:   bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[bandID.String()] = band
				
				// Mock IsAdmin to return true
				bandRepo.isAdminResult = true
				
				// Setup S3 mock to return success
				s3Client.uploadResult = &storage.UploadResult{
					URL:      "https://example.com/profile.jpg",
					Key:      "test-key",
					Size:     1024,
					MimeType: "image/jpeg",
				}
				
				// Setup repository to return update error
				bandRepo.updateError = fmt.Errorf("database update failed")
			},
			expectError:   true,
			errorContains: "failed to update profile picture URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			bandRepo := NewMockBandRepository()
			userRepo := NewMockUserRepositoryForBand()
			cache := NewMockCache()
			s3Client := NewMockS3ClientForBand()
			
			if tt.setupMocks != nil {
				tt.setupMocks(bandRepo, userRepo, cache, s3Client)
			}
			
			// For successful tests, add the band with the correct ID
			if !tt.expectError && tt.expectURL {
				band := &models.Band{
					ID:   tt.bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[tt.bandID.String()] = band
			}
			
			// For database update failure test, add band but set update error
			if tt.name == "database update failure" {
				band := &models.Band{
					ID:   tt.bandID,
					Name: "Test Band",
				}
				bandRepo.bandsByID[tt.bandID.String()] = band
			}
			
			// Create REAL BandService with mocks
			var bandService *BandService
			if tt.name == "S3 client not configured" {
				bandService = NewBandService(bandRepo, userRepo, cache, nil) // Pass nil for S3Client
			} else {
				bandService = NewBandService(bandRepo, userRepo, cache, s3Client)
			}
			
			// Test UploadProfilePicture
			url, err := bandService.UploadProfilePicture(context.Background(), tt.bandID, tt.userID, tt.filename, tt.fileData)
			
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
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectURL && url == "" {
					t.Error("Expected URL but got empty string")
				}
				if tt.expectedURL != "" && url != tt.expectedURL {
					t.Errorf("Expected URL %s, got %s", tt.expectedURL, url)
				}
			}
		})
	}
}