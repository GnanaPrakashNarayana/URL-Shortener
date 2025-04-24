package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
)

// MemoryBioPageRepository is an in-memory implementation of the BioPageRepository interface
type MemoryBioPageRepository struct {
	bioPages     map[int]*models.BioPage
	bioLinks     map[int]*models.BioLink
	bioPagesMux  sync.RWMutex
	bioLinksMux  sync.RWMutex
	nextBioPageID int
	nextBioLinkID int
}

// NewMemoryBioPageRepository creates a new in-memory bio page repository
func NewMemoryBioPageRepository() *MemoryBioPageRepository {
	return &MemoryBioPageRepository{
		bioPages:      make(map[int]*models.BioPage),
		bioLinks:      make(map[int]*models.BioLink),
		nextBioPageID: 1,
		nextBioLinkID: 1,
	}
}

// CreateBioPage creates a new bio page
func (r *MemoryBioPageRepository) CreateBioPage(ctx context.Context, bioPage *models.BioPage) error {
	r.bioPagesMux.Lock()
	defer r.bioPagesMux.Unlock()

	// Check if short code is already in use
	for _, existingPage := range r.bioPages {
		if existingPage.ShortCode == bioPage.ShortCode {
			return ErrSlugUnavailable
		}
	}

	// Assign an ID
	bioPage.ID = r.nextBioPageID
	r.nextBioPageID++

	// Store the bio page
	r.bioPages[bioPage.ID] = bioPage

	return nil
}

// GetBioPageByID retrieves a bio page by ID
func (r *MemoryBioPageRepository) GetBioPageByID(ctx context.Context, id int) (*models.BioPage, error) {
	r.bioPagesMux.RLock()
	defer r.bioPagesMux.RUnlock()

	bioPage, ok := r.bioPages[id]
	if !ok {
		return nil, ErrNotFound
	}

	// Get the links for this bio page
	bioPage.Links, _ = r.ListBioLinksByBioPageID(ctx, bioPage.ID)

	return bioPage, nil
}

// GetBioPageByShortCode retrieves a bio page by short code
func (r *MemoryBioPageRepository) GetBioPageByShortCode(ctx context.Context, shortCode string) (*models.BioPage, error) {
	r.bioPagesMux.RLock()
	defer r.bioPagesMux.RUnlock()

	for _, bioPage := range r.bioPages {
		if bioPage.ShortCode == shortCode {
			// Get the links for this bio page
			bioPage.Links, _ = r.ListBioLinksByBioPageID(ctx, bioPage.ID)
			return bioPage, nil
		}
	}

	return nil, ErrNotFound
}

// ListBioPagesByUserID lists all bio pages for a user
func (r *MemoryBioPageRepository) ListBioPagesByUserID(ctx context.Context, userID int) ([]*models.BioPage, error) {
	r.bioPagesMux.RLock()
	defer r.bioPagesMux.RUnlock()

	bioPages := []*models.BioPage{}
	for _, bioPage := range r.bioPages {
		if bioPage.UserID == userID {
			// Get the links for this bio page
			bioPage.Links, _ = r.ListBioLinksByBioPageID(ctx, bioPage.ID)
			bioPages = append(bioPages, bioPage)
		}
	}

	// Sort by creation date (newest first)
	sort.Slice(bioPages, func(i, j int) bool {
		return bioPages[i].CreatedAt.After(bioPages[j].CreatedAt)
	})

	return bioPages, nil
}

// UpdateBioPage updates a bio page
func (r *MemoryBioPageRepository) UpdateBioPage(ctx context.Context, bioPage *models.BioPage) error {
	r.bioPagesMux.Lock()
	defer r.bioPagesMux.Unlock()

	existingBioPage, ok := r.bioPages[bioPage.ID]
	if !ok {
		return ErrNotFound
	}

	// Update the timestamp
	bioPage.UpdatedAt = time.Now()

	// Preserve the created date and user ID
	bioPage.CreatedAt = existingBioPage.CreatedAt
	bioPage.UserID = existingBioPage.UserID
	bioPage.ShortCode = existingBioPage.ShortCode

	// Update the bio page
	r.bioPages[bioPage.ID] = bioPage

	return nil
}

