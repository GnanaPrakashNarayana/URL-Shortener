package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Manager handles database connections and migrations
type Manager struct {
	db     *sql.DB
	config *config.DatabaseConfig
}

// NewManager creates a new database manager
func NewManager(config *config.DatabaseConfig) (*Manager, error) {
	return &Manager{
		config: config,
	}, nil
}

// Connect connects to the database
func (m *Manager) Connect() (*sql.DB, error) {
	if m.db != nil {
		return m.db, nil
	}

	if m.config.Type != "postgres" {
		return nil, errors.New("unsupported database type")
	}

	// Connect to the database
	db, err := sql.Open("postgres", m.config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure the connection pool
	db.SetMaxOpenConns(m.config.MaxOpenConns)
	db.SetMaxIdleConns(m.config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(m.config.ConnMaxLifetime) * time.Second)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	m.db = db
	return db, nil
}

// Migrate runs database migrations
func (m *Manager) Migrate() error {
	if m.db == nil {
		return errors.New("database not connected")
	}

	log.Println("Starting database migration process...")

	// First check if the schema_migrations table exists and if it's dirty
	var exists bool
	err := m.db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if migrations table exists: %w", err)
	}

	// If table exists, check for dirty status
	if exists {
		var dirty bool
		var version int
		err = m.db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
		if err == nil {
			log.Printf("Migration table exists. Current version: %d, Dirty: %t\n", version, dirty)
			
			if dirty {
				log.Printf("Found dirty migration at version %d, fixing...\n", version)
				// Clean up the dirty migration
				_, err = m.db.Exec("UPDATE schema_migrations SET dirty = false WHERE version = $1", version)
				if err != nil {
					return fmt.Errorf("failed to clean dirty migration: %w", err)
				}
				log.Println("Successfully cleaned dirty migration state")
			}
		} else {
			log.Printf("Could not get migration version: %v\n", err)
		}
	} else {
		log.Println("Migrations table does not exist yet - it will be created during migration")
	}

	// Create a new instance of the postgres driver
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create a new migrate instance
	migrationsPath := fmt.Sprintf("file://%s", m.config.MigrationsPath)
	log.Printf("Using migrations from: %s\n", migrationsPath)
	
	migrator, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Get current version before running migrations
	version, _, vErr := migrator.Version()  // Changed "dirty" to "_" since it's not used
	if vErr != nil && vErr != migrate.ErrNilVersion {
		log.Printf("Warning: Could not get migration version: %v\n", vErr)
	} else if vErr == nil {
		log.Printf("Before migration - Version: %d\n", version)
	}

	// Run migrations
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("Migration error: %v\n", err)
		
		// Check if it's a dirty database error (compare error strings instead of types)
		if err.Error() == "Dirty database version 3. Fix and force version." {
			version, _, vErr := migrator.Version()  // Changed "dirty" to "_" since it's not used
			if vErr != nil {
				log.Printf("Warning: Could not get migration version: %v\n", vErr)
			} else {
				log.Printf("Found dirty migration at version %d, forcing version...\n", version)
				fErr := migrator.Force(int(version))
				if fErr != nil {
					return fmt.Errorf("failed to force migration version: %w", fErr)
				}
				log.Printf("Successfully forced migration to version %d\n", version)
				
				// Try running migrations again
				upErr := migrator.Up()
				if upErr != nil && upErr != migrate.ErrNoChange {
					return fmt.Errorf("failed to run migrations after forcing version: %w", upErr)
				} else {
					log.Println("Migrations completed successfully after forcing version")
					return nil
				}
			}
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	} else if err == migrate.ErrNoChange {
		log.Println("No migrations to apply")
	} else {
		log.Println("Migrations completed successfully")
	}

	return nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m.db == nil {
		return nil
	}
	return m.db.Close()
}