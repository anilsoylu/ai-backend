package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-backend/internal/models"
)

// AdminRoleMiddleware checks if the user has one of the required admin roles
func AdminRoleMiddleware(allowedRoles []models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by AuthMiddleware)
		user, exists := c.Get("user")
		if !exists {
			log.Print("User not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok {
			log.Printf("Failed to cast user from context. Type: %T", user)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range allowedRoles {
			if u.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			log.Printf("Insufficient permissions. User Role: %s, Required Roles: %v", u.Role, allowedRoles)
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		log.Printf("Access granted for user ID: %d with role: %s", u.ID, u.Role)
		c.Next()
	}
} 