// DeleteBioPage deletes a bio page
func (r *MemoryBioPageRepository) DeleteBioPage(ctx context.Context, id int) error {
	r.bioPagesMux.Lock()
	defer r.bioPagesMux.Unlock()

	if _, ok := r.bioPages[id]; !ok {
		return ErrNotFound
	}

	// Delete the bio page
	delete(r.bioPages, id)

	// Delete all links associated with this bio page
	r.bioLinksMux.Lock()
	defer r.bioLinksMux.Unlock()
	for linkID, link := range r.bioLinks {
		if link.BioPageID == id {
			delete(r.bioLinks, linkID)
		}
	}

	return nil
}

// CreateBioLink creates a new bio link
func (r *MemoryBioPageRepository) CreateBioLink(ctx context.Context, bioLink *models.BioLink) error {
	r.bioLinksMux.Lock()
	defer r.bioLinksMux.Unlock()

	// Check if the bio page exists
	r.bioPagesMux.RLock()
	_, ok := r.bioPages[bioLink.BioPageID]
	r.bioPagesMux.RUnlock()
	if !ok {
		return ErrNotFound
	}

	// Assign an ID
	bioLink.ID = r.nextBioLinkID
	r.nextBioLinkID++

	// Store the bio link
	r.bioLinks[bioLink.ID] = bioLink

	return nil
}

// GetBioLinkByID retrieves a bio link by ID
func (r *MemoryBioPageRepository) GetBioLinkByID(ctx context.Context, id int) (*models.BioLink, error) {
	r.bioLinksMux.RLock()
	defer r.bioLinksMux.RUnlock()

	bioLink, ok := r.bioLinks[id]
	if !ok {
		return nil, ErrNotFound
	}

	return bioLink, nil
}

// ListBioLinksByBioPageID lists all bio links for a bio page
func (r *MemoryBioPageRepository) ListBioLinksByBioPageID(ctx context.Context, bioPageID int) ([]*models.BioLink, error) {
	r.bioLinksMux.RLock()
	defer r.bioLinksMux.RUnlock()

	bioLinks := []*models.BioLink{}
	for _, bioLink := range r.bioLinks {
		if bioLink.BioPageID == bioPageID {
			bioLinks = append(bioLinks, bioLink)
		}
	}

	// Sort by display order
	sort.Slice(bioLinks, func(i, j int) bool {
		return bioLinks[i].DisplayOrder < bioLinks[j].DisplayOrder
	})

	return bioLinks, nil
}

// UpdateBioLink updates a bio link
func (r *MemoryBioPageRepository) UpdateBioLink(ctx context.Context, bioLink *models.BioLink) error {
	r.bioLinksMux.Lock()
	defer r.bioLinksMux.Unlock()

	if _, ok := r.bioLinks[bioLink.ID]; !ok {
		return ErrNotFound
	}

	// Update the timestamp
	bioLink.UpdatedAt = time.Now()

	// Update the bio link
	r.bioLinks[bioLink.ID] = bioLink

	return nil
}

// DeleteBioLink deletes a bio link
func (r *MemoryBioPageRepository) DeleteBioLink(ctx context.Context, id int) error {
	r.bioLinksMux.Lock()
	defer r.bioLinksMux.Unlock()

	if _, ok := r.bioLinks[id]; !ok {
		return ErrNotFound
	}

	// Delete the bio link
	delete(r.bioLinks, id)

	return nil
}

// ReorderBioLinks updates the display order of bio links
func (r *MemoryBioPageRepository) ReorderBioLinks(ctx context.Context, bioPageID int, linkIDs []int) error {
	r.bioLinksMux.Lock()
	defer r.bioLinksMux.Unlock()

	// Update the display order for each link
	for i, linkID := range linkIDs {
		bioLink, ok := r.bioLinks[linkID]
		if !ok || bioLink.BioPageID != bioPageID {
			return ErrNotFound
		}

		bioLink.DisplayOrder = i
		bioLink.UpdatedAt = time.Now()
		r.bioLinks[linkID] = bioLink
	}

	return nil
}