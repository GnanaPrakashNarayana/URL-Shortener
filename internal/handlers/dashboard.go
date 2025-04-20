package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/csrf"
)

// Dashboard handles dashboard requests
type Dashboard struct {
	shortenerService *services.ShortenerService
	templates        *template.Template
}

// NewDashboard creates a new dashboard handler
func NewDashboard(shortenerService *services.ShortenerService, templatesDir string) (*Dashboard, error) {
	// Parse templates with custom functions
	templates, err := template.New("").Funcs(GetTemplateFuncs()).ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		shortenerService: shortenerService,
		templates:        templates,
	}, nil
}

// Home handles the dashboard home page
func (h *Dashboard) Home(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Get the user's URLs
	urls, err := h.shortenerService.ListByUserID(r.Context(), user.ID)
	if err != nil {
		h.renderError(w, "Failed to list URLs", http.StatusInternalServerError)
		return
	}

	// Render the template
	data := struct {
		User      *models.User
		URLs      []*models.URLResponse
		Error     string
		CSRFToken string
	}{
		User:      user,
		URLs:      urls,
		Error:     r.URL.Query().Get("error"),
		CSRFToken: csrf.Token(r),
	}

	h.renderTemplate(w, "dashboard.html", data)
}

// ShortenURL handles the request to shorten a URL from the dashboard
func (h *Dashboard) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid form", http.StatusSeeOther)
		return
	}

	// Get the URL and custom slug
	url := r.FormValue("url")
	customSlug := r.FormValue("custom_slug")

	if url == "" {
		http.Redirect(w, r, "/dashboard?error=URL is required", http.StatusSeeOther)
		return
	}

	// Shorten the URL
	_, err := h.shortenerService.Shorten(r.Context(), url, &user.ID, customSlug)
	if err != nil {
		switch {
		case err == services.ErrInvalidURL:
			http.Redirect(w, r, "/dashboard?error=Invalid URL", http.StatusSeeOther)
		case err == services.ErrInvalidSlug:
			http.Redirect(w, r, "/dashboard?error="+err.Error(), http.StatusSeeOther)
		case err == services.ErrSlugUnavailable:
			http.Redirect(w, r, "/dashboard?error=Custom slug is already in use", http.StatusSeeOther)
		default:
			http.Redirect(w, r, "/dashboard?error=Failed to shorten URL", http.StatusSeeOther)
		}
		return
	}

	// Redirect to the dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// renderTemplate renders a template
func (h *Dashboard) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// renderError renders an error page
func (h *Dashboard) renderError(w http.ResponseWriter, message string, status int) {
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