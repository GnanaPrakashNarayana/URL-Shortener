package services

import (
	"context"
	"errors"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
)

// BioPageService handles bio page operations
type BioPageService struct {
	repo    repository.BioPageRepository
	baseURL string
}

// NewBioPageService creates a new bio page service
func NewBioPageService(repo repository.BioPageRepository, baseURL string) *BioPageService {
	return &BioPageService{
		repo:    repo,
		baseURL: baseURL,
	}
}

// CreateBioPage creates a new bio page
func (s *BioPageService) CreateBioPage(ctx context.Context, userID int, shortCode, title, description string) (*models.BioPageResponse, error) {
	if shortCode == "" {
		var err error
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		// Validate custom shortCode
		if err := validateCustomSlug(shortCode); err != nil {
			return nil, err
		}

		// Check if the shortCode is available
		_, err := s.repo.GetBioPageByShortCode(ctx, shortCode)
		if err == nil {
			// ShortCode already exists
			return nil, ErrSlugUnavailable
		} else if err != repository.ErrNotFound {
			// Some other error occurred
			return nil, err
		}
	}

	// Create a new bio page
	bioPage := models.NewBioPage(userID, shortCode, title)
	bioPage.Description = description

	// Store the bio page
	if err := s.repo.CreateBioPage(ctx, bioPage); err != nil {
		return nil, err
	}

	// Return the response
	return bioPage.ToBioPageResponse(s.baseURL), nil
}

// GetBioPage retrieves a bio page by ID
func (s *BioPageService) GetBioPage(ctx context.Context, id int) (*models.BioPageResponse, error) {
	bioPage, err := s.repo.GetBioPageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return bioPage.ToBioPageResponse(s.baseURL), nil
}

// GetBioPageByShortCode retrieves a bio page by short code
func (s *BioPageService) GetBioPageByShortCode(ctx context.Context, shortCode string) (*models.BioPageResponse, error) {
	bioPage, err := s.repo.GetBioPageByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	return bioPage.ToBioPageResponse(s.baseURL), nil
}

// IncrementBioPageVisits increments the visit count for a bio page
func (s *BioPageService) IncrementBioPageVisits(ctx context.Context, id int) error {
	bioPage, err := s.repo.GetBioPageByID(ctx, id)
	if err != nil {
		return err
	}

	bioPage.IncrementVisits()
	return s.repo.UpdateBioPage(ctx, bioPage)
}

// IncrementBioLinkVisits increments the visit count for a bio link
func (s *BioPageService) IncrementBioLinkVisits(ctx context.Context, id int) error {
	bioLink, err := s.repo.GetBioLinkByID(ctx, id)
	if err != nil {
		return err
	}

	bioLink.IncrementVisits()
	return s.repo.UpdateBioLink(ctx, bioLink)
}

// ListBioPagesByUserID lists all bio pages for a user
func (s *BioPageService) ListBioPagesByUserID(ctx context.Context, userID int) ([]*models.BioPageResponse, error) {
	bioPages, err := s.repo.ListBioPagesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]*models.BioPageResponse, 0, len(bioPages))
	for _, bioPage := range bioPages {
		responses = append(responses, bioPage.ToBioPageResponse(s.baseURL))
	}

	return responses, nil
}

// UpdateBioPage updates a bio page
func (s *BioPageService) UpdateBioPage(ctx context.Context, id int, title, description, theme, profileImageURL string, isPublished bool, customCSS string) (*models.BioPageResponse, error) {
	bioPage, err := s.repo.GetBioPageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the fields
	bioPage.Title = title
	bioPage.Description = description
	bioPage.Theme = theme
	bioPage.ProfileImageURL = profileImageURL
	bioPage.IsPublished = isPublished
	bioPage.CustomCSS = customCSS

	// Save the changes
	if err := s.repo.UpdateBioPage(ctx, bioPage); err != nil {
		return nil, err
	}

	return bioPage.ToBioPageResponse(s.baseURL), nil
}

// DeleteBioPage deletes a bio page
func (s *BioPageService) DeleteBioPage(ctx context.Context, id int) error {
	return s.repo.DeleteBioPage(ctx, id)
}

// AddBioLink adds a new link to a bio page
func (s *BioPageService) AddBioLink(ctx context.Context, bioPageID int, title, url string) (*models.BioLinkResponse, error) {
	// Validate URL
	if err := validateURL(url); err != nil {
		return nil, err
	}

	// Get the bio page
	_, err := s.repo.GetBioPageByID(ctx, bioPageID)
	if err != nil {
		return nil, err
	}

	// Get the highest display order
	links, err := s.repo.ListBioLinksByBioPageID(ctx, bioPageID)
	if err != nil {
		return nil, err
	}

	displayOrder := 0
	if len(links) > 0 {
		displayOrder = links[len(links)-1].DisplayOrder + 1
	}

	// Create a new bio link
	bioLink := models.NewBioLink(bioPageID, title, url, displayOrder)

	// Store the bio link
	if err := s.repo.CreateBioLink(ctx, bioLink); err != nil {
		return nil, err
	}

	return bioLink.ToBioLinkResponse(), nil
}

// GetBioLinkByID retrieves a bio link by ID
func (s *BioPageService) GetBioLinkByID(ctx context.Context, id int) (*models.BioLinkResponse, error) {
	bioLink, err := s.repo.GetBioLinkByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return bioLink.ToBioLinkResponse(), nil
}

// UpdateBioLink updates a bio link
func (s *BioPageService) UpdateBioLink(ctx context.Context, id int, title, url string, isEnabled bool) (*models.BioLinkResponse, error) {
	// Validate URL
	if err := validateURL(url); err != nil {
		return nil, err
	}

	// Get the bio link
	bioLink, err := s.repo.GetBioLinkByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the fields
	bioLink.Title = title
	bioLink.URL = url
	bioLink.IsEnabled = isEnabled

	// Save the changes
	if err := s.repo.UpdateBioLink(ctx, bioLink); err != nil {
		return nil, err
	}

	return bioLink.ToBioLinkResponse(), nil
}

// DeleteBioLink deletes a bio link
func (s *BioPageService) DeleteBioLink(ctx context.Context, id int) error {
	return s.repo.DeleteBioLink(ctx, id)
}

// ReorderBioLinks updates the display order of bio links
func (s *BioPageService) ReorderBioLinks(ctx context.Context, bioPageID int, linkIDs []int) error {
	return s.repo.ReorderBioLinks(ctx, bioPageID, linkIDs)
}

// generateUniqueShortCode generates a unique short code for a bio page
func (s *BioPageService) generateUniqueShortCode(ctx context.Context) (string, error) {
	for attempts := 0; attempts < 5; attempts++ {
		shortCode, err := generateRandomString(6)
		if err != nil {
			return "", err
		}

		// Check if the short code is available
		_, err = s.repo.GetBioPageByShortCode(ctx, shortCode)
		if err == repository.ErrNotFound {
			// Short code is available
			return shortCode, nil
		} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
			// An error occurred
			return "", err
		}
		// Short code is already in use, try again
	}

	return "", errors.New("failed to generate a unique short code")
}

// GetBioPageIDForLink retrieves the bio page ID for a given link ID
func (s *BioPageService) GetBioPageIDForLink(ctx context.Context, linkID int) (int, error) {
	// Get the actual bio link model, not just the response
	bioLink, err := s.repo.GetBioLinkByID(ctx, linkID)
	if err != nil {
		return 0, err
	}

	return bioLink.BioPageID, nil
}
