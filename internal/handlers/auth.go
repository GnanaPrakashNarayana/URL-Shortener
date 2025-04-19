package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Auth handles authentication requests
type Auth struct {
	authService  *services.AuthService
	templates    *template.Template
	sessionStore *sessions.CookieStore
	sessionName  string
}

// NewAuth creates a new auth handler
func NewAuth(authService *services.AuthService, templatesDir string, sessionStore *sessions.CookieStore, sessionName string) (*Auth, error) {
	// Parse templates
	templates, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &Auth{
		authService:  authService,
		templates:    templates,
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}, nil
}

// RegisterForm handles the registration form
func (h *Auth) RegisterForm(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	if middleware.GetUserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Get the redirect URL
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Render the template
	data := struct {
		Error       string
		RedirectURL string
		CSRFToken   string
	}{
		Error:       r.URL.Query().Get("error"),
		RedirectURL: redirectURL,
		CSRFToken:   csrf.Token(r),
	}

	h.renderTemplate(w, "register.html", data)
}

// Register handles the registration request
func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/auth/register?error=Invalid form", http.StatusSeeOther)
		return
	}

	// Get the form data
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	redirectURL := r.FormValue("redirect")

	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Validate the form data
	if username == "" || email == "" || password == "" {
		http.Redirect(w, r, "/auth/register?error=All fields are required", http.StatusSeeOther)
		return
	}

	if password != passwordConfirm {
		http.Redirect(w, r, "/auth/register?error=Passwords do not match", http.StatusSeeOther)
		return
	}

	// Register the user
	user, err := h.authService.RegisterUser(r.Context(), username, email, password)
	if err != nil {
		if err == services.ErrUserExists {
			http.Redirect(w, r, "/auth/register?error=Username or email already exists", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/auth/register?error=Failed to register user", http.StatusSeeOther)
		return
	}

	// Generate a token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		http.Redirect(w, r, "/auth/login?error=Failed to generate token", http.StatusSeeOther)
		return
	}

	// Store the token in a session
	session, _ := h.sessionStore.Get(r, h.sessionName)
	session.Values["token"] = token
	if err := session.Save(r, w); err != nil {
		http.Redirect(w, r, "/auth/login?error=Failed to save session", http.StatusSeeOther)
		return
	}

	// Redirect to the dashboard
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// LoginForm handles the login form
func (h *Auth) LoginForm(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	if middleware.GetUserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Get the redirect URL
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Render the template
	data := struct {
		Error       string
		RedirectURL string
		CSRFToken   string
		GoogleAuth  bool
		GitHubAuth  bool
	}{
		Error:       r.URL.Query().Get("error"),
		RedirectURL: redirectURL,
		CSRFToken:   csrf.Token(r),
		GoogleAuth:  h.authService.HasProvider(services.ProviderGoogle),
		GitHubAuth:  h.authService.HasProvider(services.ProviderGitHub),
	}

	h.renderTemplate(w, "login.html", data)
}

// Login handles the login request
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/auth/login?error=Invalid form", http.StatusSeeOther)
		return
	}

	// Get the form data
	usernameOrEmail := r.FormValue("username_or_email")
	password := r.FormValue("password")
	redirectURL := r.FormValue("redirect")

	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Validate the form data
	if usernameOrEmail == "" || password == "" {
		http.Redirect(w, r, "/auth/login?error=Username/email and password are required", http.StatusSeeOther)
		return
	}

	// Login the user
	user, err := h.authService.LoginUser(r.Context(), usernameOrEmail, password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			http.Redirect(w, r, "/auth/login?error=Invalid credentials", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/auth/login?error=Failed to login", http.StatusSeeOther)
		return
	}

	// Generate a token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		http.Redirect(w, r, "/auth/login?error=Failed to generate token", http.StatusSeeOther)
		return
	}

	// Store the token in a session
	session, _ := h.sessionStore.Get(r, h.sessionName)
	session.Values["token"] = token
	if err := session.Save(r, w); err != nil {
		http.Redirect(w, r, "/auth/login?error=Failed to save session", http.StatusSeeOther)
		return
	}

	// Redirect to the dashboard
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// Logout handles the logout request
func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session, _ := h.sessionStore.Get(r, h.sessionName)

	// Delete the token
	delete(session.Values, "token")

	// Save the session
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// OAuthLogin handles the OAuth login request
func (h *Auth) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	// Get the provider
	vars := mux.Vars(r)
	providerStr := vars["provider"]
	provider := services.Provider(providerStr)

	// Generate a state
	state, err := generateRandomState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Get the redirect URL
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Store the state and redirect URL in a session
	session, _ := h.sessionStore.Get(r, h.sessionName+"_oauth")
	session.Values["state"] = state
	session.Values["redirect"] = redirectURL
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Get the OAuth URL
	url, err := h.authService.GetOAuthURL(provider, state)
	if err != nil {
		http.Error(w, "Failed to get OAuth URL", http.StatusInternalServerError)
		return
	}

	// Redirect to the OAuth URL
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// OAuthCallback handles the OAuth callback
func (h *Auth) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Get the provider
	vars := mux.Vars(r)
	providerStr := vars["provider"]
	provider := services.Provider(providerStr)

	// Get the code and state
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Get the session
	session, err := h.sessionStore.Get(r, h.sessionName+"_oauth")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	// Get the expected state
	expectedState, ok := session.Values["state"].(string)
	if !ok {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Get the redirect URL
	redirectURL, _ := session.Values["redirect"].(string)
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Delete the session
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	// Handle the OAuth callback
	user, err := h.authService.HandleOAuthCallback(r.Context(), provider, code, state, expectedState)
	if err != nil {
		http.Redirect(w, r, "/auth/login?error=Failed to authenticate with "+providerStr, http.StatusSeeOther)
		return
	}

	// Generate a token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		http.Redirect(w, r, "/auth/login?error=Failed to generate token", http.StatusSeeOther)
		return
	}

	// Store the token in a session
	session, _ = h.sessionStore.Get(r, h.sessionName)
	session.Values["token"] = token
	if err := session.Save(r, w); err != nil {
		http.Redirect(w, r, "/auth/login?error=Failed to save session", http.StatusSeeOther)
		return
	}

	// Redirect to the dashboard
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// generateRandomState generates a random state for OAuth
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// renderTemplate renders a template
func (h *Auth) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// API authentication endpoints

// LoginAPI handles the API login request
func (h *Auth) LoginAPI(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req struct {
		UsernameOrEmail string `json:"username_or_email"`
		Password        string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.UsernameOrEmail == "" || req.Password == "" {
		http.Error(w, "Username/email and password are required", http.StatusBadRequest)
		return
	}

	// Login the user
	user, err := h.authService.LoginUser(r.Context(), req.UsernameOrEmail, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	// Generate a token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return the token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
		"user":  fmt.Sprintf("%s (%s)", user.Username, user.Email),
	})
}
