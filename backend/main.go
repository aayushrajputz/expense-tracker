package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/your-github/expense-tracker-backend/config"
	"github.com/your-github/expense-tracker-backend/database"
	"github.com/your-github/expense-tracker-backend/routes"
)

func main() {
	_ = godotenv.Load() // ignore if .env absent

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialise DB & auto-migrate
	database.Connect()
	database.Migrate(database.DB)

	// Build router
	r := routes.SetupRouter(database.DB, cfg)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Expense Tracker API listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
