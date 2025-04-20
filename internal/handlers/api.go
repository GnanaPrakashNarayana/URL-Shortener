package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/mux"
)

// API handles API requests
type API struct {
	shortenerService *services.ShortenerService
}

// NewAPI creates a new API handler
func NewAPI(shortenerService *services.ShortenerService) *API {
	return &API{
		shortenerService: shortenerService,
	}
}

// ShortenURL handles the request to shorten a URL
func (h *API) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req struct {
		URL        string `json:"url"`
		CustomSlug string `json:"custom_slug,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from context (if authenticated)
	user := middleware.GetUserFromContext(r.Context())
	var userID *int
	if user != nil {
		userID = &user.ID
	}

	// Shorten the URL
	response, err := h.shortenerService.Shorten(r.Context(), req.URL, userID, req.CustomSlug)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidURL):
			http.Error(w, "Invalid URL", http.StatusBadRequest)
		case errors.Is(err, services.ErrInvalidSlug):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, services.ErrSlugUnavailable):
			http.Error(w, "Custom slug is already in use", http.StatusConflict)
		default:
			http.Error(w, "Failed to shorten URL", http.StatusInternalServerError)
		}
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RedirectURL handles the request to redirect to the original URL
func (h *API) RedirectURL(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the URL
	url, err := h.shortenerService.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get URL", http.StatusInternalServerError)
		return
	}

	// Redirect to the original URL
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

// ListURLs handles the request to list all URLs
func (h *API) ListURLs(w http.ResponseWriter, r *http.Request) {
	// List all URLs
	urls, err := h.shortenerService.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to list URLs", http.StatusInternalServerError)
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}
