package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/app"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start the application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		if err := application.Start(); err != nil {
			log.Fatalf("Failed to start application: %v", err)
		}
	}()

	log.Printf("Server is running at %s", cfg.Server.Address)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := application.Stop(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}