package middleware

import (
	"ai-backend/internal/database"
	"ai-backend/internal/models"
	"ai-backend/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies the JWT token and sets the user in the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the Authorization header has the Bearer scheme
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check user status
		if user.Status == models.StatusBanned {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is banned"})
			c.Abort()
			return
		}

		if user.Status == models.StatusFrozen {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is frozen"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)

		c.Next()
	}
}

// RoleMiddleware checks if the user has the required role
func RoleMiddleware(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
} 