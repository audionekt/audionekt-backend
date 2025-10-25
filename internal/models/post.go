package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	AuthorID   *uuid.UUID `json:"author_id" db:"author_id"`
	AuthorType string     `json:"author_type" db:"author_type"`
	BandID     *uuid.UUID `json:"band_id" db:"band_id"`
	UserID     *uuid.UUID `json:"user_id" db:"user_id"`
	Content    string     `json:"content" db:"content"`
	MediaURLs  []string   `json:"media_urls" db:"media_urls"`
	MediaTypes []string   `json:"media_types" db:"media_types"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`

	// Joined data
	Author       interface{} `json:"author,omitempty"` // User or Band
	LikesCount   int         `json:"likes_count,omitempty"`
	RepostsCount int         `json:"reposts_count,omitempty"`
	IsLiked      bool        `json:"is_liked,omitempty"`
	IsReposted   bool        `json:"is_reposted,omitempty"`
}

type CreatePostRequest struct {
	Content    string   `json:"content" validate:"required,min=1,max=2000"`
	MediaURLs  []string `json:"media_urls,omitempty"`
	MediaTypes []string `json:"media_types,omitempty"`
}

type UpdatePostRequest struct {
	Content    *string  `json:"content,omitempty" validate:"omitempty,min=1,max=2000"`
	MediaURLs  []string `json:"media_urls,omitempty"`
	MediaTypes []string `json:"media_types,omitempty"`
}

type PostResponse struct {
	ID           uuid.UUID   `json:"id"`
	AuthorID     *uuid.UUID  `json:"author_id"`
	AuthorType   string      `json:"author_type"`
	BandID       *uuid.UUID  `json:"band_id"`
	UserID       *uuid.UUID  `json:"user_id"`
	Content      string      `json:"content"`
	MediaURLs    []string    `json:"media_urls"`
	MediaTypes   []string    `json:"media_types"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Author       interface{} `json:"author,omitempty"`
	LikesCount   int         `json:"likes_count"`
	RepostsCount int         `json:"reposts_count"`
	IsLiked      bool        `json:"is_liked"`
	IsReposted   bool        `json:"is_reposted"`
}

func (p *Post) ToResponse() *PostResponse {
	return &PostResponse{
		ID:           p.ID,
		AuthorID:     p.AuthorID,
		AuthorType:   p.AuthorType,
		BandID:       p.BandID,
		UserID:       p.UserID,
		Content:      p.Content,
		MediaURLs:    p.MediaURLs,
		MediaTypes:   p.MediaTypes,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		Author:       p.Author,
		LikesCount:   p.LikesCount,
		RepostsCount: p.RepostsCount,
		IsLiked:      p.IsLiked,
		IsReposted:   p.IsReposted,
	}
}
