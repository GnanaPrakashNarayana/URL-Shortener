package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// API handles API requests
type API struct {
	shortenerService *services.ShortenerService
	templates        *template.Template
}

// NewAPI creates a new API handler
func NewAPI(shortenerService *services.ShortenerService, templates *template.Template) *API {
	return &API{
		shortenerService: shortenerService,
		templates:        templates,
	}
}

// ShortenURL handles the request to shorten a URL
func (h *API) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req struct {
		URL        string `json:"url"`
		CustomSlug string `json:"custom_slug,omitempty"`
		ExpiresIn  int64  `json:"expires_in,omitempty"` // Duration in seconds
		Password   string `json:"password,omitempty"`   // Optional password
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

	// Convert expires_in to time.Duration
	var expiresIn *time.Duration
	if req.ExpiresIn > 0 {
		duration := time.Duration(req.ExpiresIn) * time.Second
		expiresIn = &duration
	}

	// Shorten the URL with optional password
	response, err := h.shortenerService.Shorten(r.Context(), req.URL, userID, req.CustomSlug, expiresIn, req.Password)
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

	// Try to get the URL without checking password to see if it exists and if it's password protected
	url, err := h.shortenerService.GetWithoutPassword(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, services.ErrURLExpired) {
			http.Error(w, "URL not found or has expired", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get URL", http.StatusInternalServerError)
		return
	}

	// Check if the URL is password protected
	if url.IsPasswordProtected() {
		// Check if password is in session - simulating a checked password
		session, _ := h.GetPasswordSession(r, id)
		if !session {
			// Redirect to password entry form
			http.Redirect(w, r, "/password/"+id, http.StatusFound)
			return
		}
	}

	// Increment visit count
	if err := h.shortenerService.IncrementVisitCount(r.Context(), url); err != nil {
		// Log error but continue with redirect
		// You might want to implement proper logging here
	}

	// Redirect to the original URL
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

// GetPasswordSession checks if a password session exists for the URL
func (h *API) GetPasswordSession(r *http.Request, urlID string) (bool, error) {
	// In a real application, you would check a session or cookie to see if the password has been verified
	// For simplicity, we'll use a query parameter here
	verified := r.URL.Query().Get("verified")
	return verified == "true", nil
}

// VerifyPassword handles the password verification
func (h *API) VerifyPassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")

	// Verify the password
	isValid, err := h.shortenerService.VerifyPassword(r.Context(), id, password)
	if err != nil || !isValid {
		// Redirect back to password form with error
		http.Redirect(w, r, "/password/"+id+"?error=Invalid password", http.StatusSeeOther)
		return
	}

	// In a real app, you would set a session/cookie to remember that the password was verified
	// For simplicity, use a query parameter
	http.Redirect(w, r, "/"+id+"?verified=true", http.StatusSeeOther)
}

// PasswordForm handles rendering the password form
func (h *API) PasswordForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if the URL exists
	url, err := h.shortenerService.GetWithoutPassword(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, services.ErrURLExpired) {
			http.Error(w, "URL not found or has expired", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get URL", http.StatusInternalServerError)
		return
	}

	// Check if the URL is actually password protected
	if !url.IsPasswordProtected() {
		http.Redirect(w, r, "/"+id, http.StatusFound)
		return
	}

	// Get user from context (if authenticated)
	user := middleware.GetUserFromContext(r.Context())

	// Render the password form
	data := struct {
		ID          string
		Error       string
		User        interface{}
		CSRFToken   string
		RedirectURL string
	}{
		ID:          id,
		Error:       r.URL.Query().Get("error"),
		User:        user,
		CSRFToken:   csrf.Token(r),
		RedirectURL: "/" + id,
	}

	// Render the password template
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "password.html", data); err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
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
