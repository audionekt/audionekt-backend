package repository

import (
	"context"

	"musicapp/internal/db"
	"musicapp/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FollowRepository struct {
	db *db.DB
}

func NewFollowRepository(db *db.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Create(ctx context.Context, follow *models.Follow) error {
	query := `
		INSERT INTO follows (id, follower_id, following_type, following_user_id, following_band_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		ON CONFLICT (follower_id, following_type, following_user_id, following_band_id) DO NOTHING
	`

	_, err := r.db.Pool.Exec(ctx, query,
		follow.FollowerID, follow.FollowingType, follow.FollowingUserID, follow.FollowingBandID,
	)
	return err
}

func (r *FollowRepository) Delete(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) error {
	query := `
		DELETE FROM follows 
		WHERE follower_id = $1 AND following_type = $2 
		AND (following_user_id = $3 OR following_band_id = $4)
	`

	_, err := r.db.Pool.Exec(ctx, query, followerID, followingType, followingUserID, followingBandID)
	return err
}

func (r *FollowRepository) IsFollowing(ctx context.Context, followerID uuid.UUID, followingType string, followingUserID, followingBandID *uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM follows 
			WHERE follower_id = $1 AND following_type = $2 
			AND (following_user_id = $3 OR following_band_id = $4)
		)
	`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, followerID, followingType, followingUserID, followingBandID).Scan(&exists)
	return exists, err
}

func (r *FollowRepository) GetFollowers(ctx context.Context, followingType string, followingUserID, followingBandID *uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	var query string
	var args []interface{}

	if followingType == "user" && followingUserID != nil {
		query = `
			SELECT f.id, f.follower_id, f.following_type, f.following_user_id, f.following_band_id, f.created_at,
				u.id, u.username, u.email, u.display_name, u.bio, u.profile_picture_url,
				ST_Y(u.location::geometry) as lat, ST_X(u.location::geometry) as lng,
				u.city, u.country, u.genres, u.skills,
				u.spotify_url, u.soundcloud_url, u.instagram_handle,
				u.created_at, u.updated_at
			FROM follows f
			JOIN users u ON f.follower_id = u.id
			WHERE f.following_type = $1 AND f.following_user_id = $2
			ORDER BY f.created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{followingType, followingUserID, limit, offset}
	} else if followingType == "band" && followingBandID != nil {
		query = `
			SELECT f.id, f.follower_id, f.following_type, f.following_user_id, f.following_band_id, f.created_at,
				u.id, u.username, u.email, u.display_name, u.bio, u.profile_picture_url,
				ST_Y(u.location::geometry) as lat, ST_X(u.location::geometry) as lng,
				u.city, u.country, u.genres, u.skills,
				u.spotify_url, u.soundcloud_url, u.instagram_handle,
				u.created_at, u.updated_at
			FROM follows f
			JOIN users u ON f.follower_id = u.id
			WHERE f.following_type = $1 AND f.following_band_id = $2
			ORDER BY f.created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{followingType, followingBandID, limit, offset}
	} else {
		return nil, pgx.ErrNoRows
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []*models.Follow
	for rows.Next() {
		follow, err := r.scanFollowWithUser(rows)
		if err != nil {
			return nil, err
		}
		follows = append(follows, follow)
	}

	return follows, rows.Err()
}

func (r *FollowRepository) GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*models.Follow, error) {
	query := `
		SELECT f.id, f.follower_id, f.following_type, f.following_user_id, f.following_band_id, f.created_at,
			u.id, u.username, u.email, u.display_name, u.bio, u.profile_picture_url,
			ST_Y(u.location::geometry) as lat, ST_X(u.location::geometry) as lng,
			u.city, u.country, u.genres, u.skills,
			u.spotify_url, u.soundcloud_url, u.instagram_handle,
			u.created_at, u.updated_at,
			b.id, b.name, b.bio, b.profile_picture_url,
			ST_Y(b.location::geometry) as band_lat, ST_X(b.location::geometry) as band_lng,
			b.city, b.country, b.genres, b.looking_for,
			b.created_at, b.updated_at
		FROM follows f
		LEFT JOIN users u ON f.following_user_id = u.id
		LEFT JOIN bands b ON f.following_band_id = b.id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, followerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []*models.Follow
	for rows.Next() {
		follow, err := r.scanFollowWithUserAndBand(rows)
		if err != nil {
			return nil, err
		}
		follows = append(follows, follow)
	}

	return follows, rows.Err()
}

func (r *FollowRepository) scanFollowWithUser(row pgx.Row) (*models.Follow, error) {
	var follow models.Follow
	var user models.User
	var lat, lng *float64

	err := row.Scan(
		&follow.ID, &follow.FollowerID, &follow.FollowingType, &follow.FollowingUserID, &follow.FollowingBandID, &follow.CreatedAt,
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio, &user.ProfilePictureURL,
		&lat, &lng, &user.City, &user.Country, &user.Genres, &user.Skills,
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

	follow.Follower = &user
	return &follow, nil
}

func (r *FollowRepository) scanFollowWithUserAndBand(row pgx.Row) (*models.Follow, error) {
	var follow models.Follow
	var user models.User
	var band models.Band
	var lat, lng, bandLat, bandLng *float64

	err := row.Scan(
		&follow.ID, &follow.FollowerID, &follow.FollowingType, &follow.FollowingUserID, &follow.FollowingBandID, &follow.CreatedAt,
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio, &user.ProfilePictureURL,
		&lat, &lng, &user.City, &user.Country, &user.Genres, &user.Skills,
		&user.SpotifyURL, &user.SoundcloudURL, &user.InstagramHandle,
		&user.CreatedAt, &user.UpdatedAt,
		&band.ID, &band.Name, &band.Bio, &band.ProfilePictureURL,
		&bandLat, &bandLng, &band.City, &band.Country, &band.Genres, &band.LookingFor,
		&band.CreatedAt, &band.UpdatedAt,
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

	if bandLat != nil && bandLng != nil {
		band.Location = &models.Location{
			Latitude:  *bandLat,
			Longitude: *bandLng,
		}
	}

	if follow.FollowingType == "user" {
		follow.FollowingUser = &user
	} else if follow.FollowingType == "band" {
		follow.FollowingBand = &band
	}

	return &follow, nil
}
