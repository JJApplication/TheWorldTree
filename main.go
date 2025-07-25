package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"twt/config"
	"twt/models"
	"twt/server"
	"twt/services"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "config.toml", "Path to configuration file")
	flag.Parse()

	log.SetPrefix("[TwT] ")

	// Load configuration
	if err := config.LoadConfig(*configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := config.GetConfig()
	log.Printf("Configuration loaded successfully")

	// Initialize database
	db, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize GitHub service
	githubService := services.NewGitHubService(cfg.Github.Token)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start servers based on configuration
	if cfg.Server.HTTP.Enable {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.StartHTTPServer(db, githubService); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}()
	}
	if cfg.Server.GRPC.Enable {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.StartHTTPServer(db, githubService); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}()
	}
	if !cfg.Server.HTTP.Enable && !cfg.Server.GRPC.Enable {
		log.Fatalf("All server type is disabled")
	}

	log.Printf("Server(s) started successfully")
	log.Printf("HTTP server: //%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)
	log.Printf("gRPC server: %s", cfg.Server.GRPC.Address)
	log.Printf("Press Ctrl+C to shutdown")

	// Wait for shutdown signal
	<-quit
	log.Println("Shutting down servers...")

	// Note: In a production environment, you would implement graceful shutdown
	// for both HTTP and gRPC servers here

	wg.Wait()
	log.Println("Servers stopped")
}

func initializeDatabase() (*models.DB, error) {
	cfg := config.GetConfig()
	dbPath := cfg.Database.Path

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := models.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	log.Printf("Database initialized at: %s", dbPath)
	return db, nil
}
