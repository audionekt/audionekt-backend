package handlers

import (
	"encoding/json"
	"net/http"

	"musicapp/internal/middleware"
	"musicapp/internal/models"
	"musicapp/internal/service"
	"musicapp/pkg/utils"

	"github.com/google/uuid"
)

type FollowHandler struct {
	followService *service.FollowService
}

func NewFollowHandler(followService *service.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

func (h *FollowHandler) Follow(w http.ResponseWriter, r *http.Request) {
	var req models.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.FollowingType == "" {
		utils.WriteError(w, http.StatusBadRequest, "Following type is required")
		return
	}

	if req.FollowingType == "user" && req.FollowingUserID == nil {
		utils.WriteError(w, http.StatusBadRequest, "Following user ID is required for user type")
		return
	}

	if req.FollowingType == "band" && req.FollowingBandID == nil {
		utils.WriteError(w, http.StatusBadRequest, "Following band ID is required for band type")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	followerID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use service to follow
	if req.FollowingType == "user" {
		err = h.followService.FollowUser(r.Context(), followerID, *req.FollowingUserID)
	} else if req.FollowingType == "band" {
		err = h.followService.FollowBand(r.Context(), followerID, *req.FollowingBandID)
	}

	if err != nil {
		if err.Error() == "you cannot follow yourself" {
			utils.WriteError(w, http.StatusBadRequest, "You cannot follow yourself")
			return
		}
		if err.Error() == "already following this user" || err.Error() == "already following this band" {
			utils.WriteError(w, http.StatusConflict, "Already following")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to follow")
		return
	}

	utils.WriteSuccess(w, "Successfully followed", nil)
}

func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	var req models.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.FollowingType == "" {
		utils.WriteError(w, http.StatusBadRequest, "Following type is required")
		return
	}

	if req.FollowingType == "user" && req.FollowingUserID == nil {
		utils.WriteError(w, http.StatusBadRequest, "Following user ID is required for user type")
		return
	}

	if req.FollowingType == "band" && req.FollowingBandID == nil {
		utils.WriteError(w, http.StatusBadRequest, "Following band ID is required for band type")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	followerID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use service to unfollow
	if req.FollowingType == "user" {
		err = h.followService.UnfollowUser(r.Context(), followerID, *req.FollowingUserID)
	} else if req.FollowingType == "band" {
		err = h.followService.UnfollowBand(r.Context(), followerID, *req.FollowingBandID)
	}

	if err != nil {
		if err.Error() == "not following this user" || err.Error() == "not following this band" {
			utils.WriteError(w, http.StatusNotFound, "Not following")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to unfollow")
		return
	}

	utils.WriteSuccess(w, "Successfully unfollowed", nil)
}
