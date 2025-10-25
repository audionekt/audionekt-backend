package repository

import (
	"context"

	"musicapp/internal/db"
	"musicapp/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BandRepository struct {
	db *db.DB
}

func NewBandRepository(db *db.DB) *BandRepository {
	return &BandRepository{db: db}
}

func (r *BandRepository) Create(ctx context.Context, band *models.Band) error {
	query := `
		INSERT INTO bands (id, name, bio, profile_picture_url, location, city, country, genres, looking_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography, $7, $8, $9, $10, NOW(), NOW())
	`

	var lat, lng *float64
	if band.Location != nil {
		lat = &band.Location.Latitude
		lng = &band.Location.Longitude
	}

	_, err := r.db.Pool.Exec(ctx, query,
		band.ID, band.Name, band.Bio, band.ProfilePictureURL,
		lat, lng, band.City, band.Country,
		band.Genres, band.LookingFor,
	)
	return err
}

func (r *BandRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Band, error) {
	query := `
		SELECT id, name, bio, profile_picture_url, 
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, looking_for, created_at, updated_at
		FROM bands 
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)
	return r.scanBand(row)
}

func (r *BandRepository) Update(ctx context.Context, band *models.Band) error {
	query := `
		UPDATE bands SET 
			name = $2, bio = $3, profile_picture_url = $4,
			location = ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography,
			city = $7, country = $8, genres = $9, looking_for = $10,
			updated_at = NOW()
		WHERE id = $1
	`

	var lat, lng *float64
	if band.Location != nil {
		lat = &band.Location.Latitude
		lng = &band.Location.Longitude
	}

	_, err := r.db.Pool.Exec(ctx, query,
		band.ID, band.Name, band.Bio, band.ProfilePictureURL,
		lat, lng, band.City, band.Country,
		band.Genres, band.LookingFor,
	)
	return err
}

func (r *BandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM bands WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *BandRepository) GetNearby(ctx context.Context, lat, lng float64, radiusKm int, limit int) ([]*models.Band, error) {
	query := `
		SELECT id, name, bio, profile_picture_url, 
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, looking_for, created_at, updated_at,
			ST_Distance(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_meters
		FROM bands 
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

	var bands []*models.Band
	for rows.Next() {
		band, err := r.scanBandWithDistance(rows)
		if err != nil {
			return nil, err
		}
		bands = append(bands, band)
	}

	return bands, rows.Err()
}

func (r *BandRepository) AddMember(ctx context.Context, bandID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO band_members (id, band_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (band_id, user_id) DO NOTHING
	`

	_, err := r.db.Pool.Exec(ctx, query, bandID, userID, role)
	return err
}

func (r *BandRepository) RemoveMember(ctx context.Context, bandID, userID uuid.UUID) error {
	query := `DELETE FROM band_members WHERE band_id = $1 AND user_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, bandID, userID)
	return err
}

func (r *BandRepository) GetMembers(ctx context.Context, bandID uuid.UUID) ([]*models.BandMember, error) {
	query := `
		SELECT bm.id, bm.band_id, bm.user_id, bm.role, bm.joined_at,
			u.id, u.username, u.email, u.display_name, u.bio, u.profile_picture_url,
			ST_Y(u.location::geometry) as lat, ST_X(u.location::geometry) as lng,
			u.city, u.country, u.genres, u.skills,
			u.spotify_url, u.soundcloud_url, u.instagram_handle,
			u.created_at, u.updated_at
		FROM band_members bm
		JOIN users u ON bm.user_id = u.id
		WHERE bm.band_id = $1
		ORDER BY bm.joined_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, bandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.BandMember
	for rows.Next() {
		member, err := r.scanBandMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

func (r *BandRepository) IsMember(ctx context.Context, bandID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM band_members WHERE band_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, bandID, userID).Scan(&exists)
	return exists, err
}

func (r *BandRepository) IsAdmin(ctx context.Context, bandID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM band_members WHERE band_id = $1 AND user_id = $2 AND role = 'Admin')`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, bandID, userID).Scan(&exists)
	return exists, err
}

func (r *BandRepository) GetUserBands(ctx context.Context, userID uuid.UUID) ([]*models.BandMember, error) {
	query := `
		SELECT bm.id, bm.band_id, bm.user_id, bm.role, bm.joined_at,
			b.id, b.name, b.bio, b.profile_picture_url,
			ST_Y(b.location::geometry) as lat, ST_X(b.location::geometry) as lng,
			b.city, b.country, b.genres, b.looking_for,
			b.created_at, b.updated_at
		FROM band_members bm
		JOIN bands b ON bm.band_id = b.id
		WHERE bm.user_id = $1
		ORDER BY bm.joined_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*models.BandMember
	for rows.Next() {
		membership, err := r.scanUserBandMembership(rows)
		if err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}

	return memberships, rows.Err()
}

// GetAll gets all bands with pagination
func (r *BandRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Band, error) {
	query := `
		SELECT id, name, bio, profile_picture_url,
			ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng,
			city, country, genres, looking_for, created_at, updated_at
		FROM bands
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bands []*models.Band
	for rows.Next() {
		band, err := r.scanBand(rows)
		if err != nil {
			return nil, err
		}
		bands = append(bands, band)
	}

	return bands, rows.Err()
}

func (r *BandRepository) scanBand(row pgx.Row) (*models.Band, error) {
	var band models.Band
	var lat, lng *float64

	err := row.Scan(
		&band.ID, &band.Name, &band.Bio, &band.ProfilePictureURL,
		&lat, &lng, &band.City, &band.Country,
		&band.Genres, &band.LookingFor,
		&band.CreatedAt, &band.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lat != nil && lng != nil {
		band.Location = &models.Location{
			Latitude:  *lat,
			Longitude: *lng,
		}
	}

	return &band, nil
}

func (r *BandRepository) scanBandWithDistance(row pgx.Row) (*models.Band, error) {
	var band models.Band
	var lat, lng *float64
	var distance float64

	err := row.Scan(
		&band.ID, &band.Name, &band.Bio, &band.ProfilePictureURL,
		&lat, &lng, &band.City, &band.Country,
		&band.Genres, &band.LookingFor,
		&band.CreatedAt, &band.UpdatedAt, &distance,
	)

	if err != nil {
		return nil, err
	}

	if lat != nil && lng != nil {
		band.Location = &models.Location{
			Latitude:  *lat,
			Longitude: *lng,
		}
	}

	return &band, nil
}

func (r *BandRepository) scanBandMember(row pgx.Row) (*models.BandMember, error) {
	var member models.BandMember
	var user models.User
	var lat, lng *float64

	err := row.Scan(
		&member.ID, &member.BandID, &member.UserID, &member.Role, &member.JoinedAt,
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

	member.User = &user
	return &member, nil
}

func (r *BandRepository) scanUserBandMembership(row pgx.Row) (*models.BandMember, error) {
	var membership models.BandMember
	var band models.Band
	var lat, lng *float64

	err := row.Scan(
		&membership.ID, &membership.BandID, &membership.UserID, &membership.Role, &membership.JoinedAt,
		&band.ID, &band.Name, &band.Bio, &band.ProfilePictureURL,
		&lat, &lng, &band.City, &band.Country, &band.Genres, &band.LookingFor,
		&band.CreatedAt, &band.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lat != nil && lng != nil {
		band.Location = &models.Location{
			Latitude:  *lat,
			Longitude: *lng,
		}
	}

	// Store band info in the Band field
	membership.Band = &band

	return &membership, nil
}
