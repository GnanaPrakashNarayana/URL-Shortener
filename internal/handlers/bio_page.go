package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// BioPage handles bio page requests
type BioPage struct {
	bioPageService *services.BioPageService
	templates      *template.Template
}

// NewBioPage creates a new bio page handler
func NewBioPage(bioPageService *services.BioPageService, templatesDir string) (*BioPage, error) {
	// Create a new template with functions
	tmpl := template.New("")

	// Add template functions
	tmpl = tmpl.Funcs(GetTemplateFuncs())

	// Parse templates
	templates, err := tmpl.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &BioPage{
		bioPageService: bioPageService,
		templates:      templates,
	}, nil
}

// ListBioPages lists all bio pages for the current user
func (h *BioPage) ListBioPages(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Get all bio pages for the user
	bioPages, err := h.bioPageService.ListBioPagesByUserID(r.Context(), user.ID)
	if err != nil {
		h.renderError(w, "Failed to list bio pages", http.StatusInternalServerError)
		return
	}

	// Render the template
	data := struct {
		User      *models.User
		BioPages  []*models.BioPageResponse
		Error     string
		CSRFToken string
	}{
		User:      user,
		BioPages:  bioPages,
		Error:     r.URL.Query().Get("error"),
		CSRFToken: csrf.Token(r),
	}

	h.renderTemplate(w, "bio_pages_list.html", data)
}

// CreateBioPageForm displays the create bio page form
func (h *BioPage) CreateBioPageForm(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Render the template
	data := struct {
		User      *models.User
		Error     string
		CSRFToken string
		Themes    []string
	}{
		User:      user,
		Error:     r.URL.Query().Get("error"),
		CSRFToken: csrf.Token(r),
		Themes:    models.BioPageThemes,
	}

	h.renderTemplate(w, "create_bio_page.html", data)
}

// CreateBioPage handles the creation of a new bio page
func (h *BioPage) CreateBioPage(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/bio/create?error=Invalid form", http.StatusSeeOther)
		return
	}

	// Get the form values
	title := r.FormValue("title")
	shortCode := r.FormValue("short_code")
	description := r.FormValue("description")

	if title == "" {
		http.Redirect(w, r, "/bio/create?error=Title is required", http.StatusSeeOther)
		return
	}

	// Create the bio page
	bioPage, err := h.bioPageService.CreateBioPage(r.Context(), user.ID, shortCode, title, description)
	if err != nil {
		switch err {
		case services.ErrInvalidSlug:
			http.Redirect(w, r, "/bio/create?error=Invalid short code format", http.StatusSeeOther)
		case services.ErrSlugUnavailable:
			http.Redirect(w, r, "/bio/create?error=Short code is already in use", http.StatusSeeOther)
		default:
			http.Redirect(w, r, "/bio/create?error=Failed to create bio page", http.StatusSeeOther)
		}
		return
	}

	// Redirect to the edit page
	http.Redirect(w, r, "/bio/edit/"+strconv.Itoa(bioPage.ID), http.StatusSeeOther)
}

