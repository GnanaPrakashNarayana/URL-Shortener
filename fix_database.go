package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func FixDatabase() {
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

	// Check tables and recreate if necessary
	tables := []string{"urls", "users", "oauth_accounts", "bio_pages", "bio_links"}
	for _, table := range tables {
		err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", table).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check if %s table exists: %v", table, err)
		}
		
		fmt.Printf("Table %s exists: %t\n", table, exists)
	}

	fmt.Println("Database check complete")
}