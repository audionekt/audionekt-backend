package models

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	FollowerID      uuid.UUID  `json:"follower_id" db:"follower_id"`
	FollowingType   string     `json:"following_type" db:"following_type"`
	FollowingUserID *uuid.UUID `json:"following_user_id" db:"following_user_id"`
	FollowingBandID *uuid.UUID `json:"following_band_id" db:"following_band_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`

	// Joined data
	Follower      *User `json:"follower,omitempty"`
	FollowingUser *User `json:"following_user,omitempty"`
	FollowingBand *Band `json:"following_band,omitempty"`
}

type FollowRequest struct {
	FollowingType   string     `json:"following_type" validate:"required,oneof=user band"`
	FollowingUserID *uuid.UUID `json:"following_user_id,omitempty"`
	FollowingBandID *uuid.UUID `json:"following_band_id,omitempty"`
}

type FollowResponse struct {
	ID              uuid.UUID  `json:"id"`
	FollowerID      uuid.UUID  `json:"follower_id"`
	FollowingType   string     `json:"following_type"`
	FollowingUserID *uuid.UUID `json:"following_user_id"`
	FollowingBandID *uuid.UUID `json:"following_band_id"`
	CreatedAt       time.Time  `json:"created_at"`
	Follower        *User      `json:"follower,omitempty"`
	FollowingUser   *User      `json:"following_user,omitempty"`
	FollowingBand   *Band      `json:"following_band,omitempty"`
}

func (f *Follow) ToResponse() *FollowResponse {
	return &FollowResponse{
		ID:              f.ID,
		FollowerID:      f.FollowerID,
		FollowingType:   f.FollowingType,
		FollowingUserID: f.FollowingUserID,
		FollowingBandID: f.FollowingBandID,
		CreatedAt:       f.CreatedAt,
		Follower:        f.Follower,
		FollowingUser:   f.FollowingUser,
		FollowingBand:   f.FollowingBand,
	}
}
