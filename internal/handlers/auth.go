package handlers

import (
	"encoding/json"
	"net/http"

	"musicapp/internal/middleware"
	"musicapp/internal/models"
	"musicapp/internal/service"
	"musicapp/pkg/utils"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Username string           `json:"username" validate:"required,min=3,max=50"`
	Email    string           `json:"email" validate:"required,email"`
	Password string           `json:"password" validate:"required,min=8"`
	Location *models.Location `json:"location,omitempty"`
	City     *string          `json:"city,omitempty"`
	Country  *string          `json:"country,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string               `json:"token"`
	User  *models.UserResponse `json:"user"`
}

// @Summary Register a new user
// @Description Create a new user account with profile information
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} AuthResponse "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "User already exists"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Convert to service request
	createReq := &models.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Location: req.Location,
		City:     req.City,
		Country:  req.Country,
	}

	// Use service to register user
	user, token, err := h.authService.RegisterUser(r.Context(), createReq)
	if err != nil {
		if err.Error() == "user with email "+req.Email+" already exists" {
			utils.WriteError(w, http.StatusConflict, "User with this email already exists")
			return
		}
		if err.Error() == "username "+req.Username+" already taken" {
			utils.WriteError(w, http.StatusConflict, "Username already taken")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	response := AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	utils.WriteCreated(w, "User registered successfully", response)
}

// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Use service to login user
	user, token, err := h.authService.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	response := AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	utils.WriteSuccess(w, "Login successful", response)
}

// @Summary Logout user
// @Description Logout user and invalidate JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 401 {object} map[string]interface{} "Invalid token"
// @Failure 500 {object} map[string]interface{} "Failed to logout"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get JTI and user ID from context (set by auth middleware)
	jti, ok := middleware.GetJTIFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Use service to logout user
	if err := h.authService.LogoutUser(r.Context(), jti, userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	utils.WriteSuccess(w, "Logout successful", nil)
}
