package models

import (
	"time"
)

// URL represents a shortened URL
type URL struct {
	ID          string    `json:"id"`           // Shortened ID
	OriginalURL string    `json:"original_url"` // Original URL
	CreatedAt   time.Time `json:"created_at"`   // Creation time
	Visits      int       `json:"visits"`       // Number of visits
	LastVisitAt time.Time `json:"last_visit_at,omitempty"` // Last visit time
}

// URLResponse represents the response to be sent to the client
type URLResponse struct {
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	Visits      int       `json:"visits"`
}

// NewURL creates a new URL
func NewURL(id, originalURL string) *URL {
	return &URL{
		ID:          id,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		Visits:      0,
	}
}

// IncrementVisits increments the visit count
func (u *URL) IncrementVisits() {
	u.Visits++
	u.LastVisitAt = time.Now()
}