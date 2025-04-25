package main

import (
	"log"
	"os"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/config"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/database"
	"github.com/joho/godotenv"
)

func fixDatabase() {
	// Your database fixing code here
	log.Println("=== RUNNING DATABASE FIX ===")
	// ...implementation...
	log.Println("=== DATABASE FIX COMPLETED ===")
}

func main() {
	fixDatabase()
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Failed to load .env file: %v\n", err)
	}

	// Check environment variables
	log.Println("Environment variables:")
	log.Printf("DB_TYPE: %s\n", os.Getenv("DB_TYPE"))
	log.Printf("DB_DSN: %s\n", os.Getenv("DB_DSN"))
	log.Printf("DB_MIGRATIONS_PATH: %s\n", os.Getenv("DB_MIGRATIONS_PATH"))

	// Create database config
	dbConfig := &config.DatabaseConfig{
		Type:            os.Getenv("DB_TYPE"),
		DSN:             os.Getenv("DB_DSN"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300,
		MigrationsPath:  os.Getenv("DB_MIGRATIONS_PATH"),
	}

	// Check if values are empty, set defaults
	if dbConfig.Type == "" {
		dbConfig.Type = "postgres"
		log.Println("Using default DB_TYPE: postgres")
	}
	if dbConfig.DSN == "" {
		dbConfig.DSN = "postgres://gnanaprakashnarayanairivisetty@localhost:5432/url_shortener?sslmode=disable"
		log.Println("Using default DB_DSN")
	}
	if dbConfig.MigrationsPath == "" {
		dbConfig.MigrationsPath = "migrations"
		log.Println("Using default migrations path: migrations")
	}

	// Create database manager
	manager, err := database.NewManager(dbConfig)
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}

	// Connect to the database
	_, err = manager.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer manager.Close()

	log.Println("Successfully connected to database!")

	// List migrations directory
	files, err := os.ReadDir(dbConfig.MigrationsPath)
	if err != nil {
		log.Printf("Warning: Failed to read migrations directory: %v\n", err)
	} else {
		log.Println("Migration files:")
		for _, file := range files {
			log.Printf("- %s\n", file.Name())
		}
	}

	// Try running migrations
	log.Println("Running migrations...")
	err = manager.Migrate()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Test completed successfully!")
}
