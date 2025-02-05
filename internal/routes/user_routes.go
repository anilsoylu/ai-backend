package routes

import (
	"ai-backend/internal/handlers/user"
	"ai-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures the user routes
func SetupUserRoutes(router *gin.Engine, userHandler *user.UserHandler) {
	userGroup := router.Group("/api/users")
	{
		// Protected routes that require authentication
		userGroup.Use(middleware.AuthMiddleware())
		
		// All authenticated users can access this endpoint
		userGroup.PUT("/status", user.UpdateUserStatus)
		userGroup.PUT("/profile", userHandler.UpdateProfile)
	}
} 