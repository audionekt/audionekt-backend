package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "musicapp/docs"
	"musicapp/internal/config"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	server *http.Server
	deps   *Dependencies
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Setup dependencies
	deps, err := setupDependencies(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup dependencies: %w", err)
	}

	// Setup routes
	router := setupRoutes(deps)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		config: cfg,
		server: httpServer,
		deps:   deps,
	}, nil
}

// Start starts the server and handles graceful shutdown
// This function is not suitable for unit testing as it blocks indefinitely
//
//go:noinline
func (s *Server) Start() error {
	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}

// Close closes all dependencies
func (s *Server) Close() {
	if s.deps != nil {
		s.deps.Close()
	}
}
