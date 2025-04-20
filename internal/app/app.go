package app

import (
	"context"
	"crypto/rand"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/config"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/database"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/handlers"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/middleware"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// App represents the application
type App struct {
	config         *config.Config
	repo           repository.Repository
	userRepo       repository.UserRepository
	server         *http.Server
	apiHandler     *handlers.API
	webHandler     *handlers.Web
	authHandler    *handlers.Auth
	dashHandler    *handlers.Dashboard
	qrCodeHandler  *handlers.QRCode
	dbManager      *database.Manager
	authMiddleware *middleware.AuthMiddleware
	sessionStore   *sessions.CookieStore
	qrCodeService  *services.QRCodeService
}

// New creates a new application
func New(cfg *config.Config) (*App, error) {
	var repo repository.Repository
	var userRepo repository.UserRepository
	var dbManager *database.Manager
	var err error

	// Initialize the repository based on the configuration
	if cfg.Database.Type == "postgres" {
		// Create database manager
		dbManager, err = database.NewManager(&cfg.Database)
		if err != nil {
			return nil, err
		}

		// Connect to the database
		db, err := dbManager.Connect()
		if err != nil {
			return nil, err
		}

		// Run migrations
		if err := dbManager.Migrate(); err != nil {
			return nil, err
		}

		// Create PostgreSQL repository
		repo, err = repository.NewPostgresRepository(db)
		if err != nil {
			return nil, err
		}

		// Create PostgreSQL user repository
		userRepo, err = repository.NewPostgresUserRepository(db)
		if err != nil {
			return nil, err
		}
	} else {
		// Fall back to memory repository
		repo = repository.NewMemoryRepository()
		userRepo = repository.NewMemoryUserRepository()
	}

	// Create session store
	// Generate a random key for the session store
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, err
	}
	sessionStore := sessions.NewCookieStore(sessionKey)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.Auth.SessionCookieMaxAge,
		HttpOnly: true,
		Secure:   cfg.Auth.SessionCookieSecure,
	}

	// Create services
	shortenerService := services.NewShortenerService(
		repo,
		cfg.Shortener.BaseURL,
		cfg.Shortener.KeyLength,
	)

	authService := services.NewAuthService(userRepo, &cfg.Auth)

	// Create QR code service
	qrCodeService := services.NewQRCodeService()

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, sessionStore, cfg.Auth.SessionCookieName)

	// Create API handler
	apiTemplates, err := template.New("").Funcs(handlers.GetTemplateFuncs()).ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		return nil, err
	}
	apiHandler := handlers.NewAPI(shortenerService, apiTemplates)

	// Create web handler
	webHandler, err := handlers.NewWeb(shortenerService, "templates")
	if err != nil {
		return nil, err
	}

	// Create auth handler
	authHandler, err := handlers.NewAuth(authService, "templates", sessionStore, cfg.Auth.SessionCookieName)
	if err != nil {
		return nil, err
	}

	// Create dashboard handler
	dashHandler, err := handlers.NewDashboard(shortenerService, "templates")
	if err != nil {
		return nil, err
	}

	// Create QR code handler
	qrCodeHandler, err := handlers.NewQRCode(qrCodeService, "templates")
	if err != nil {
		return nil, err
	}

	// Create router
	router := mux.NewRouter()

	// Add auth middleware to all routes
	router.Use(authMiddleware.Auth)

	// CSRF protection - UPDATED CONFIG
	csrfMiddleware := csrf.Protect(
		[]byte(cfg.Auth.CSRFKey),
		csrf.Secure(cfg.Auth.SessionCookieSecure), // Use the same setting as session cookie
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "CSRF validation failed: "+csrf.FailureReason(r).Error(), http.StatusForbidden)
		})),
		csrf.TrustedOrigins([]string{"localhost:8080", "127.0.0.1:8080"}),
	)
	router.Use(csrfMiddleware)

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/shorten", apiHandler.ShortenURL).Methods(http.MethodPost)
	apiRouter.HandleFunc("/urls", apiHandler.ListURLs).Methods(http.MethodGet)
	apiRouter.HandleFunc("/auth/login", authHandler.LoginAPI).Methods(http.MethodPost)

	// Auth routes
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.RegisterForm).Methods(http.MethodGet)
	authRouter.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	authRouter.HandleFunc("/login", authHandler.LoginForm).Methods(http.MethodGet)
	authRouter.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
	authRouter.HandleFunc("/logout", authHandler.Logout).Methods(http.MethodGet)
	authRouter.HandleFunc("/oauth/{provider}", authHandler.OAuthLogin).Methods(http.MethodGet)
	authRouter.HandleFunc("/oauth/{provider}/callback", authHandler.OAuthCallback).Methods(http.MethodGet)

	// Dashboard routes
	dashRouter := router.PathPrefix("/dashboard").Subrouter()
	dashRouter.Use(authMiddleware.RequireAuth)
	dashRouter.HandleFunc("", dashHandler.Home).Methods(http.MethodGet)
	dashRouter.HandleFunc("/", dashHandler.Home).Methods(http.MethodGet)
	dashRouter.HandleFunc("/shorten", dashHandler.ShortenURL).Methods(http.MethodPost)

	// QR Code routes
	router.HandleFunc("/qrcode/generate", qrCodeHandler.Generate).Methods(http.MethodGet)
	router.HandleFunc("/qrcode/preview/{id}", qrCodeHandler.Preview).Methods(http.MethodGet)

	// Admin routes (example - not implemented yet)
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware.RequireAdmin)
	// adminRouter.HandleFunc("", adminHandler.Home).Methods(http.MethodGet)

	// Web routes
	router.HandleFunc("/", webHandler.Home).Methods(http.MethodGet)
	router.HandleFunc("/shorten", webHandler.ShortenURL).Methods(http.MethodPost)
	router.HandleFunc("/{id}", apiHandler.RedirectURL).Methods(http.MethodGet)

	// Password verification routes
	router.HandleFunc("/password/{id}", apiHandler.PasswordForm).Methods(http.MethodGet)
	router.HandleFunc("/verify-password/{id}", apiHandler.VerifyPassword).Methods(http.MethodPost)

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Create server
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	return &App{
		config:         cfg,
		repo:           repo,
		userRepo:       userRepo,
		server:         server,
		apiHandler:     apiHandler,
		webHandler:     webHandler,
		authHandler:    authHandler,
		dashHandler:    dashHandler,
		qrCodeHandler:  qrCodeHandler,
		dbManager:      dbManager,
		authMiddleware: authMiddleware,
		sessionStore:   sessionStore,
		qrCodeService:  qrCodeService,
	}, nil
}

// Start starts the application
func (a *App) Start() error {
	return a.server.ListenAndServe()
}

// Stop stops the application
func (a *App) Stop() error {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	// Close the repository
	if err := a.repo.Close(); err != nil {
		return err
	}

	// Close the database manager if it exists
	if a.dbManager != nil {
		if err := a.dbManager.Close(); err != nil {
			return err
		}
	}

	return nil
}