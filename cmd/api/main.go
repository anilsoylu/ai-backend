package main

import (
	"log"
	"os"

	"ai-backend/internal/database"
	"ai-backend/internal/middleware"
	"ai-backend/internal/models"
	"ai-backend/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	database.InitDB()

	// Migrate database schemas
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Session{},
		&models.VerificationToken{},
		&models.Question{},
		&models.Answer{},
		&models.Vote{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed default user
	if err := database.SeedDefaultUser(); err != nil {
		log.Fatal("Failed to seed default user:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Add global error handler
	r.Use(middleware.ErrorHandler())

	// Setup routes
	routes.SetupAuthRoutes(r)
	routes.SetupUserRoutes(r)

	// Basic health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 