package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Username          string    `json:"username" db:"username"`
	Email             string    `json:"email" db:"email"`
	PasswordHash      string    `json:"-" db:"password_hash"`
	DisplayName       *string   `json:"display_name" db:"display_name"`
	Bio               *string   `json:"bio" db:"bio"`
	ProfilePictureURL *string   `json:"profile_picture_url" db:"profile_picture_url"`
	Location          *Location `json:"location" db:"location"`
	City              *string   `json:"city" db:"city"`
	Country           *string   `json:"country" db:"country"`
	Genres            []string  `json:"genres" db:"genres"`
	Skills            []string  `json:"skills" db:"skills"`
	SpotifyURL        *string   `json:"spotify_url" db:"spotify_url"`
	SoundcloudURL     *string   `json:"soundcloud_url" db:"soundcloud_url"`
	InstagramHandle   *string   `json:"instagram_handle" db:"instagram_handle"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CreateUserRequest struct {
	Username string    `json:"username" validate:"required,min=3,max=50"`
	Email    string    `json:"email" validate:"required,email"`
	Password string    `json:"password" validate:"required,min=8"`
	Location *Location `json:"location,omitempty"`
	City     *string   `json:"city,omitempty"`
	Country  *string   `json:"country,omitempty"`
}

type UpdateUserRequest struct {
	DisplayName     *string   `json:"display_name,omitempty"`
	Bio             *string   `json:"bio,omitempty"`
	Location        *Location `json:"location,omitempty"`
	City            *string   `json:"city,omitempty"`
	Country         *string   `json:"country,omitempty"`
	Genres          []string  `json:"genres,omitempty"`
	Skills          []string  `json:"skills,omitempty"`
	SpotifyURL      *string   `json:"spotify_url,omitempty"`
	SoundcloudURL   *string   `json:"soundcloud_url,omitempty"`
	InstagramHandle *string   `json:"instagram_handle,omitempty"`
}

type UserResponse struct {
	ID                uuid.UUID `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	DisplayName       *string   `json:"display_name"`
	Bio               *string   `json:"bio"`
	ProfilePictureURL *string   `json:"profile_picture_url"`
	Location          *Location `json:"location"`
	City              *string   `json:"city"`
	Country           *string   `json:"country"`
	Genres            []string  `json:"genres"`
	Skills            []string  `json:"skills"`
	SpotifyURL        *string   `json:"spotify_url"`
	SoundcloudURL     *string   `json:"soundcloud_url"`
	InstagramHandle   *string   `json:"instagram_handle"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:                u.ID,
		Username:          u.Username,
		Email:             u.Email,
		DisplayName:       u.DisplayName,
		Bio:               u.Bio,
		ProfilePictureURL: u.ProfilePictureURL,
		Location:          u.Location,
		City:              u.City,
		Country:           u.Country,
		Genres:            u.Genres,
		Skills:            u.Skills,
		SpotifyURL:        u.SpotifyURL,
		SoundcloudURL:     u.SoundcloudURL,
		InstagramHandle:   u.InstagramHandle,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}
