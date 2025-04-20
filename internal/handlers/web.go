package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// Web handles web requests
type Web struct {
	shortenerService *services.ShortenerService
	templates        *template.Template
}

// NewWeb creates a new web handler
func NewWeb(shortenerService *services.ShortenerService, templatesDir string) (*Web, error) {
	// Create a new template with functions
	tmpl := template.New("")

	// Add template functions
	tmpl = tmpl.Funcs(GetTemplateFuncs())

	// Parse templates
	templates, err := tmpl.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &Web{
		shortenerService: shortenerService,
		templates:        templates,
	}, nil
}

// Home handles the home page request
func (h *Web) Home(w http.ResponseWriter, r *http.Request) {
	// Get all URLs
	urls, err := h.shortenerService.List(r.Context())
	if err != nil {
		h.renderError(w, "Failed to list URLs", http.StatusInternalServerError)
		return
	}

	// Get the user from the context (if any)
	user := middleware.GetUserFromContext(r.Context())

	// Render the template
	data := struct {
		URLs      []*models.URLResponse
		Error     string
		User      *models.User
		CSRFToken string
	}{
		URLs:      urls,
		Error:     r.URL.Query().Get("error"),
		User:      user,
		CSRFToken: csrf.Token(r),
	}

	h.renderTemplate(w, "home.html", data)
}

// ShortenURL handles the request to shorten a URL
func (h *Web) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/?error=Invalid form", http.StatusSeeOther)
		return
	}

	// Get the URL and custom slug
	url := r.FormValue("url")
	customSlug := r.FormValue("custom_slug")
	expirationUnit := r.FormValue("expiration_unit")
	expirationValue := r.FormValue("expiration_value")
	password := r.FormValue("password") // Get password value

	if url == "" {
		http.Redirect(w, r, "/?error=URL is required", http.StatusSeeOther)
		return
	}

	// Parse expiration time
	var expiresIn *time.Duration
	if expirationValue != "" && expirationUnit != "" {
		value, err := strconv.ParseInt(expirationValue, 10, 64)
		if err != nil || value <= 0 {
			http.Redirect(w, r, "/?error=Invalid expiration value", http.StatusSeeOther)
			return
		}

		var duration time.Duration
		switch expirationUnit {
		case "minutes":
			duration = time.Duration(value) * time.Minute
		case "hours":
			duration = time.Duration(value) * time.Hour
		case "days":
			duration = time.Duration(value) * 24 * time.Hour
		case "weeks":
			duration = time.Duration(value) * 7 * 24 * time.Hour
		default:
			http.Redirect(w, r, "/?error=Invalid expiration unit", http.StatusSeeOther)
			return
		}

		expiresIn = &duration
	}

	// Get the user from the context (if any)
	user := middleware.GetUserFromContext(r.Context())
	var userID *int
	if user != nil {
		userID = &user.ID
	}

	// Shorten the URL
	_, err := h.shortenerService.Shorten(r.Context(), url, userID, customSlug, expiresIn, password)
	if err != nil {
		switch {
		case err == services.ErrInvalidURL:
			http.Redirect(w, r, "/?error=Invalid URL", http.StatusSeeOther)
		case err == services.ErrInvalidSlug:
			http.Redirect(w, r, "/?error="+err.Error(), http.StatusSeeOther)
		case err == services.ErrSlugUnavailable:
			http.Redirect(w, r, "/?error=Custom slug is already in use", http.StatusSeeOther)
		default:
			http.Redirect(w, r, "/?error=Failed to shorten URL", http.StatusSeeOther)
		}
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PasswordForm handles rendering the password form
func (h *Web) PasswordForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if the URL exists
	url, err := h.shortenerService.GetWithoutPassword(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, services.ErrURLExpired) {
			h.renderError(w, "URL not found or has expired", http.StatusNotFound)
			return
		}
		h.renderError(w, "Failed to get URL", http.StatusInternalServerError)
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

	h.renderTemplate(w, "password.html", data)
}

// VerifyPassword handles the password verification
func (h *Web) VerifyPassword(w http.ResponseWriter, r *http.Request) {
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

// renderTemplate renders a template
func (h *Web) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
}

// renderError renders an error page
func (h *Web) renderError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	data := struct {
		Message string
		Status  int
	}{
		Message: message,
		Status:  status,
	}
	if err := h.templates.ExecuteTemplate(w, "error.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
