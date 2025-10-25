package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"musicapp/internal/middleware"
	"musicapp/internal/models"
	"musicapp/internal/service"
	"musicapp/pkg/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// @Summary Create a new post
// @Description Create a new post (user or band post)
// @Tags Posts
// @Accept json
// @Produce json
// @Param post body models.CreatePostRequest true "Post creation data"
// @Security BearerAuth
// @Success 201 {object} models.PostResponse "Post created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /posts [post]
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use service to create post
	post, err := h.postService.CreatePost(r.Context(), userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, "Post created successfully", post.ToResponse())
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user ID if available
	var currentUserID *uuid.UUID
	if userIDStr, ok := middleware.GetUserIDFromContext(r.Context()); ok {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			currentUserID = &userID
		}
	}

	// Use service to get post
	post, err := h.postService.GetPost(r.Context(), postID, currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Post not found")
		return
	}

	utils.WriteSuccess(w, "Post retrieved successfully", post.ToResponse())
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Use service to update post
	post, err := h.postService.UpdatePost(r.Context(), postID, userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Post updated successfully", post.ToResponse())
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use service to delete post
	if err := h.postService.DeletePost(r.Context(), postID, userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Post deleted successfully", nil)
}

func (h *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use service to like post
	if err := h.postService.LikePost(r.Context(), userID, postID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Post liked successfully", nil)
}

func (h *PostHandler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use service to unlike post
	if err := h.postService.UnlikePost(r.Context(), userID, postID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Post unliked successfully", nil)
}

func (h *PostHandler) Repost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Use service to repost
	if err := h.postService.Repost(r.Context(), userID, postID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Post reposted successfully", nil)
}

// @Summary Get user feed
// @Description Get personalized feed for the authenticated user
// @Tags Posts
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of posts to return" example(20)
// @Param offset query int false "Number of posts to skip" example(0)
// @Security BearerAuth
// @Success 200 {array} models.PostResponse "Feed retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /feed [get]
func (h *PostHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
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

	// Use service to get feed
	posts, err := h.postService.GetFeed(r.Context(), userID, limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve feed")
		return
	}

	// Convert to response format
	var postResponses []*models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToResponse())
	}

	utils.WriteSuccess(w, "Feed retrieved successfully", postResponses)
}

func (h *PostHandler) GetExploreFeed(w http.ResponseWriter, r *http.Request) {
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

	// Use service to get explore feed
	posts, err := h.postService.GetExploreFeed(r.Context(), limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve explore feed")
		return
	}

	// Convert to response format
	var postResponses []*models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToResponse())
	}

	utils.WriteSuccess(w, "Explore feed retrieved successfully", postResponses)
}

// @Summary Upload media to post
// @Description Upload image or audio files to an existing post
// @Tags Posts
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Post ID"
// @Param media formData file true "Media file (image or audio)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Media uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid file or post ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /posts/{id}/media [post]
func (h *PostHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(100 << 20) // 100MB limit for media
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	files := r.MultipartForm.File["media"]
	if len(files) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "No media files uploaded")
		return
	}

	var mediaURLs []string
	var mediaTypes []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Failed to open file")
			return
		}
		defer file.Close()

		// Read file data
		fileData, err := io.ReadAll(file)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Failed to read file")
			return
		}

		// Determine media type
		contentType := fileHeader.Header.Get("Content-Type")
		var mediaType string

		if strings.HasPrefix(contentType, "image/") {
			mediaType = "image"
		} else if strings.HasPrefix(contentType, "audio/") {
			mediaType = "audio"
		} else if strings.HasPrefix(contentType, "video/") {
			mediaType = "video"
			utils.WriteError(w, http.StatusBadRequest, "Video upload not yet supported")
			return
		} else {
			utils.WriteError(w, http.StatusBadRequest, "Invalid file type. Only images and audio are allowed")
			return
		}

		// Use service to upload media
		mediaURL, uploadedMediaType, err := h.postService.UploadMedia(r.Context(), postID, userID, fileHeader.Filename, fileData, mediaType)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		mediaURLs = append(mediaURLs, mediaURL)
		mediaTypes = append(mediaTypes, uploadedMediaType)
	}

	response := map[string]interface{}{
		"media_urls":  mediaURLs,
		"media_types": mediaTypes,
	}

	utils.WriteSuccess(w, "Media uploaded successfully", response)
}

// @Summary Get all posts
// @Description Get all posts with pagination
// @Tags Posts
// @Accept json
// @Produce json
// @Param limit query int false "Number of posts to return (default: 20, max: 100)" default(20)
// @Param offset query int false "Number of posts to skip (default: 0)" default(0)
// @Success 200 {object} map[string]interface{} "Posts retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid pagination parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /posts [get]
func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
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

	posts, err := h.postService.GetAllPosts(r.Context(), limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	// Convert to response format
	var postResponses []*models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToResponse())
	}

	response := map[string]interface{}{
		"posts":  postResponses,
		"limit":  limit,
		"offset": offset,
		"count":  len(postResponses),
	}

	utils.WriteSuccess(w, "Posts retrieved successfully", response)
}
