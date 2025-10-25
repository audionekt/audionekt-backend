package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"musicapp/internal/middleware"
	"musicapp/internal/models"
	"musicapp/internal/service"
	"musicapp/pkg/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService *service.UserService
	bandService *service.BandService
}

func NewUserHandler(userService *service.UserService, bandService *service.BandService) *UserHandler {
	return &UserHandler{
		userService: userService,
		bandService: bandService,
	}
}

// @Summary Get user profile
// @Description Get user profile by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserResponse "User retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.WriteSuccess(w, "User retrieved successfully", user.ToResponse())
}

// @Summary Update user profile
// @Description Update user profile information
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.UpdateUserRequest true "User update data"
// @Security BearerAuth
// @Success 200 {object} models.UserResponse "User updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - can only update own profile"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check if user is updating their own profile
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if currentUserID != userIDStr {
		utils.WriteError(w, http.StatusForbidden, "You can only update your own profile")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	utils.WriteSuccess(w, "User updated successfully", user.ToResponse())
}

func (h *UserHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	posts, err := h.userService.GetUserPosts(r.Context(), userID, limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	// Convert to response format
	var postResponses []*models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToResponse())
	}

	utils.WriteSuccess(w, "User posts retrieved successfully", postResponses)
}

func (h *UserHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	followers, err := h.userService.GetFollowers(r.Context(), userID, limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve followers")
		return
	}

	// Convert to response format
	var userResponses []*models.UserResponse
	for _, user := range followers {
		userResponses = append(userResponses, user.ToResponse())
	}

	utils.WriteSuccess(w, "Followers retrieved successfully", userResponses)
}

func (h *UserHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	following, err := h.userService.GetFollowing(r.Context(), userID, limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve following")
		return
	}

	// Convert to response format
	var userResponses []*models.UserResponse
	for _, user := range following {
		userResponses = append(userResponses, user.ToResponse())
	}

	utils.WriteSuccess(w, "Following retrieved successfully", userResponses)
}

// @Summary Get nearby users
// @Description Get users within a specified radius of given coordinates
// @Tags Users
// @Accept json
// @Produce json
// @Param lat query number true "Latitude" example(37.7749)
// @Param lng query number true "Longitude" example(-122.4194)
// @Param radius query int false "Radius in kilometers" example(50)
// @Param limit query int false "Maximum number of users to return" example(20)
// @Success 200 {array} models.UserResponse "Nearby users retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid coordinates"
// @Router /users/nearby [get]
func (h *UserHandler) GetNearbyUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	radiusStr := r.URL.Query().Get("radius")
	limitStr := r.URL.Query().Get("limit")

	if latStr == "" || lngStr == "" {
		utils.WriteError(w, http.StatusBadRequest, "Latitude and longitude are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid latitude")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid longitude")
		return
	}

	radius := 50 // Default 50km
	if radiusStr != "" {
		if r, err := strconv.Atoi(radiusStr); err == nil && r > 0 && r <= 500 {
			radius = r
		}
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	users, err := h.userService.GetNearbyUsers(r.Context(), lat, lng, radius, limit)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Convert to response format
	var userResponses []*models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	utils.WriteSuccess(w, "Nearby users retrieved successfully", userResponses)
}

// @Summary Upload profile picture
// @Description Upload a profile picture for the user
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "User ID"
// @Param profile_picture formData file true "Profile picture image file"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Profile picture uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid file or user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - can only update own profile"
// @Router /users/{id}/profile-picture [post]
func (h *UserHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check if user is updating their own profile
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if currentUserID != userIDStr {
		utils.WriteError(w, http.StatusForbidden, "You can only update your own profile")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, handler, err := r.FormFile("profile_picture")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Failed to read file")
		return
	}

	// Use service to upload profile picture
	profilePictureURL, err := h.userService.UploadProfilePicture(r.Context(), userID, handler.Filename, fileData)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"profile_picture_url": profilePictureURL,
	}

	utils.WriteSuccess(w, "Profile picture uploaded successfully", response)
}

// @Summary Get user's bands
// @Description Get all bands that a user is a member of
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {array} models.BandMember "User bands retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Router /users/{id}/bands [get]
func (h *UserHandler) GetUserBands(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use band service to get user bands
	bands, err := h.bandService.GetUserBands(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve user bands")
		return
	}

	utils.WriteSuccess(w, "User bands retrieved successfully", bands)
}

// @Summary Get all users
// @Description Get all users with pagination
// @Tags Users
// @Accept json
// @Produce json
// @Param limit query int false "Number of users to return (default: 20, max: 100)" default(20)
// @Param offset query int false "Number of users to skip (default: 0)" default(0)
// @Success 200 {object} map[string]interface{} "Users retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid pagination parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := h.userService.GetAllUsers(r.Context(), limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	response := map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	}

	utils.WriteSuccess(w, "Users retrieved successfully", response)
}
