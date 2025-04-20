package services

import (
	"context"
	"testing"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/repository"
)

func TestShortenerService_Shorten(t *testing.T) {
	// Create a repository
	repo := repository.NewMemoryRepository()

	// Create a shortener service
	service := NewShortenerService(repo, "http://localhost:8080", 6)

	// Test shortening a valid URL
	ctx := context.Background()
	resp, err := service.Shorten(ctx, "https://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}

	if resp.OriginalURL != "https://example.com" {
		t.Errorf("Expected original URL to be https://example.com, got %s", resp.OriginalURL)
	}

	if resp.ShortURL == "" {
		t.Error("Expected short URL to be non-empty")
	}

	// Test shortening an invalid URL
	_, err = service.Shorten(ctx, "invalid-url", nil)
	if err != ErrInvalidURL {
		t.Errorf("Expected ErrInvalidURL, got %v", err)
	}
}

func TestShortenerService_Get(t *testing.T) {
	// Create a repository
	repo := repository.NewMemoryRepository()

	// Create a shortener service
	service := NewShortenerService(repo, "http://localhost:8080", 6)

	// Shorten a URL
	ctx := context.Background()
	resp, err := service.Shorten(ctx, "https://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}

	// Extract the ID from the short URL
	id := resp.ShortURL[len("http://localhost:8080/"):]

	// Get the URL
	url, err := service.Get(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}

	if url.OriginalURL != "https://example.com" {
		t.Errorf("Expected original URL to be https://example.com, got %s", url.OriginalURL)
	}

	// Check that the visit count was incremented
	if url.Visits != 1 {
		t.Errorf("Expected visit count to be 1, got %d", url.Visits)
	}
}