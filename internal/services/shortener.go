package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
)

const (
	// Charset for generating short URLs
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// ErrInvalidURL is returned when the provided URL is invalid
var ErrInvalidURL = errors.New("invalid URL")

// ShortenerService is responsible for shortening URLs
type ShortenerService struct {
	repo      repository.Repository
	baseURL   string
	keyLength int
}

// NewShortenerService creates a new shortener service
func NewShortenerService(repo repository.Repository, baseURL string, keyLength int) *ShortenerService {
	return &ShortenerService{
		repo:      repo,
		baseURL:   baseURL,
		keyLength: keyLength,
	}
}

// Shorten shortens a URL
func (s *ShortenerService) Shorten(ctx context.Context, originalURL string) (*models.URLResponse, error) {
	// Validate URL
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	// Generate a unique ID
	id, err := s.generateUniqueID(ctx)
	if err != nil {
		return nil, err
	}

	// Create a new URL
	shortenedURL := models.NewURL(id, originalURL)

	// Store the URL
	if err := s.repo.Store(ctx, shortenedURL); err != nil {
		return nil, err
	}

	// Return the response
	return &models.URLResponse{
		ShortURL:    s.baseURL + "/" + id,
		OriginalURL: originalURL,
		CreatedAt:   shortenedURL.CreatedAt,
		Visits:      shortenedURL.Visits,
	}, nil
}

// Get retrieves a URL by its ID
func (s *ShortenerService) Get(ctx context.Context, id string) (*models.URL, error) {
	url, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Increment visit count
	url.IncrementVisits()
	
	// Update the URL
	if err := s.repo.Update(ctx, url); err != nil {
		// Log the error but don't fail the request
		return url, nil
	}

	return url, nil
}

// List lists all URLs
func (s *ShortenerService) List(ctx context.Context) ([]*models.URLResponse, error) {
	urls, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.URLResponse, 0, len(urls))
	for _, u := range urls {
		responses = append(responses, &models.URLResponse{
			ShortURL:    s.baseURL + "/" + u.ID,
			OriginalURL: u.OriginalURL,
			CreatedAt:   u.CreatedAt,
			Visits:      u.Visits,
		})
	}

	return responses, nil
}

// generateUniqueID generates a unique ID for a URL
func (s *ShortenerService) generateUniqueID(ctx context.Context) (string, error) {
	for {
		id, err := generateRandomString(s.keyLength)
		if err != nil {
			return "", err
		}

		// Check if the ID already exists
		_, err = s.repo.GetByID(ctx, id)
		if err == repository.ErrNotFound {
			// ID is unique
			return id, nil
		} else if err != nil {
			// An error occurred
			return "", err
		}

		// ID already exists, try again
	}
}

// generateRandomString generates a random string of the given length
func generateRandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

// validateURL validates a URL
func validateURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}

	if parsed.Host == "" {
		return ErrInvalidURL
	}

	return nil
}