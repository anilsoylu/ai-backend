package auth

import (
	"ai-backend/internal/database"
	"ai-backend/internal/models"
	"ai-backend/pkg/email"
	"ai-backend/pkg/utils"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required,min=3"` // email veya username
	Password   string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

type UpdatePasswordRequest struct {
	ResetToken   string `json:"reset_token" binding:"required"`
	NewPassword  string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Login handles user login with email or username
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Try to find user by email or username
	result := database.DB.Where("email = ? OR username = ?", req.Identifier, req.Identifier).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if user.Password == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check user status
	if user.Status == models.StatusBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is banned"})
		return
	}

	if user.Status == models.StatusFrozen {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is frozen"})
		return
	}

	if user.Status == models.StatusPassive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is passive. Please contact support to reactivate your account."})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Check if username already exists
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	now := time.Now()
	password := string(hashedPassword)
	user := models.User{
		Username:      &req.Username,
		Email:         &req.Email,
		Password:      &password,
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		EmailVerified: &now,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// RequestPasswordReset handles password reset requests
func RequestPasswordReset(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Güvenlik nedeniyle kullanıcıya spesifik hata dönmüyoruz
		c.JSON(http.StatusOK, ResetPasswordResponse{
			Message: "If your email is registered, you will receive a password reset link",
		})
		return
	}

	// Generate reset token
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}
	resetToken := hex.EncodeToString(token)

	// Set expiration time (1 hour from now)
	expiresAt := time.Now().Add(1 * time.Hour)

	// Save verification token
	verificationToken := models.VerificationToken{
		Identifier: *user.Email,
		Token:      resetToken,
		Expires:    expiresAt,
	}

	if err := database.DB.Create(&verificationToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
		return
	}

	// Send reset email
	if err := email.SendPasswordResetEmail(*user.Email, resetToken); err != nil {
		// Delete the token if email sending fails
		database.DB.Delete(&verificationToken)
		
		// Log the actual error but don't expose it to the user
		log.Printf("Failed to send reset email to %s: %v", *user.Email, err)
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email. Please try again later."})
		return
	}

	log.Printf("Password reset email sent successfully to: %s", *user.Email)
	c.JSON(http.StatusOK, ResetPasswordResponse{
		Message: "Password reset instructions have been sent to your email",
	})
}

// UpdatePassword handles password updates using reset token
func UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find valid token
	var verificationToken models.VerificationToken
	if err := database.DB.Where("token = ? AND expires > ?", req.ResetToken, time.Now()).First(&verificationToken).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Find user
	var user models.User
	if err := database.DB.Where("email = ?", verificationToken.Identifier).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Update password
	password := string(hashedPassword)
	user.Password = &password

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Delete used token
	database.DB.Delete(&verificationToken)

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// ChangePassword handles password changes for authenticated users
func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user := userInterface.(models.User)

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Update password
	password := string(hashedPassword)
	user.Password = &password

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
} 