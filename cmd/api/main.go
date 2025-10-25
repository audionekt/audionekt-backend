// @title Music Producer Social Network API
// @version 1.0
// @description A social media platform for music producers to connect, collaborate, and discover nearby talent through location-based matching
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/yourusername/musicapp
// @contact.email support@musicapp.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"

	"musicapp/cmd/api/server"
	"musicapp/internal/config"
)

//go:noinline
func main() {
	// Load configuration
	cfg := config.Load()

	// Create and start server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
