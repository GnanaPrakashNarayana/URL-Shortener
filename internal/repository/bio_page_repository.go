package repository

import (
	"context"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
)

// BioPageRepository defines the interface for bio page storage
type BioPageRepository interface {
	// CreateBioPage creates a new bio page
	CreateBioPage(ctx context.Context, bioPage *models.BioPage) error

	// GetBioPageByID retrieves a bio page by ID
	GetBioPageByID(ctx context.Context, id int) (*models.BioPage, error)

	// GetBioPageByShortCode retrieves a bio page by short code
	GetBioPageByShortCode(ctx context.Context, shortCode string) (*models.BioPage, error)

	// ListBioPagesByUserID lists all bio pages for a user
	ListBioPagesByUserID(ctx context.Context, userID int) ([]*models.BioPage, error)

	// UpdateBioPage updates a bio page
	UpdateBioPage(ctx context.Context, bioPage *models.BioPage) error

	// DeleteBioPage deletes a bio page
	DeleteBioPage(ctx context.Context, id int) error

	// CreateBioLink creates a new bio link
	CreateBioLink(ctx context.Context, bioLink *models.BioLink) error

	// GetBioLinkByID retrieves a bio link by ID
	GetBioLinkByID(ctx context.Context, id int) (*models.BioLink, error)

	// ListBioLinksByBioPageID lists all bio links for a bio page
	ListBioLinksByBioPageID(ctx context.Context, bioPageID int) ([]*models.BioLink, error)

	// UpdateBioLink updates a bio link
	UpdateBioLink(ctx context.Context, bioLink *models.BioLink) error

	// DeleteBioLink deletes a bio link
	DeleteBioLink(ctx context.Context, id int) error

	// ReorderBioLinks updates the display order of bio links
	ReorderBioLinks(ctx context.Context, bioPageID int, linkIDs []int) error
}