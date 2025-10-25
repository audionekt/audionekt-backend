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

type BandHandler struct {
	bandService *service.BandService
}

func NewBandHandler(bandService *service.BandService) *BandHandler {
	return &BandHandler{
		bandService: bandService,
	}
}

// @Summary Create a new band
// @Description Create a new band with the current user as admin
// @Tags Bands
// @Accept json
// @Produce json
// @Param band body models.CreateBandRequest true "Band creation data"
// @Security BearerAuth
// @Success 201 {object} models.BandResponse "Band created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /bands [post]
func (h *BandHandler) CreateBand(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBandRequest
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

	// Use service to create band
	band, err := h.bandService.CreateBand(r.Context(), userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, "Band created successfully", band.ToResponse())
}

// @Summary Get band details
// @Description Get band information by ID
// @Tags Bands
// @Accept json
// @Produce json
// @Param id path string true "Band ID"
// @Success 200 {object} models.BandResponse "Band retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid band ID"
// @Failure 404 {object} map[string]interface{} "Band not found"
// @Router /bands/{id} [get]
func (h *BandHandler) GetBand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bandIDStr := vars["id"]

	bandID, err := uuid.Parse(bandIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid band ID")
		return
	}

	band, err := h.bandService.GetBand(r.Context(), bandID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Band not found")
		return
	}

	utils.WriteSuccess(w, "Band retrieved successfully", band.ToResponse())
}

func (h *BandHandler) UpdateBand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bandIDStr := vars["id"]

	bandID, err := uuid.Parse(bandIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid band ID")
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

	var req models.UpdateBandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Use service to update band
	band, err := h.bandService.UpdateBand(r.Context(), bandID, userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Band updated successfully", band.ToResponse())
}

func (h *BandHandler) DeleteBand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bandIDStr := vars["id"]

	bandID, err := uuid.Parse(bandIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid band ID")
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

	// Use service to delete band
	if err := h.bandService.DeleteBand(r.Context(), bandID, userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Band deleted successfully", nil)
}

func (h *BandHandler) JoinBand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bandIDStr := vars["id"]

	bandID, err := uuid.Parse(bandIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid band ID")
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

	// Use service to join band
	if err := h.bandService.JoinBand(r.Context(), bandID, userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Successfully joined band", nil)
}

func (h *BandHandler) LeaveBand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bandIDStr := vars["id"]

	bandID, err := uuid.Parse(bandIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid band ID")
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

	// Use service to leave band
	if err := h.bandService.LeaveBand(r.Context(), bandID, userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, "Successfully left band", nil)
}

// @Summary Get band members
// @Description Get all members of a band
// @Tags Bands
// @Accept json
// @Produce json
// @Param id path string true "Band ID"
// @Success 200 {array} models.BandMember "Band members retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid band ID"
// @Router /bands/{id}/members [get]
func (h *BandHandler) GetBandMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bandIDStr := vars["id"]

	bandID, err := uuid.Parse(bandIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid band ID")
		return
	}

	// Use service to get band members
	members, err := h.bandService.GetBandMembers(r.Context(), bandID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve band members")
		return
	}

	utils.WriteSuccess(w, "Band members retrieved successfully", members)
}

func (h *BandHandler) GetNearbyBands(w http.ResponseWriter, r *http.Request) {
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

	// Use service to get nearby bands
	bands, err := h.bandService.GetNearbyBands(r.Context(), lat, lng, radius, limit)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Convert to response format
	var bandResponses []*models.BandResponse
	for _, band := range bands {
		bandResponses = append(bandResponses, band.ToResponse())
	}

	utils.WriteSuccess(w, "Nearby bands retrieved successfully", bandResponses)
}

func (h *BandHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bandIDStr := vars["id"]

	bandID, err := uuid.Parse(bandIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid band ID")
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
	profilePictureURL, err := h.bandService.UploadProfilePicture(r.Context(), bandID, userID, handler.Filename, fileData)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"profile_picture_url": profilePictureURL,
	}

	utils.WriteSuccess(w, "Profile picture uploaded successfully", response)
}

// @Summary Get all bands
// @Description Get all bands with pagination
// @Tags Bands
// @Accept json
// @Produce json
// @Param limit query int false "Number of bands to return (default: 20, max: 100)" default(20)
// @Param offset query int false "Number of bands to skip (default: 0)" default(0)
// @Success 200 {object} map[string]interface{} "Bands retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid pagination parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /bands [get]
func (h *BandHandler) GetAllBands(w http.ResponseWriter, r *http.Request) {
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

	bands, err := h.bandService.GetAllBands(r.Context(), limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve bands")
		return
	}

	// Convert to response format
	var bandResponses []*models.BandResponse
	for _, band := range bands {
		bandResponses = append(bandResponses, band.ToResponse())
	}

	response := map[string]interface{}{
		"bands":  bandResponses,
		"limit":  limit,
		"offset": offset,
		"count":  len(bandResponses),
	}

	utils.WriteSuccess(w, "Bands retrieved successfully", response)
}
