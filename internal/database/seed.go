package database

import (
	"ai-backend/internal/models"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SeedDefaultUser creates a default admin user if it doesn't exist
func SeedDefaultUser() error {
	var user models.User

	// Check if user already exists
	result := DB.Where("email = ?", os.Getenv("DEFAULT_EMAIL")).First(&user)
	if result.Error == nil {
		log.Println("Default user already exists")
		return nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("DEFAULT_PASSWORD")), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Get current time for email verification
	now := time.Now()
	username := os.Getenv("DEFAULT_USERNAME")
	email := os.Getenv("DEFAULT_EMAIL")
	role := models.UserRole(os.Getenv("DEFAULT_ROLES"))
	password := string(hashedPassword)

	// Create default user
	defaultUser := models.User{
		Username:      &username,
		Email:         &email,
		Password:      &password,
		Role:          role,
		Status:        models.StatusActive,
		EmailVerified: &now,
	}

	// Save to database
	result = DB.Create(&defaultUser)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Default user created with email: %s and role: %s", email, role)
	return nil
} 