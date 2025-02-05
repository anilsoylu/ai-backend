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
		
		// All authenticated users can access these endpoints
		userGroup.PUT("/status", user.UpdateUserStatus)
		userGroup.PUT("/profile", userHandler.UpdateProfile)
		userGroup.DELETE("/account", userHandler.DeleteAccount)
		userGroup.POST("/freeze", userHandler.FreezeAccount)
		userGroup.GET("/freeze/history", userHandler.GetFreezeHistory)
	}
} 