package database

import (
	"database/sql"
	"errors"
	"fmt"
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

	m.db = db
	return db, nil
}

// Migrate runs database migrations
func (m *Manager) Migrate() error {
	if m.db == nil {
		return errors.New("database not connected")
	}

	// Create a new instance of the postgres driver
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create a new migrate instance
	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", m.config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migrations
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
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