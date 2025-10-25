package seeder

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"musicapp/internal/db"
	"musicapp/internal/models"
	"musicapp/internal/repository"
	"musicapp/pkg/utils"
)

// Seeder handles database seeding
type Seeder struct {
	db         *db.DB
	userRepo   *repository.UserRepository
	bandRepo   *repository.BandRepository
	postRepo   *repository.PostRepository
	followRepo *repository.FollowRepository
}

// New creates a new seeder instance
func New(database *db.DB) *Seeder {
	return &Seeder{
		db:         database,
		userRepo:   repository.NewUserRepository(database),
		bandRepo:   repository.NewBandRepository(database),
		postRepo:   repository.NewPostRepository(database),
		followRepo: repository.NewFollowRepository(database),
	}
}

// SeedAll seeds the database with comprehensive fake data
func (s *Seeder) SeedAll(ctx context.Context) error {
	log.Println("🌱 Starting database seeding...")

	// Set random seed for consistent results
	gofakeit.Seed(time.Now().UnixNano())

	// Create users
	users, err := s.seedUsers(ctx, 50)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}
	log.Printf("✅ Created %d users", len(users))

	// Create bands
	bands, err := s.seedBands(ctx, 20, users)
	if err != nil {
		return fmt.Errorf("failed to seed bands: %w", err)
	}
	log.Printf("✅ Created %d bands", len(bands))

	// Create posts
	posts, err := s.seedPosts(ctx, 200, users, bands)
	if err != nil {
		return fmt.Errorf("failed to seed posts: %w", err)
	}
	log.Printf("✅ Created %d posts", len(posts))

	// Create follows
	err = s.seedFollows(ctx, users, bands)
	if err != nil {
		return fmt.Errorf("failed to seed follows: %w", err)
	}
	log.Printf("✅ Created follows")

	// Create likes
	err = s.seedLikes(ctx, users, posts)
	if err != nil {
		return fmt.Errorf("failed to seed likes: %w", err)
	}
	log.Printf("✅ Created likes")

	// Create reposts
	err = s.seedReposts(ctx, users, posts)
	if err != nil {
		return fmt.Errorf("failed to seed reposts: %w", err)
	}
	log.Printf("✅ Created reposts")

	log.Println("🎉 Database seeding completed successfully!")
	return nil
}

