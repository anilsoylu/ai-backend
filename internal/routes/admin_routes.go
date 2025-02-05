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

	// Role management
	adminGroup.PUT("/users/role", admin.UpdateUserRole(db))
	adminGroup.GET("/users/:user_id/role-history", admin.GetUserRoleHistory(db))
	adminGroup.GET("/role-histories", admin.GetAllRoleHistories(db))
} 