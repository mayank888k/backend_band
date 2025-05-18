package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/modernband/booking/internal/database"
	"github.com/modernband/booking/internal/handlers"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Set default port if not specified
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Initialize database connection
	_ = database.GetDB() // Initialize SQLite

	// Set up the router
	router := gin.Default()
	handlers.SetupRoutes(router)

	// Start the server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
