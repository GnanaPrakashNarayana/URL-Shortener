package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Shortener ShortenerConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Address string
}

// ShortenerConfig holds the shortener configuration
type ShortenerConfig struct {
	BaseURL    string
	KeyLength  int
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

	return &Config{
		Server: ServerConfig{
			Address: address,
		},
		Shortener: ShortenerConfig{
			BaseURL:    baseURL,
			KeyLength:  keyLength,
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