// seedUsers creates fake users
func (s *Seeder) seedUsers(ctx context.Context, count int) ([]*models.User, error) {
	var users []*models.User
	musicGenres := []string{"Hip-Hop", "Electronic", "Rock", "Jazz", "Pop", "R&B", "Trap", "EDM", "Indie", "Folk"}
	musicSkills := []string{"Producer", "Vocalist", "Guitarist", "Drummer", "Bassist", "Pianist", "DJ", "Sound Engineer", "Mixing", "Mastering"}

	for i := 0; i < count; i++ {
		user := &models.User{
			ID:       uuid.New(),
			Username: gofakeit.Username(),
			Email:    gofakeit.Email(),
			PasswordHash: func() string {
				hash, _ := utils.HashPassword("password123")
				return hash
			}(),
			DisplayName:     stringPtr(gofakeit.Name()),
			Bio:             stringPtr(gofakeit.HipsterSentence()),
			City:            stringPtr(gofakeit.City()),
			Country:         stringPtr(gofakeit.Country()),
			Genres:          randomSubset(musicGenres, 1, 4),
			Skills:          randomSubset(musicSkills, 1, 3),
			SpotifyURL:      stringPtr(fmt.Sprintf("https://open.spotify.com/artist/%s", gofakeit.UUID())),
			SoundcloudURL:   stringPtr(fmt.Sprintf("https://soundcloud.com/%s", gofakeit.Username())),
			InstagramHandle: stringPtr(gofakeit.Username()),
		}

		// Add random location (San Francisco area)
		user.Location = &models.Location{
			Latitude:  37.7749 + (gofakeit.Float64Range(-0.5, 0.5)), // SF area
			Longitude: -122.4194 + (gofakeit.Float64Range(-0.5, 0.5)),
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// seedBands creates fake bands
func (s *Seeder) seedBands(ctx context.Context, count int, users []*models.User) ([]*models.Band, error) {
	var bands []*models.Band
	musicGenres := []string{"Hip-Hop", "Electronic", "Rock", "Jazz", "Pop", "R&B", "Trap", "EDM", "Indie", "Folk"}
	lookingFor := []string{"Vocalist", "Guitarist", "Drummer", "Bassist", "Pianist", "DJ", "Sound Engineer", "Producer"}

	for i := 0; i < count; i++ {
		band := &models.Band{
			ID:         uuid.New(),
			Name:       gofakeit.Company() + " Collective",
			Bio:        stringPtr(gofakeit.HipsterSentence()),
			City:       stringPtr(gofakeit.City()),
			Country:    stringPtr(gofakeit.Country()),
			Genres:     randomSubset(musicGenres, 1, 3),
			LookingFor: randomSubset(lookingFor, 1, 4),
		}

		// Add random location
		band.Location = &models.Location{
			Latitude:  37.7749 + (gofakeit.Float64Range(-0.5, 0.5)),
			Longitude: -122.4194 + (gofakeit.Float64Range(-0.5, 0.5)),
		}

		if err := s.bandRepo.Create(ctx, band); err != nil {
			return nil, err
		}

		// Add random members to the band
		memberCount := gofakeit.IntRange(2, 6)
		selectedUsers := randomSubsetUsers(users, memberCount)

		for j, user := range selectedUsers {
			role := "Member"
			if j == 0 {
				role = "Admin" // First member is admin
			} else if gofakeit.Bool() {
				role = randomChoice([]string{"Producer", "Vocalist", "Guitarist", "Drummer"})
			}

			if err := s.bandRepo.AddMember(ctx, band.ID, user.ID, role); err != nil {
				return nil, err
			}
		}

		bands = append(bands, band)
	}

	return bands, nil
}

// seedPosts creates fake posts
func (s *Seeder) seedPosts(ctx context.Context, count int, users []*models.User, bands []*models.Band) ([]*models.Post, error) {
	var posts []*models.Post
	postContents := []string{
		"Just dropped a new beat! Check it out 🎵",
		"Working on some new material in the studio today",
		"Looking for a vocalist to collaborate on this track",
		"New EP coming soon! Stay tuned 🔥",
		"Had an amazing jam session with the band last night",
		"Just finished mixing this track, what do you think?",
		"Looking for a drummer for our upcoming gig",
		"New single out now on all platforms!",
		"Studio session vibes ✨",
		"Working on some experimental sounds today",
	}

	for i := 0; i < count; i++ {
		// Randomly choose between user post and band post
		isBandPost := gofakeit.Bool()

		var authorID *uuid.UUID
		var authorType string
		var userID *uuid.UUID
		var bandID *uuid.UUID

		if isBandPost && len(bands) > 0 {
			band := randomChoice(bands)
			authorID = &band.ID
			authorType = "band"
			bandID = &band.ID
		} else {
			user := randomChoice(users)
			authorID = &user.ID
			authorType = "user"
			userID = &user.ID
		}

		post := &models.Post{
			ID:         uuid.New(),
			AuthorID:   authorID,
			AuthorType: authorType,
			UserID:     userID,
			BandID:     bandID,
			Content:    randomChoice(postContents),
			MediaURLs:  []string{}, // We'll add some fake media URLs
			MediaTypes: []string{},
		}

		// Add some fake media URLs occasionally
		if gofakeit.Bool() {
			mediaType := randomChoice([]string{"image", "audio"})
			post.MediaURLs = []string{fmt.Sprintf("https://example.com/media/%s", gofakeit.UUID())}
			post.MediaTypes = []string{mediaType}
		}

		if err := s.postRepo.Create(ctx, post); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// seedFollows creates fake follow relationships
func (s *Seeder) seedFollows(ctx context.Context, users []*models.User, bands []*models.Band) error {
	// Users follow other users
	for _, user := range users {
		followCount := gofakeit.IntRange(5, 20)
		usersToFollow := randomSubsetUsers(users, followCount)

		for _, userToFollow := range usersToFollow {
			if user.ID != userToFollow.ID { // Don't follow yourself
				follow := &models.Follow{
					FollowerID:      user.ID,
					FollowingType:   "user",
					FollowingUserID: &userToFollow.ID,
				}
				if err := s.followRepo.Create(ctx, follow); err != nil {
					// Ignore duplicate follow errors
					continue
				}
			}
		}

		// Users follow bands
		bandFollowCount := gofakeit.IntRange(2, 8)
		bandsToFollow := randomSubsetBands(bands, bandFollowCount)

		for _, band := range bandsToFollow {
			follow := &models.Follow{
				FollowerID:      user.ID,
				FollowingType:   "band",
				FollowingBandID: &band.ID,
			}
			if err := s.followRepo.Create(ctx, follow); err != nil {
				// Ignore duplicate follow errors
				continue
			}
		}
	}

	return nil
}

// seedLikes creates fake likes on posts
func (s *Seeder) seedLikes(ctx context.Context, users []*models.User, posts []*models.Post) error {
	for _, post := range posts {
		// Each post gets liked by 0-15 random users
		likeCount := gofakeit.IntRange(0, 15)
		usersToLike := randomSubsetUsers(users, likeCount)

		for _, user := range usersToLike {
			if err := s.postRepo.LikePost(ctx, user.ID, post.ID); err != nil {
				// Ignore duplicate like errors
				continue
			}
		}
	}

	return nil
}

// seedReposts creates fake reposts
func (s *Seeder) seedReposts(ctx context.Context, users []*models.User, posts []*models.Post) error {
	for _, post := range posts {
		// Each post gets reposted by 0-5 random users
		repostCount := gofakeit.IntRange(0, 5)
		usersToRepost := randomSubsetUsers(users, repostCount)

		for _, user := range usersToRepost {
			if err := s.postRepo.Repost(ctx, user.ID, post.ID); err != nil {
				// Ignore duplicate repost errors
				continue
			}
		}
	}

	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func randomChoice[T any](slice []T) T {
	return slice[gofakeit.IntRange(0, len(slice)-1)]
}

func randomSubset[T any](slice []T, min, max int) []T {
	count := gofakeit.IntRange(min, max)
	if count > len(slice) {
		count = len(slice)
	}

	// Shuffle and take first count elements
	shuffled := make([]T, len(slice))
	copy(shuffled, slice)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

func randomSubsetUsers(users []*models.User, count int) []*models.User {
	if count > len(users) {
		count = len(users)
	}

	shuffled := make([]*models.User, len(users))
	copy(shuffled, users)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

func randomSubsetBands(bands []*models.Band, count int) []*models.Band {
	if count > len(bands) {
		count = len(bands)
	}

	shuffled := make([]*models.Band, len(bands))
	copy(shuffled, bands)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
