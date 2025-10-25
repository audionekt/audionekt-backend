package server

import (
	"log"

	"musicapp/internal/cache"
	"musicapp/internal/config"
	"musicapp/internal/db"
	"musicapp/internal/handlers"
	"musicapp/internal/middleware"
	"musicapp/internal/repository"
	"musicapp/internal/service"
	"musicapp/internal/storage"
)

// Dependencies holds all application dependencies
type Dependencies struct {
	// Infrastructure
	Database *db.DB
	Redis    *cache.Cache
	S3       *storage.S3Client

	// Repositories
	UserRepo   *repository.UserRepository
	BandRepo   *repository.BandRepository
	PostRepo   *repository.PostRepository
	FollowRepo *repository.FollowRepository

	// Services
	AuthService   *service.AuthService
	UserService   *service.UserService
	BandService   *service.BandService
	PostService   *service.PostService
	FollowService *service.FollowService

	// Handlers
	AuthHandler   *handlers.AuthHandler
	UserHandler   *handlers.UserHandler
	BandHandler   *handlers.BandHandler
	PostHandler   *handlers.PostHandler
	FollowHandler *handlers.FollowHandler

	// Middleware
	AuthMiddleware    *middleware.AuthMiddleware
	LoggingMiddleware *middleware.LoggingMiddleware
}

// setupDependencies initializes all application dependencies
func setupDependencies(cfg *config.Config) (*Dependencies, error) {
	// Initialize database
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Initialize Redis
	redisCache, err := cache.New(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	// Initialize S3 client (optional)
	var s3Client *storage.S3Client
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" && cfg.S3BucketName != "" {
		s3Client, err = storage.NewS3Client(cfg.AWSRegion, cfg.S3BucketName, cfg.S3CDNURL)
		if err != nil {
			log.Printf("Warning: Failed to initialize S3 client: %v", err)
		} else {
			log.Println("S3 client initialized successfully")
		}
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, redisCache)
	loggingMiddleware := middleware.NewLoggingMiddleware()

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	bandRepo := repository.NewBandRepository(database)
	postRepo := repository.NewPostRepository(database)
	followRepo := repository.NewFollowRepository(database)

	// Initialize services
	authService := service.NewAuthService(userRepo, redisCache, authMiddleware)
	userService := service.NewUserService(userRepo, redisCache, s3Client)
	bandService := service.NewBandService(bandRepo, userRepo, redisCache, s3Client)
	postService := service.NewPostService(postRepo, userRepo, bandRepo, redisCache, s3Client)
	followService := service.NewFollowService(followRepo, userRepo, bandRepo, redisCache)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, bandService)
	bandHandler := handlers.NewBandHandler(bandService)
	postHandler := handlers.NewPostHandler(postService)
	followHandler := handlers.NewFollowHandler(followService)

	return &Dependencies{
		// Infrastructure
		Database: database,
		Redis:    redisCache,
		S3:       s3Client,

		// Repositories
		UserRepo:   userRepo,
		BandRepo:   bandRepo,
		PostRepo:   postRepo,
		FollowRepo: followRepo,

		// Services
		AuthService:   authService,
		UserService:   userService,
		BandService:   bandService,
		PostService:   postService,
		FollowService: followService,

		// Handlers
		AuthHandler:   authHandler,
		UserHandler:   userHandler,
		BandHandler:   bandHandler,
		PostHandler:   postHandler,
		FollowHandler: followHandler,

		// Middleware
		AuthMiddleware:    authMiddleware,
		LoggingMiddleware: loggingMiddleware,
	}, nil
}

// Close closes all dependencies
func (d *Dependencies) Close() {
	if d.Database != nil {
		d.Database.Close()
	}
	if d.Redis != nil {
		d.Redis.Close()
	}
}
