package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
)

// ErrNotFound is returned when a URL is not found
var ErrNotFound = errors.New("URL not found")

// MemoryRepository is an in-memory repository
type MemoryRepository struct {
	urls  map[string]*models.URL
	mutex sync.RWMutex
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		urls: make(map[string]*models.URL),
	}
}

// Store stores a URL in the repository
func (r *MemoryRepository) Store(ctx context.Context, url *models.URL) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.urls[url.ID] = url
	return nil
}

// GetByID retrieves a URL by its ID
func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*models.URL, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	url, ok := r.urls[id]
	if !ok {
		return nil, ErrNotFound
	}
	
	return url, nil
}

// Update updates a URL in the repository
func (r *MemoryRepository) Update(ctx context.Context, url *models.URL) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	_, ok := r.urls[url.ID]
	if !ok {
		return ErrNotFound
	}
	
	r.urls[url.ID] = url
	return nil
}

// List lists all URLs in the repository
func (r *MemoryRepository) List(ctx context.Context) ([]*models.URL, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	urls := make([]*models.URL, 0, len(r.urls))
	for _, url := range r.urls {
		urls = append(urls, url)
	}
	
	return urls, nil
}

// ListByUserID lists all URLs for a user
func (r *MemoryRepository) ListByUserID(ctx context.Context, userID int) ([]*models.URL, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	urls := make([]*models.URL, 0)
	for _, url := range r.urls {
		if url.UserID != nil && *url.UserID == userID {
			urls = append(urls, url)
		}
	}
	
	return urls, nil
}

// Close closes the repository
func (r *MemoryRepository) Close() error {
	return nil
}