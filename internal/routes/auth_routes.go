package routes

import (
	"ai-backend/internal/handlers/auth"
	"ai-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures the auth routes
func SetupAuthRoutes(router *gin.Engine) {
	authGroup := router.Group("/api/auth")
	{
		// Public routes
		authGroup.POST("/login", auth.Login)
		authGroup.POST("/register", auth.Register)
		authGroup.POST("/reset-password", auth.RequestPasswordReset)
		authGroup.POST("/update-password", auth.UpdatePassword)

		// Protected routes
		authGroup.Use(middleware.AuthMiddleware())
		authGroup.POST("/change-password", auth.ChangePassword)
	}
} 