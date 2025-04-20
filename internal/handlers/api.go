package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

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
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Shorten the URL
	response, err := h.shortenerService.Shorten(r.Context(), req.URL, nil)
	if err != nil {
		if errors.Is(err, services.ErrInvalidURL) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to shorten URL", http.StatusInternalServerError)
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