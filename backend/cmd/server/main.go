package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/your-github/expense-tracker-backend/config"
	"github.com/your-github/expense-tracker-backend/internal/app"
	"github.com/your-github/expense-tracker-backend/internal/repo"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := repo.NewDB(cfg.Database.DSN)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := repo.Close(db); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		}
	}()

	// Run migrations
	if err := repo.RunMigrations(cfg.Database.DSN); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize repositories
	repositories := repo.NewRepositories(db)

	// Initialize application
	application := app.NewApp(cfg, repositories, logger)

	// Start the server
	go func() {
		if err := application.Start(); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}
