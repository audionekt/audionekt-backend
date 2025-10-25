package repository

import (
	"context"

	"musicapp/internal/db"
	"musicapp/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PostRepository struct {
	db *db.DB
}

func NewPostRepository(db *db.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	query := `
		INSERT INTO posts (id, author_id, author_type, band_id, user_id, content, media_urls, media_types, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`

	_, err := r.db.Pool.Exec(ctx, query,
		post.ID, post.AuthorID, post.AuthorType, post.BandID, post.UserID,
		post.Content, post.MediaURLs, post.MediaTypes,
	)
	return err
}

func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	query := `
		SELECT p.id, p.author_id, p.author_type, p.band_id, p.user_id, p.content, 
			p.media_urls, p.media_types, p.created_at, p.updated_at,
			COALESCE(l.likes_count, 0) as likes_count,
			COALESCE(r.reposts_count, 0) as reposts_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as likes_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as reposts_count
			FROM reposts
			GROUP BY post_id
		) r ON p.id = r.post_id
		WHERE p.id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)
	return r.scanPost(row)
}

func (r *PostRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT p.id, p.author_id, p.author_type, p.band_id, p.user_id, p.content, 
			p.media_urls, p.media_types, p.created_at, p.updated_at,
			COALESCE(l.likes_count, 0) as likes_count,
			COALESCE(r.reposts_count, 0) as reposts_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as likes_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as reposts_count
			FROM reposts
			GROUP BY post_id
		) r ON p.id = r.post_id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := r.scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (r *PostRepository) GetByBandID(ctx context.Context, bandID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT p.id, p.author_id, p.author_type, p.band_id, p.user_id, p.content, 
			p.media_urls, p.media_types, p.created_at, p.updated_at,
			COALESCE(l.likes_count, 0) as likes_count,
			COALESCE(r.reposts_count, 0) as reposts_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as likes_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as reposts_count
			FROM reposts
			GROUP BY post_id
		) r ON p.id = r.post_id
		WHERE p.band_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, bandID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := r.scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (r *PostRepository) GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, error) {
	// If userID is empty (for explore feed), return recent posts
	if userID == uuid.Nil {
		query := `
			SELECT p.id, p.author_id, p.author_type, p.band_id, p.user_id, p.content, 
				p.media_urls, p.media_types, p.created_at, p.updated_at,
				COALESCE(l.likes_count, 0) as likes_count,
				COALESCE(r.reposts_count, 0) as reposts_count
			FROM posts p
			LEFT JOIN (
				SELECT post_id, COUNT(*) as likes_count
				FROM likes
				GROUP BY post_id
			) l ON p.id = l.post_id
			LEFT JOIN (
				SELECT post_id, COUNT(*) as reposts_count
				FROM reposts
				GROUP BY post_id
			) r ON p.id = r.post_id
			ORDER BY p.created_at DESC
			LIMIT $1 OFFSET $2
		`

		rows, err := r.db.Pool.Query(ctx, query, limit, offset)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var posts []*models.Post
		for rows.Next() {
			post, err := r.scanPost(rows)
			if err != nil {
				return nil, err
			}
			posts = append(posts, post)
		}

		return posts, rows.Err()
	}

	// Personalized feed for specific user
	query := `
		SELECT p.id, p.author_id, p.author_type, p.band_id, p.user_id, p.content, 
			p.media_urls, p.media_types, p.created_at, p.updated_at,
			COALESCE(l.likes_count, 0) as likes_count,
			COALESCE(r.reposts_count, 0) as reposts_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as likes_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as reposts_count
			FROM reposts
			GROUP BY post_id
		) r ON p.id = r.post_id
		WHERE p.author_id IN (
			SELECT following_user_id FROM follows WHERE follower_id = $1 AND following_type = 'user'
			UNION
			SELECT following_band_id FROM follows WHERE follower_id = $1 AND following_type = 'band'
		)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := r.scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (r *PostRepository) Update(ctx context.Context, post *models.Post) error {
	query := `
		UPDATE posts SET 
			content = $2, media_urls = $3, media_types = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		post.ID, post.Content, post.MediaURLs, post.MediaTypes,
	)
	return err
}

func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *PostRepository) LikePost(ctx context.Context, userID, postID uuid.UUID) error {
	query := `
		INSERT INTO likes (id, user_id, post_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		ON CONFLICT (user_id, post_id) DO NOTHING
	`

	_, err := r.db.Pool.Exec(ctx, query, userID, postID)
	return err
}

func (r *PostRepository) UnlikePost(ctx context.Context, userID, postID uuid.UUID) error {
	query := `DELETE FROM likes WHERE user_id = $1 AND post_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, postID)
	return err
}

func (r *PostRepository) Repost(ctx context.Context, userID, postID uuid.UUID) error {
	query := `
		INSERT INTO reposts (id, user_id, post_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		ON CONFLICT (user_id, post_id) DO NOTHING
	`

	_, err := r.db.Pool.Exec(ctx, query, userID, postID)
	return err
}

func (r *PostRepository) IsLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND post_id = $2)`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, userID, postID).Scan(&exists)
	return exists, err
}

func (r *PostRepository) IsReposted(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM reposts WHERE user_id = $1 AND post_id = $2)`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, userID, postID).Scan(&exists)
	return exists, err
}

func (r *PostRepository) scanPost(row pgx.Row) (*models.Post, error) {
	var post models.Post

	err := row.Scan(
		&post.ID, &post.AuthorID, &post.AuthorType, &post.BandID, &post.UserID,
		&post.Content, &post.MediaURLs, &post.MediaTypes,
		&post.CreatedAt, &post.UpdatedAt,
		&post.LikesCount, &post.RepostsCount,
	)

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetAll gets all posts with pagination
func (r *PostRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT p.id, p.author_id, p.author_type, p.band_id, p.user_id, p.content, 
			p.media_urls, p.media_types, p.created_at, p.updated_at,
			COALESCE(l.likes_count, 0) as likes_count,
			COALESCE(r.reposts_count, 0) as reposts_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as likes_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as reposts_count
			FROM reposts
			GROUP BY post_id
		) r ON p.id = r.post_id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := r.scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}
