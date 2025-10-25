package repository

import (
	"context"

	"musicapp/internal/db"
	"musicapp/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, display_name, bio, 
			profile_picture_url, location, city, country, genres, skills, 
			spotify_url, soundcloud_url, instagram_handle, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 
			ST_SetSRID(ST_MakePoint($8, $9), 4326)::geography, $10, $11, $12, $13, 
			$14, $15, $16, NOW(), NOW())
	`

	var lat, lng *float64
	if user.Location != nil {
		lat = &user.Location.Latitude
		lng = &user.Location.Longitude
	}

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.DisplayName, user.Bio, user.ProfilePictureURL,
		lat, lng, user.City, user.Country,
		user.Genres, user.Skills,
		user.SpotifyURL, user.SoundcloudURL, user.InstagramHandle,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, 
			profile_picture_url, 
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, skills, 
			spotify_url, soundcloud_url, instagram_handle, 
			created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)
	return r.scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, 
			profile_picture_url, 
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, skills, 
			spotify_url, soundcloud_url, instagram_handle, 
			created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, email)
	return r.scanUser(row)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, 
			profile_picture_url, 
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, skills, 
			spotify_url, soundcloud_url, instagram_handle, 
			created_at, updated_at
		FROM users 
		WHERE username = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, username)
	return r.scanUser(row)
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET 
			display_name = $2, bio = $3, profile_picture_url = $4,
			location = ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography,
			city = $7, country = $8, genres = $9, skills = $10,
			spotify_url = $11, soundcloud_url = $12, instagram_handle = $13,
			updated_at = NOW()
		WHERE id = $1
	`

	var lat, lng *float64
	if user.Location != nil {
		lat = &user.Location.Latitude
		lng = &user.Location.Longitude
	}

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID, user.DisplayName, user.Bio, user.ProfilePictureURL,
		lat, lng, user.City, user.Country,
		user.Genres, user.Skills,
		user.SpotifyURL, user.SoundcloudURL, user.InstagramHandle,
	)
	return err
}

func (r *UserRepository) GetNearby(ctx context.Context, lat, lng float64, radiusKm int, limit int) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, 
			profile_picture_url, 
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, skills, 
			spotify_url, soundcloud_url, instagram_handle, 
			created_at, updated_at,
			ST_Distance(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_meters
		FROM users 
		WHERE ST_DWithin(
			location,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3
		)
		ORDER BY distance_meters
		LIMIT $4
	`

	radiusMeters := radiusKm * 1000
	rows, err := r.db.Pool.Query(ctx, query, lng, lat, radiusMeters, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := r.scanUserWithDistance(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepository) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.display_name, u.bio, 
			u.profile_picture_url, 
			ST_Y(u.location::geometry) as lat, ST_X(u.location::geometry) as lng,
			u.city, u.country, u.genres, u.skills, 
			u.spotify_url, u.soundcloud_url, u.instagram_handle, 
			u.created_at, u.updated_at
		FROM users u
		JOIN follows f ON u.id = f.follower_id
		WHERE f.following_type = 'user' AND f.following_user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepository) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.display_name, u.bio, 
			u.profile_picture_url, 
			ST_Y(u.location::geometry) as lat, ST_X(u.location::geometry) as lng,
			u.city, u.country, u.genres, u.skills, 
			u.spotify_url, u.soundcloud_url, u.instagram_handle, 
			u.created_at, u.updated_at
		FROM users u
		JOIN follows f ON u.id = f.following_user_id
		WHERE f.follower_id = $1 AND f.following_type = 'user'
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetAll gets all users with pagination
func (r *UserRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, profile_picture_url,
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, skills, spotify_url, soundcloud_url, instagram_handle,
			created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepository) scanUser(row pgx.Row) (*models.User, error) {
	var user models.User
	var lat, lng *float64

	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.Bio, &user.ProfilePictureURL,
		&lat, &lng, &user.City, &user.Country,
		&user.Genres, &user.Skills,
		&user.SpotifyURL, &user.SoundcloudURL, &user.InstagramHandle,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lat != nil && lng != nil {
		user.Location = &models.Location{
			Latitude:  *lat,
			Longitude: *lng,
		}
	}

	return &user, nil
}

func (r *UserRepository) scanUserWithDistance(row pgx.Row) (*models.User, error) {
	var user models.User
	var lat, lng *float64
	var distance float64

	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.Bio, &user.ProfilePictureURL,
		&lat, &lng, &user.City, &user.Country,
		&user.Genres, &user.Skills,
		&user.SpotifyURL, &user.SoundcloudURL, &user.InstagramHandle,
		&user.CreatedAt, &user.UpdatedAt, &distance,
	)

	if err != nil {
		return nil, err
	}

	if lat != nil && lng != nil {
		user.Location = &models.Location{
			Latitude:  *lat,
			Longitude: *lng,
		}
	}

	return &user, nil
}
