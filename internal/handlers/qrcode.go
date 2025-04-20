package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// QRCode handles QR code requests
type QRCode struct {
	qrCodeService *services.QRCodeService
	templates     *template.Template
}

// NewQRCode creates a new QR code handler
func NewQRCode(qrCodeService *services.QRCodeService, templatesDir string) (*QRCode, error) {
	// Create a new template with functions
	tmpl := template.New("")
	
	// Add template functions
	tmpl = tmpl.Funcs(GetTemplateFuncs())
	
	// Parse templates
	templates, err := tmpl.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &QRCode{
		qrCodeService: qrCodeService,
		templates:     templates,
	}, nil
}

// Generate handles the request to generate a QR code
func (h *QRCode) Generate(w http.ResponseWriter, r *http.Request) {
	// Get the URL from the query
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Get the format from the query
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "png"
	}

	// Get the size from the query
	size := 256
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	// Set options
	options := services.DefaultQRCodeOptions()
	options.Size = size

	// Generate QR code
	qrFormat := services.QRCodeFormat(format)
	
	// Set content type based on format
	switch qrFormat {
	case services.QRCodeFormatPNG:
		w.Header().Set("Content-Type", "image/png")
	case services.QRCodeFormatSVG:
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}

	// Set content disposition for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=qrcode.%s", format))

	// Write QR code to response
	err := h.qrCodeService.WriteQRCode(w, url, qrFormat, options)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}
}

// Preview handles the request to preview a QR code
func (h *QRCode) Preview(w http.ResponseWriter, r *http.Request) {
	// Get the URL ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Construct the full URL
	baseURL := r.URL.Query().Get("base_url")
	if baseURL == "" {
		// Use the request host as fallback
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseURL = fmt.Sprintf("%s://%s", scheme, r.Host)
	}
	
	fullURL := baseURL + "/" + id
	
	// Get the format from the query
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "png"
	}
	
	// Generate QR code as base64
	qrFormat := services.QRCodeFormat(format)
	options := services.DefaultQRCodeOptions()
	
	base64QR, err := h.qrCodeService.GenerateBase64(fullURL, qrFormat, options)
	if err != nil {
		http.Error(w, "Failed to generate QR code: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user from context (if authenticated)
	user := middleware.GetUserFromContext(r.Context())
	
	// Render the template
	data := struct {
		ID           string
		URL          string
		QRCodeBase64 string
		Format       string
		User         interface{}
		CSRFToken    string
	}{
		ID:           id,
		URL:          fullURL,
		QRCodeBase64: base64QR,
		Format:       format,
		User:         user,
		CSRFToken:    csrf.Token(r),
	}
	
	h.renderTemplate(w, "qrcode.html", data)
}

// renderTemplate renders a template
func (h *QRCode) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
}