package models

import (
	"time"

	"github.com/google/uuid"
)

type Band struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Bio               *string   `json:"bio" db:"bio"`
	ProfilePictureURL *string   `json:"profile_picture_url" db:"profile_picture_url"`
	Location          *Location `json:"location" db:"location"`
	City              *string   `json:"city" db:"city"`
	Country           *string   `json:"country" db:"country"`
	Genres            []string  `json:"genres" db:"genres"`
	LookingFor        []string  `json:"looking_for" db:"looking_for"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type BandMember struct {
	ID       uuid.UUID `json:"id" db:"id"`
	BandID   uuid.UUID `json:"band_id" db:"band_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Role     *string   `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	User     *User     `json:"user,omitempty"`
	Band     *Band     `json:"band,omitempty"`
}

type CreateBandRequest struct {
	Name       string    `json:"name" validate:"required,min=1,max=100"`
	Bio        *string   `json:"bio,omitempty"`
	Location   *Location `json:"location,omitempty"`
	City       *string   `json:"city,omitempty"`
	Country    *string   `json:"country,omitempty"`
	Genres     []string  `json:"genres,omitempty"`
	LookingFor []string  `json:"looking_for,omitempty"`
}

type UpdateBandRequest struct {
	Name       *string   `json:"name,omitempty"`
	Bio        *string   `json:"bio,omitempty"`
	Location   *Location `json:"location,omitempty"`
	City       *string   `json:"city,omitempty"`
	Country    *string   `json:"country,omitempty"`
	Genres     []string  `json:"genres,omitempty"`
	LookingFor []string  `json:"looking_for,omitempty"`
}

type BandResponse struct {
	ID                uuid.UUID    `json:"id"`
	Name              string       `json:"name"`
	Bio               *string      `json:"bio"`
	ProfilePictureURL *string      `json:"profile_picture_url"`
	Location          *Location    `json:"location"`
	City              *string      `json:"city"`
	Country           *string      `json:"country"`
	Genres            []string     `json:"genres"`
	LookingFor        []string     `json:"looking_for"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	Members           []BandMember `json:"members,omitempty"`
	MemberCount       int          `json:"member_count,omitempty"`
}

func (b *Band) ToResponse() *BandResponse {
	return &BandResponse{
		ID:                b.ID,
		Name:              b.Name,
		Bio:               b.Bio,
		ProfilePictureURL: b.ProfilePictureURL,
		Location:          b.Location,
		City:              b.City,
		Country:           b.Country,
		Genres:            b.Genres,
		LookingFor:        b.LookingFor,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
}
