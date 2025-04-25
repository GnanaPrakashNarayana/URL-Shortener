package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/config"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Failed to load .env file: %v\n", err)
	}

	// Print current working directory to help with debugging
	pwd, _ := os.Getwd()
	log.Printf("Current directory: %s\n", pwd)

	// PART 1: Fix the database directly
	fixDatabase()

	// PART 2: Test connection with migration
	testDatabase()
}

// fixDatabase fixes issues with the database migrations
func fixDatabase() {
	log.Println("=== RUNNING DATABASE FIX ===")

	// Get database connection string from environment
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://gnanaprakashnarayanairivisetty@localhost:5432/url_shortener?sslmode=disable"
	}

	// Connect to the database
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Successfully connected to the database")

	// Check if the schema_migrations table exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if migrations table exists: %v", err)
	}

	if !exists {
		fmt.Println("schema_migrations table does not exist. Creating it...")
		_, err := db.Exec(`CREATE TABLE schema_migrations (
			version bigint NOT NULL,
			dirty boolean NOT NULL,
			CONSTRAINT schema_migrations_pkey PRIMARY KEY (version)
		)`)
		if err != nil {
			log.Fatalf("Failed to create schema_migrations table: %v", err)
		}

		// Insert a clean migration at version 0
		_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (0, false)")
		if err != nil {
			log.Fatalf("Failed to insert initial migration record: %v", err)
		}

		fmt.Println("Created schema_migrations table and initialized with version 0")
	} else {
		// Get current migration version and status
		var version int64
		var dirty bool
		err = db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}

		fmt.Printf("Current migration version: %d, Dirty: %t\n", version, dirty)

		// Fix dirty migration if needed
		if dirty {
			fmt.Printf("Fixing dirty migration at version %d\n", version)
			_, err = db.Exec("UPDATE schema_migrations SET dirty = false WHERE version = $1", version)
			if err != nil {
				log.Fatalf("Failed to fix dirty migration: %v", err)
			}
			fmt.Println("Successfully fixed dirty migration")
		}
	}

	log.Println("=== DATABASE FIX COMPLETED ===")
}

// testDatabase tests the database connection and migrations
func testDatabase() {
	log.Println("=== RUNNING DATABASE TEST ===")

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
	log.Println("=== DATABASE TEST COMPLETED ===")
}
