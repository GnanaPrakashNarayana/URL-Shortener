package repository

import (
	"context"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
)

// UserRepository defines the interface for user storage
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int) (*models.User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Update updates a user
	Update(ctx context.Context, user *models.User) error

	// Delete deletes a user
	Delete(ctx context.Context, id int) error

	// List lists all users
	List(ctx context.Context) ([]*models.User, error)

	// CreateOAuthAccount creates a new OAuth account
	CreateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error

	// GetUserByOAuthAccount retrieves a user by OAuth account
	GetUserByOAuthAccount(ctx context.Context, provider, providerUserID string) (*models.User, error)
}