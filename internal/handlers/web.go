package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
    "github.com/GnanaPrakashNarayana/url-shortener/internal/services"
)

// Web handles web requests
type Web struct {
	shortenerService *services.ShortenerService
	templates        *template.Template
}

// NewWeb creates a new web handler
func NewWeb(shortenerService *services.ShortenerService, templatesDir string) (*Web, error) {
	// Parse templates
	templates, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
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

	// Render the template
	data := struct {
		URLs []*models.URLResponse
		Error string
	}{
		URLs:  urls,
		Error: r.URL.Query().Get("error"),
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

	// Get the URL
	url := r.FormValue("url")
	if url == "" {
		http.Redirect(w, r, "/?error=URL is required", http.StatusSeeOther)
		return
	}

	// Shorten the URL
	_, err := h.shortenerService.Shorten(r.Context(), url)
	if err != nil {
		if err == services.ErrInvalidURL {
			http.Redirect(w, r, "/?error=Invalid URL", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/?error=Failed to shorten URL", http.StatusSeeOther)
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renderTemplate renders a template
func (h *Web) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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