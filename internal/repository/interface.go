package repository

import (
	"context"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
)

// Repository defines the interface for URL storage
type Repository interface {
	// Store stores a URL in the repository
	Store(ctx context.Context, url *models.URL) error

	// GetByID retrieves a URL by its ID
	GetByID(ctx context.Context, id string) (*models.URL, error)

	// Update updates a URL in the repository
	Update(ctx context.Context, url *models.URL) error

	// List lists all URLs
	List(ctx context.Context) ([]*models.URL, error)

	// ListByUserID lists all URLs for a user
	ListByUserID(ctx context.Context, userID int) ([]*models.URL, error)

	// Close closes the repository
	Close() error
}