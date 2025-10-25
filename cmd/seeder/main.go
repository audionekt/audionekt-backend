package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"musicapp/internal/config"
	"musicapp/internal/db"
	"musicapp/internal/seeder"
)

func main() {
	// Parse command line flags
	var (
		dropDB = flag.Bool("drop", false, "Drop all tables before seeding")
		help   = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Drop database if requested
	if *dropDB {
		log.Println("🗑️  Dropping all tables...")
		if err := dropAllTables(database); err != nil {
			log.Fatalf("Failed to drop tables: %v", err)
		}
		log.Println("✅ Tables dropped successfully")
	}

	// Create seeder
	seeder := seeder.New(database)

	// Seed the database
	ctx := context.Background()
	if err := seeder.SeedAll(ctx); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("🎉 Seeding completed successfully!")
}

// dropAllTables drops all tables in the database
func dropAllTables(database *db.DB) error {
	tables := []string{
		"jwt_blacklist",
		"messages",
		"comments",
		"reposts",
		"likes",
		"follows",
		"posts",
		"band_members",
		"bands",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if _, err := database.Pool.Exec(context.Background(), query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
