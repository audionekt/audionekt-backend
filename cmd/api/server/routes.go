package server

import (
	"context"
	"net/http"
	"time"

	"musicapp/internal/middleware"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// setupRoutes configures all application routes
func setupRoutes(deps *Dependencies) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(deps.LoggingMiddleware.Logging)
	router.Use(middleware.CORS)

	// Health check endpoints
	setupHealthRoutes(router, deps)

	// Swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	setupAuthRoutes(api, deps)
	setupUserRoutes(api, deps)
	setupBandRoutes(api, deps)
	setupPostRoutes(api, deps)
	setupFollowRoutes(api, deps)
	setupFeedRoutes(api, deps)

	return router
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(router *mux.Router, deps *Dependencies) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Check database
		if err := deps.Database.HealthCheck(ctx); err != nil {
			http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
			return
		}

		// Check Redis
		if err := deps.Redis.HealthCheck(ctx); err != nil {
			http.Error(w, "Redis unhealthy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}).Methods("GET")
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(api *mux.Router, deps *Dependencies) {
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", deps.AuthHandler.Register).Methods("POST")
	auth.HandleFunc("/login", deps.AuthHandler.Login).Methods("POST")
	auth.Handle("/logout", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.AuthHandler.Logout))).Methods("POST")
}

// setupUserRoutes configures user routes
func setupUserRoutes(api *mux.Router, deps *Dependencies) {
	users := api.PathPrefix("/users").Subrouter()
	users.HandleFunc("", deps.UserHandler.GetAllUsers).Methods("GET")
	users.HandleFunc("/{id}", deps.UserHandler.GetUser).Methods("GET")
	users.Handle("/{id}", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.UserHandler.UpdateUser))).Methods("PUT")
	users.HandleFunc("/{id}/posts", deps.UserHandler.GetUserPosts).Methods("GET")
	users.HandleFunc("/{id}/followers", deps.UserHandler.GetFollowers).Methods("GET")
	users.HandleFunc("/{id}/following", deps.UserHandler.GetFollowing).Methods("GET")
	users.HandleFunc("/{id}/bands", deps.UserHandler.GetUserBands).Methods("GET")
	users.HandleFunc("/nearby", deps.UserHandler.GetNearbyUsers).Methods("GET")
	users.Handle("/{id}/profile-picture", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.UserHandler.UploadProfilePicture))).Methods("POST")
}

// setupBandRoutes configures band routes
func setupBandRoutes(api *mux.Router, deps *Dependencies) {
	bands := api.PathPrefix("/bands").Subrouter()
	bands.HandleFunc("", deps.BandHandler.GetAllBands).Methods("GET")
	bands.Handle("", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BandHandler.CreateBand))).Methods("POST")
	bands.HandleFunc("/{id}", deps.BandHandler.GetBand).Methods("GET")
	bands.Handle("/{id}", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BandHandler.UpdateBand))).Methods("PUT")
	bands.Handle("/{id}", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BandHandler.DeleteBand))).Methods("DELETE")
	bands.Handle("/{id}/join", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BandHandler.JoinBand))).Methods("POST")
	bands.Handle("/{id}/leave", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BandHandler.LeaveBand))).Methods("POST")
	bands.HandleFunc("/{id}/members", deps.BandHandler.GetBandMembers).Methods("GET")
	bands.HandleFunc("/nearby", deps.BandHandler.GetNearbyBands).Methods("GET")
	bands.Handle("/{id}/profile-picture", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BandHandler.UploadProfilePicture))).Methods("POST")
}

// setupPostRoutes configures post routes
func setupPostRoutes(api *mux.Router, deps *Dependencies) {
	posts := api.PathPrefix("/posts").Subrouter()
	posts.HandleFunc("", deps.PostHandler.GetAllPosts).Methods("GET")
	posts.Handle("", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.CreatePost))).Methods("POST")
	posts.HandleFunc("/{id}", deps.PostHandler.GetPost).Methods("GET")
	posts.Handle("/{id}", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.UpdatePost))).Methods("PUT")
	posts.Handle("/{id}", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.DeletePost))).Methods("DELETE")
	posts.Handle("/{id}/like", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.LikePost))).Methods("POST")
	posts.Handle("/{id}/like", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.UnlikePost))).Methods("DELETE")
	posts.Handle("/{id}/repost", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.Repost))).Methods("POST")
	posts.Handle("/{id}/media", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.UploadMedia))).Methods("POST")
}

// setupFollowRoutes configures follow routes
func setupFollowRoutes(api *mux.Router, deps *Dependencies) {
	follows := api.PathPrefix("/follow").Subrouter()
	follows.Handle("", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.FollowHandler.Follow))).Methods("POST")
	follows.Handle("", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.FollowHandler.Unfollow))).Methods("DELETE")
}

// setupFeedRoutes configures feed routes
func setupFeedRoutes(api *mux.Router, deps *Dependencies) {
	feed := api.PathPrefix("/feed").Subrouter()
	feed.Handle("", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PostHandler.GetFeed))).Methods("GET")
	feed.HandleFunc("/explore", deps.PostHandler.GetExploreFeed).Methods("GET")
}
