package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/config"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/handlers"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
)

// App represents the application
type App struct {
	config    *config.Config
	repo      repository.Repository
	server    *http.Server
	apiHandler *handlers.API
	webHandler *handlers.Web
}

// New creates a new application
func New(cfg *config.Config) (*App, error) {
	// Create repository
	repo := repository.NewMemoryRepository()

	// Create shortener service
	shortenerService := services.NewShortenerService(
		repo,
		cfg.Shortener.BaseURL,
		cfg.Shortener.KeyLength,
	)

	// Create API handler
	apiHandler := handlers.NewAPI(shortenerService)

	// Create web handler
	webHandler, err := handlers.NewWeb(shortenerService, "templates")
	if err != nil {
		return nil, err
	}

	// Create router
	router := mux.NewRouter()

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/shorten", apiHandler.ShortenURL).Methods(http.MethodPost)
	apiRouter.HandleFunc("/urls", apiHandler.ListURLs).Methods(http.MethodGet)

	// Web routes
	router.HandleFunc("/", webHandler.Home).Methods(http.MethodGet)
	router.HandleFunc("/shorten", webHandler.ShortenURL).Methods(http.MethodPost)
	router.HandleFunc("/{id}", apiHandler.RedirectURL).Methods(http.MethodGet)

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Create server
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	return &App{
		config:    cfg,
		repo:      repo,
		server:    server,
		apiHandler: apiHandler,
		webHandler: webHandler,
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

	return nil
}