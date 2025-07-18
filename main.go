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
	serverType := flag.String("server", "both", "Server type: http, grpc, or both")
	flag.Parse()

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
	switch *serverType {
	case "http":
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.StartHTTPServer(db, githubService); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}()
	case "grpc":
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.StartGRPCServer(db, githubService); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	case "both":
		// Start HTTP server
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.StartHTTPServer(db, githubService); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}()

		// Start gRPC server
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.StartGRPCServer(db, githubService); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	default:
		log.Fatalf("Invalid server type: %s. Use 'http', 'grpc', or 'both'", *serverType)
	}

	log.Printf("Server(s) started successfully")
	log.Printf("HTTP server: http://localhost:%d", cfg.Server.HTTPPort)
	log.Printf("gRPC server: localhost:%d", cfg.Server.GRPCPort)
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
