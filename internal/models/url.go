package models

import (
	"time"
)

// URL represents a shortened URL
type URL struct {
	ID           string     `json:"id"`           // Shortened ID
	OriginalURL  string     `json:"original_url"` // Original URL
	CreatedAt    time.Time  `json:"created_at"`   // Creation time
	Visits       int        `json:"visits"`       // Number of visits
	LastVisitAt  time.Time  `json:"last_visit_at,omitempty"` // Last visit time
	UserID       *int       `json:"user_id,omitempty"`       // ID of the user who created the URL
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`    // Expiration time (nil for never)
}

// URLResponse represents the response to be sent to the client
type URLResponse struct {
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	Visits      int        `json:"visits"`
	UserID      *int       `json:"user_id,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// NewURL creates a new URL
func NewURL(id, originalURL string, userID *int, expiresAt *time.Time) *URL {
	return &URL{
		ID:          id,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		Visits:      0,
		UserID:      userID,
		ExpiresAt:   expiresAt,
	}
}

// IncrementVisits increments the visit count
func (u *URL) IncrementVisits() {
	u.Visits++
	u.LastVisitAt = time.Now()
}

// HasExpired checks if the URL has expired
func (u *URL) HasExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}