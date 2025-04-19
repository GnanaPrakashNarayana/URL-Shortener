package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	Shortener ShortenerConfig
	Database  DatabaseConfig
	Auth      AuthConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Address string
}

// ShortenerConfig holds the shortener configuration
type ShortenerConfig struct {
	BaseURL   string
	KeyLength int
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	// Type is the database type (memory, postgres, mysql, mongodb)
	Type string
	// DSN is the data source name for the database connection
	DSN string
	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns int
	// MaxIdleConns is the maximum number of idle connections in the pool
	MaxIdleConns int
	// ConnMaxLifetime is the maximum amount of time a connection may be reused in seconds
	ConnMaxLifetime int
	// MigrationsPath is the path to the migrations directory
	MigrationsPath string
}

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	// JWT secret key
	JWTSecret string
	// JWT expiration time
	JWTExpirationMinutes int
	// Session cookie name
	SessionCookieName string
	// Session cookie secure flag
	SessionCookieSecure bool
	// Session cookie max age in seconds
	SessionCookieMaxAge int
	// CSRF Token
	CSRFKey string
	// OAuth providers
	OAuth OAuthConfig
}

// OAuthConfig holds the OAuth providers configuration
type OAuthConfig struct {
	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	// GitHub OAuth
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
}

// Load loads the configuration from environment variables or .env file
func Load(configPath string) (*Config, error) {
	// Load .env file if it exists
	if configPath != "" {
		if err := godotenv.Load(configPath); err != nil {
			return nil, err
		}
	} else {
		// Try to load .env file from the current directory
		_ = godotenv.Load()
	}

	// Get server address from environment or use default
	address := getEnv("SERVER_ADDRESS", ":8080")

	// Get base URL from environment or use default
	baseURL := getEnv("BASE_URL", "http://localhost:8080")
	// Remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Get key length from environment or use default
	keyLength := 6

	// Get database type from environment or use default
	dbType := getEnv("DB_TYPE", "memory")

	// Get database DSN from environment or use default
	dbDSN := getEnv("DB_DSN", "")

	// Get database connection pool settings
	maxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "10"))
	maxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME", "300"))
	migrationsPath := getEnv("DB_MIGRATIONS_PATH", "migrations")

	// Auth config
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	jwtExpirationMinutes, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_MINUTES", "1440")) // 24 hours
	sessionCookieName := getEnv("SESSION_COOKIE_NAME", "url_shortener_session")
	sessionCookieSecure, _ := strconv.ParseBool(getEnv("SESSION_COOKIE_SECURE", "false"))
	sessionCookieMaxAge, _ := strconv.Atoi(getEnv("SESSION_COOKIE_MAX_AGE", "86400")) // 24 hours
	csrfKey := getEnv("CSRF_KEY", "32-byte-long-auth-key")

	// OAuth config
	googleClientID := getEnv("GOOGLE_CLIENT_ID", "")
	googleClientSecret := getEnv("GOOGLE_CLIENT_SECRET", "")
	googleRedirectURL := getEnv("GOOGLE_REDIRECT_URL", baseURL+"/auth/google/callback")

	githubClientID := getEnv("GITHUB_CLIENT_ID", "")
	githubClientSecret := getEnv("GITHUB_CLIENT_SECRET", "")
	githubRedirectURL := getEnv("GITHUB_REDIRECT_URL", baseURL+"/auth/github/callback")

	return &Config{
		Server: ServerConfig{
			Address: address,
		},
		Shortener: ShortenerConfig{
			BaseURL:   baseURL,
			KeyLength: keyLength,
		},
		Database: DatabaseConfig{
			Type:            dbType,
			DSN:             dbDSN,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			MigrationsPath:  migrationsPath,
		},
		Auth: AuthConfig{
			JWTSecret:            jwtSecret,
			JWTExpirationMinutes: jwtExpirationMinutes,
			SessionCookieName:    sessionCookieName,
			SessionCookieSecure:  sessionCookieSecure,
			SessionCookieMaxAge:  sessionCookieMaxAge,
			CSRFKey:              csrfKey,
			OAuth: OAuthConfig{
				GoogleClientID:     googleClientID,
				GoogleClientSecret: googleClientSecret,
				GoogleRedirectURL:  googleRedirectURL,
				GitHubClientID:     githubClientID,
				GitHubClientSecret: githubClientSecret,
				GitHubRedirectURL:  githubRedirectURL,
			},
		},
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}