// EditBioPageForm displays the edit bio page form
func (h *BioPage) EditBioPageForm(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Get the bio page ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.renderError(w, "Invalid bio page ID", http.StatusBadRequest)
		return
	}

	// Get the bio page
	bioPage, err := h.bioPageService.GetBioPage(r.Context(), id)
	if err != nil {
		h.renderError(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the user owns the bio page
	if bioPage.UserID != user.ID {
		h.renderError(w, "You don't have permission to edit this bio page", http.StatusForbidden)
		return
	}

	// Render the template
	data := struct {
		User      *models.User
		BioPage   *models.BioPageResponse
		Error     string
		Success   string
		CSRFToken string
		Themes    []string
	}{
		User:      user,
		BioPage:   bioPage,
		Error:     r.URL.Query().Get("error"),
		Success:   r.URL.Query().Get("success"),
		CSRFToken: csrf.Token(r),
		Themes:    models.BioPageThemes,
	}

	h.renderTemplate(w, "edit_bio_page.html", data)
}

// UpdateBioPage handles updating a bio page
func (h *BioPage) UpdateBioPage(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Get the bio page ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.renderError(w, "Invalid bio page ID", http.StatusBadRequest)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/bio/edit/"+vars["id"]+"?error=Invalid form", http.StatusSeeOther)
		return
	}

	// Get the form values
	title := r.FormValue("title")
	description := r.FormValue("description")
	theme := r.FormValue("theme")
	profileImageURL := r.FormValue("profile_image_url")
	isPublishedStr := r.FormValue("is_published")
	customCSS := r.FormValue("custom_css")

	if title == "" {
		http.Redirect(w, r, "/bio/edit/"+vars["id"]+"?error=Title is required", http.StatusSeeOther)
		return
	}

	// Parse boolean values
	isPublished := isPublishedStr == "on" || isPublishedStr == "true"

	// Get the current bio page to check ownership
	currentBioPage, err := h.bioPageService.GetBioPage(r.Context(), id)
	if err != nil {
		h.renderError(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the user owns the bio page
	if currentBioPage.UserID != user.ID {
		h.renderError(w, "You don't have permission to edit this bio page", http.StatusForbidden)
		return
	}

	// Update the bio page
	_, err = h.bioPageService.UpdateBioPage(r.Context(), id, title, description, theme, profileImageURL, isPublished, customCSS)
	if err != nil {
		http.Redirect(w, r, "/bio/edit/"+vars["id"]+"?error=Failed to update bio page", http.StatusSeeOther)
		return
	}

	// Redirect back to the edit page with success message
	http.Redirect(w, r, "/bio/edit/"+vars["id"]+"?success=Bio page updated successfully", http.StatusSeeOther)
}

// DeleteBioPage handles deleting a bio page
func (h *BioPage) DeleteBioPage(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Get the bio page ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.renderError(w, "Invalid bio page ID", http.StatusBadRequest)
		return
	}

	// Get the current bio page to check ownership
	currentBioPage, err := h.bioPageService.GetBioPage(r.Context(), id)
	if err != nil {
		h.renderError(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the user owns the bio page
	if currentBioPage.UserID != user.ID {
		h.renderError(w, "You don't have permission to delete this bio page", http.StatusForbidden)
		return
	}

	// Delete the bio page
	err = h.bioPageService.DeleteBioPage(r.Context(), id)
	if err != nil {
		http.Redirect(w, r, "/bio/pages?error=Failed to delete bio page", http.StatusSeeOther)
		return
	}

	// Redirect to the bio pages list
	http.Redirect(w, r, "/bio/pages?success=Bio page deleted successfully", http.StatusSeeOther)
}

// AddBioLink handles adding a new link to a bio page
func (h *BioPage) AddBioLink(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the bio page ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid bio page ID", http.StatusBadRequest)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	// Get the form values
	title := r.FormValue("title")
	url := r.FormValue("url")

	if title == "" || url == "" {
		http.Error(w, "Title and URL are required", http.StatusBadRequest)
		return
	}

	// Get the current bio page to check ownership
	currentBioPage, err := h.bioPageService.GetBioPage(r.Context(), id)
	if err != nil {
		http.Error(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the user owns the bio page
	if currentBioPage.UserID != user.ID {
		http.Error(w, "You don't have permission to edit this bio page", http.StatusForbidden)
		return
	}

	// Add the bio link
	_, err = h.bioPageService.AddBioLink(r.Context(), id, title, url)
	if err != nil {
		http.Error(w, "Failed to add link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the edit page
	http.Redirect(w, r, "/bio/edit/"+vars["id"], http.StatusSeeOther)
}

// UpdateBioLink handles updating a bio link
func (h *BioPage) UpdateBioLink(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the bio link ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid bio link ID", http.StatusBadRequest)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	// Get the form values
	title := r.FormValue("title")
	url := r.FormValue("url")
	isEnabledStr := r.FormValue("is_enabled")

	if title == "" || url == "" {
		http.Error(w, "Title and URL are required", http.StatusBadRequest)
		return
	}

	// Parse boolean values
	isEnabled := isEnabledStr == "on" || isEnabledStr == "true"

	// Get the bio link to check ownership
	_, err = h.bioPageService.GetBioLinkByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Bio link not found", http.StatusNotFound)
		return
	}

	// Get the bio page ID for this link
	bioPageID, err := h.bioPageService.GetBioPageIDForLink(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get bio page", http.StatusInternalServerError)
		return
	}

	// Get the bio page to check ownership
	bioPage, err := h.bioPageService.GetBioPage(r.Context(), bioPageID)
	if err != nil {
		http.Error(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the user owns the bio page
	if bioPage.UserID != user.ID {
		http.Error(w, "You don't have permission to edit this bio link", http.StatusForbidden)
		return
	}

	// Update the bio link
	_, err = h.bioPageService.UpdateBioLink(r.Context(), id, title, url, isEnabled)
	if err != nil {
		http.Error(w, "Failed to update link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the edit page
	http.Redirect(w, r, "/bio/edit/"+strconv.Itoa(bioPage.ID), http.StatusSeeOther)
}

// DeleteBioLink handles deleting a bio link
func (h *BioPage) DeleteBioLink(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the bio link ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid bio link ID", http.StatusBadRequest)
		return
	}

	// Get the bio link to check ownership
	_, err = h.bioPageService.GetBioLinkByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Bio link not found", http.StatusNotFound)
		return
	}

	// Get the bio page ID for this link
	bioPageID, err := h.bioPageService.GetBioPageIDForLink(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get bio page", http.StatusInternalServerError)
		return
	}

	// Get the bio page to check ownership
	bioPage, err := h.bioPageService.GetBioPage(r.Context(), bioPageID)
	if err != nil {
		http.Error(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the user owns the bio page
	if bioPage.UserID != user.ID {
		http.Error(w, "You don't have permission to delete this bio link", http.StatusForbidden)
		return
	}

	// Delete the bio link
	err = h.bioPageService.DeleteBioLink(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete link", http.StatusInternalServerError)
		return
	}

	// Redirect back to the edit page
	http.Redirect(w, r, "/bio/edit/"+strconv.Itoa(bioPage.ID), http.StatusSeeOther)
}

// ReorderBioLinks handles reordering bio links
func (h *BioPage) ReorderBioLinks(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the bio page ID from the URL
	vars := mux.Vars(r)
	bioPageID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid bio page ID", http.StatusBadRequest)
		return
	}

	// Parse the request body
	var request struct {
		LinkIDs []int `json:"link_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the current bio page to check ownership
	currentBioPage, err := h.bioPageService.GetBioPage(r.Context(), bioPageID)
	if err != nil {
		http.Error(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the user owns the bio page
	if currentBioPage.UserID != user.ID {
		http.Error(w, "You don't have permission to edit this bio page", http.StatusForbidden)
		return
	}

	// Reorder the bio links
	err = h.bioPageService.ReorderBioLinks(r.Context(), bioPageID, request.LinkIDs)
	if err != nil {
		http.Error(w, "Failed to reorder links", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// ViewBioPage displays the public bio page
func (h *BioPage) ViewBioPage(w http.ResponseWriter, r *http.Request) {
	// Get the bio page short code from the URL
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	// Get the bio page
	bioPage, err := h.bioPageService.GetBioPageByShortCode(r.Context(), shortCode)
	if err != nil {
		h.renderError(w, "Bio page not found", http.StatusNotFound)
		return
	}

	// Check if the bio page is published
	if !bioPage.IsPublished {
		h.renderError(w, "This bio page is not published", http.StatusNotFound)
		return
	}

	// Increment the visit count
	err = h.bioPageService.IncrementBioPageVisits(r.Context(), bioPage.ID)
	if err != nil {
		// Log the error but continue with the request
	}

	// Get the user from the context (if any)
	user := middleware.GetUserFromContext(r.Context())

	// Check if the user owns the bio page
	isOwner := user != nil && user.ID == bioPage.UserID

	// Render the template
	data := struct {
		BioPage   *models.BioPageResponse
		User      *models.User
		IsOwner   bool
		CSRFToken string
	}{
		BioPage:   bioPage,
		User:      user,
		IsOwner:   isOwner,
		CSRFToken: csrf.Token(r),
	}

	h.renderTemplate(w, "bio_page.html", data)
}

// RedirectBioLink handles redirecting to a bio link
func (h *BioPage) RedirectBioLink(w http.ResponseWriter, r *http.Request) {
	// Get the bio link ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.renderError(w, "Invalid bio link ID", http.StatusBadRequest)
		return
	}

	// Get the bio link
	bioLink, err := h.bioPageService.GetBioLinkByID(r.Context(), id)
	if err != nil {
		h.renderError(w, "Bio link not found", http.StatusNotFound)
		return
	}

	// Check if the bio link is enabled
	if !bioLink.IsEnabled {
		h.renderError(w, "This link is disabled", http.StatusNotFound)
		return
	}

	// Increment the visit count
	err = h.bioPageService.IncrementBioLinkVisits(r.Context(), id)
	if err != nil {
		// Log the error but continue with the request
	}

	// Redirect to the URL
	http.Redirect(w, r, bioLink.URL, http.StatusFound)
}

// renderTemplate renders a template
func (h *BioPage) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
}

// renderError renders an error page
func (h *BioPage) renderError(w http.ResponseWriter, message string, status int) {
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
