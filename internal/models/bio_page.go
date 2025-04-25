package models

import (
	"time"
)

// BioPage represents a bio link page
type BioPage struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	ShortCode       string     `json:"short_code"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	Theme           string     `json:"theme"`
	ProfileImageURL string     `json:"profile_image_url,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Visits          int        `json:"visits"`
	LastVisitAt     time.Time  `json:"last_visit_at,omitempty"`
	IsPublished     bool       `json:"is_published"`
	CustomCSS       string     `json:"custom_css,omitempty"`
	Links           []*BioLink `json:"links,omitempty"`
}

// BioLink represents a link in a bio page
type BioLink struct {
	ID           int       `json:"id"`
	BioPageID    int       `json:"bio_page_id"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	DisplayOrder int       `json:"display_order"`
	Icon         string    `json:"icon,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Visits       int       `json:"visits"`
	IsEnabled    bool      `json:"is_enabled"`
}

// BioPageThemes defines the available themes
var BioPageThemes = []string{
	"default",
	"dark",
	"light",
	"blue",
	"green",
	"purple",
	"orange",
	"minimal",
}

// NewBioPage creates a new bio page
func NewBioPage(userID int, shortCode, title string) *BioPage {
	now := time.Now()
	return &BioPage{
		UserID:      userID,
		ShortCode:   shortCode,
		Title:       title,
		Theme:       "default",
		CreatedAt:   now,
		UpdatedAt:   now,
		Visits:      0,
		IsPublished: false,
		Links:       make([]*BioLink, 0),
	}
}

// NewBioLink creates a new bio link
func NewBioLink(bioPageID int, title, url string, displayOrder int) *BioLink {
	now := time.Now()
	return &BioLink{
		BioPageID:    bioPageID,
		Title:        title,
		URL:          url,
		DisplayOrder: displayOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
		Visits:       0,
		IsEnabled:    true,
	}
}

// IncrementVisits increments the visit count for a bio page
func (b *BioPage) IncrementVisits() {
	b.Visits++
	b.LastVisitAt = time.Now()
}

// IncrementVisits increments the visit count for a bio link
func (b *BioLink) IncrementVisits() {
	b.Visits++
	b.UpdatedAt = time.Now()
}

// BioPageResponse represents the response format for a bio page
type BioPageResponse struct {
	ID              int              `json:"id"`
	UserID          int              `json:"user_id"`
	ShortCode       string           `json:"short_code"`
	ShortURL        string           `json:"short_url"`
	Title           string           `json:"title"`
	Description     string           `json:"description,omitempty"`
	Theme           string           `json:"theme"`
	ProfileImageURL string           `json:"profile_image_url,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	Visits          int              `json:"visits"`
	IsPublished     bool             `json:"is_published"`
	CustomCSS       string           `json:"custom_css,omitempty"` // Added this field
	Links           []*BioLinkResponse `json:"links,omitempty"`
}

// BioLinkResponse represents the response format for a bio link
type BioLinkResponse struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	DisplayOrder int       `json:"display_order"`
	Icon         string    `json:"icon,omitempty"`
	Visits       int       `json:"visits"`
	IsEnabled    bool      `json:"is_enabled"`
}

// ToBioPageResponse converts a bio page to a response
func (b *BioPage) ToBioPageResponse(baseURL string) *BioPageResponse {
	response := &BioPageResponse{
		ID:              b.ID,
		UserID:          b.UserID,
		ShortCode:       b.ShortCode,
		ShortURL:        baseURL + "/b/" + b.ShortCode,
		Title:           b.Title,
		Description:     b.Description,
		Theme:           b.Theme,
		ProfileImageURL: b.ProfileImageURL,
		CreatedAt:       b.CreatedAt,
		Visits:          b.Visits,
		IsPublished:     b.IsPublished,
		CustomCSS:       b.CustomCSS, // Added this line to copy the CustomCSS field
		Links:           make([]*BioLinkResponse, 0),
	}

	if b.Links != nil {
		for _, link := range b.Links {
			response.Links = append(response.Links, link.ToBioLinkResponse())
		}
	}

	return response
}

// ToBioLinkResponse converts a bio link to a response
func (b *BioLink) ToBioLinkResponse() *BioLinkResponse {
	return &BioLinkResponse{
		ID:           b.ID,
		Title:        b.Title,
		URL:          b.URL,
		DisplayOrder: b.DisplayOrder,
		Icon:         b.Icon,
		Visits:       b.Visits,
		IsEnabled:    b.IsEnabled,
	}
}