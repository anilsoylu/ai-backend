package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ai-backend/internal/handlers/admin"
	"ai-backend/internal/middleware"
	"ai-backend/internal/models"
)

func SetupAdminRoutes(router *gin.Engine, db *gorm.DB) {
	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.AdminRoleMiddleware([]models.UserRole{models.RoleAdmin, models.RoleSuperAdmin}))

	adminGroup.PUT("/users/role", admin.UpdateUserRole(db))
} 