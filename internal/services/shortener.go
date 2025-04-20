package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Charset for generating short URLs
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Common errors
var (
	ErrInvalidURL      = errors.New("invalid URL")
	ErrInvalidSlug     = errors.New("invalid custom slug: must contain only letters, numbers, hyphens, and underscores")
	ErrSlugUnavailable = errors.New("custom slug is already in use")
	ErrURLExpired      = errors.New("URL has expired")
	ErrInvalidPassword = errors.New("invalid password")
)

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

// Shorten shortens a URL, optionally with a custom slug, expiration time, and password protection
func (s *ShortenerService) Shorten(ctx context.Context, originalURL string, userID *int, customSlug string, expiresIn *time.Duration, password string) (*models.URLResponse, error) {
	// Validate URL
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	var id string
	var err error

	// Use custom slug if provided, otherwise generate a random ID
	if customSlug != "" {
		// Validate custom slug
		if err := validateCustomSlug(customSlug); err != nil {
			return nil, err
		}

		// Check if the slug is available
		_, err := s.repo.GetByID(ctx, customSlug)
		if err == nil {
			// Slug already exists
			return nil, ErrSlugUnavailable
		} else if err != repository.ErrNotFound {
			// Some other error occurred
			return nil, err
		}

		// Slug is available
		id = customSlug
	} else {
		// Generate a random ID
		id, err = s.generateUniqueID(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Calculate expiration time if provided
	var expiresAt *time.Time
	if expiresIn != nil && *expiresIn > 0 {
		t := time.Now().Add(*expiresIn)
		expiresAt = &t
	}

	// Create a new URL
	shortenedURL := models.NewURL(id, originalURL, userID, expiresAt)

	// Hash the password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		shortenedURL.PasswordHash = string(hashedPassword)
	}

	// Store the URL
	if err := s.repo.Store(ctx, shortenedURL); err != nil {
		return nil, err
	}

	// Return the response
	return &models.URLResponse{
		ShortURL:           s.baseURL + "/" + id,
		OriginalURL:        originalURL,
		CreatedAt:          shortenedURL.CreatedAt,
		Visits:             shortenedURL.Visits,
		UserID:             userID,
		ExpiresAt:          expiresAt,
		IsPasswordProtected: shortenedURL.PasswordHash != "",
	}, nil
}

// VerifyPassword checks if the provided password is correct for the URL
func (s *ShortenerService) VerifyPassword(ctx context.Context, id string, password string) (bool, error) {
	url, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	// If no password is set, return true
	if url.PasswordHash == "" {
		return true, nil
	}

	// Check the password
	err = bcrypt.CompareHashAndPassword([]byte(url.PasswordHash), []byte(password))
	if err != nil {
		return false, ErrInvalidPassword
	}

	return true, nil
}

// Get retrieves a URL by its ID
func (s *ShortenerService) Get(ctx context.Context, id string) (*models.URL, error) {
	url, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if URL has expired (should be redundant as repository now checks this)
	if url.HasExpired() {
		return nil, ErrURLExpired
	}

	return url, nil
}

// GetWithoutPassword retrieves a URL by its ID, but doesn't increment the counter or reveal if it's password-protected
func (s *ShortenerService) GetWithoutPassword(ctx context.Context, id string) (*models.URL, error) {
	url, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if URL has expired (should be redundant as repository now checks this)
	if url.HasExpired() {
		return nil, ErrURLExpired
	}

	return url, nil
}

// IncrementVisitCount increments the visit counter for a URL
func (s *ShortenerService) IncrementVisitCount(ctx context.Context, url *models.URL) error {
	url.IncrementVisits()
	
	// Update the URL
	if err := s.repo.Update(ctx, url); err != nil {
		// Log the error but don't fail the request
		return err
	}

	return nil
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
			ShortURL:           s.baseURL + "/" + u.ID,
			OriginalURL:        u.OriginalURL,
			CreatedAt:          u.CreatedAt,
			Visits:             u.Visits,
			UserID:             u.UserID,
			ExpiresAt:          u.ExpiresAt,
			IsPasswordProtected: u.PasswordHash != "",
		})
	}

	return responses, nil
}

// ListByUserID lists all URLs for a user
func (s *ShortenerService) ListByUserID(ctx context.Context, userID int) ([]*models.URLResponse, error) {
	urls, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.URLResponse, 0, len(urls))
	for _, u := range urls {
		responses = append(responses, &models.URLResponse{
			ShortURL:           s.baseURL + "/" + u.ID,
			OriginalURL:        u.OriginalURL,
			CreatedAt:          u.CreatedAt,
			Visits:             u.Visits,
			UserID:             u.UserID,
			ExpiresAt:          u.ExpiresAt,
			IsPasswordProtected: u.PasswordHash != "",
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

// validateCustomSlug validates a custom slug
func validateCustomSlug(slug string) error {
	// Check if slug contains only allowed characters: letters, numbers, hyphens, and underscores
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", slug)
	if err != nil {
		return err
	}

	if !matched {
		return ErrInvalidSlug
	}

	// Check for offensive words or other restrictions
	// This is a simplified check - you might want to expand this
	offensiveWords := []string{"admin", "login", "signup", "api"}
	slugLower := strings.ToLower(slug)
	
	for _, word := range offensiveWords {
		if slugLower == word {
			return errors.New("this custom slug is not allowed")
		}
	}

	return nil
}