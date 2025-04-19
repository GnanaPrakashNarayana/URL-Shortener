package repository

import (
	"context"
	"database/sql"
	"errors"
	"time" // We'll make sure to use this

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresRepository is a PostgreSQL implementation of the Repository interface
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	return &PostgresRepository{
		db: db,
	}, nil
}

// Store stores a URL in the repository
func (r *PostgresRepository) Store(ctx context.Context, url *models.URL) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// If LastVisitAt is zero, set it to NULL
	var lastVisitAt interface{}
	if url.LastVisitAt.IsZero() {
		lastVisitAt = nil
	} else {
		lastVisitAt = url.LastVisitAt
	}

	// Insert the URL - using time values for created_at and last_visit_at
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO urls (id, original_url, created_at, visits, last_visit_at) VALUES ($1, $2, $3, $4, $5)",
		url.ID,
		url.OriginalURL,
		url.CreatedAt,  // This uses time.Time
		url.Visits,
		lastVisitAt,    // This uses time.Time
	)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// GetByID retrieves a URL by its ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*models.URL, error) {
	// Query the URL
	var url models.URL
	var lastVisitAt sql.NullTime  // This uses time.Time

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, original_url, created_at, visits, last_visit_at FROM urls WHERE id = $1",
		id,
	).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.CreatedAt,     // This uses time.Time
		&url.Visits,
		&lastVisitAt,       // This uses time.Time
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Set LastVisitAt if not NULL
	if lastVisitAt.Valid {
		url.LastVisitAt = lastVisitAt.Time  // This uses time.Time
	} else {
		url.LastVisitAt = time.Time{}  // Using time package explicitly here
	}

	return &url, nil
}

// Update updates a URL in the repository
func (r *PostgresRepository) Update(ctx context.Context, url *models.URL) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// If LastVisitAt is zero, set it to NULL
	var lastVisitAt interface{}
	if url.LastVisitAt.IsZero() {
		lastVisitAt = nil
	} else {
		lastVisitAt = url.LastVisitAt  // This uses time.Time
	}

	// Update the URL
	result, err := tx.ExecContext(
		ctx,
		"UPDATE urls SET original_url = $1, visits = $2, last_visit_at = $3 WHERE id = $4",
		url.OriginalURL,
		url.Visits,
		lastVisitAt,    // This uses time.Time
		url.ID,
	)
	if err != nil {
		return err
	}

	// Check if the URL was updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Commit the transaction
	return tx.Commit()
}

// List lists all URLs in the repository
func (r *PostgresRepository) List(ctx context.Context) ([]*models.URL, error) {
	// Query all URLs
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, original_url, created_at, visits, last_visit_at FROM urls ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the rows
	urls := []*models.URL{}
	for rows.Next() {
		var url models.URL
		var lastVisitAt sql.NullTime  // This uses time.Time

		err := rows.Scan(
			&url.ID,
			&url.OriginalURL,
			&url.CreatedAt,     // This uses time.Time
			&url.Visits,
			&lastVisitAt,       // This uses time.Time
		)
		if err != nil {
			return nil, err
		}

		// Set LastVisitAt if not NULL
		if lastVisitAt.Valid {
			url.LastVisitAt = lastVisitAt.Time  // This uses time.Time
		} else {
			url.LastVisitAt = time.Time{}  // Using time package explicitly
		}

		urls = append(urls, &url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// Close closes the repository
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}