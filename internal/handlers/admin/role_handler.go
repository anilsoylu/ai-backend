package admin

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ai-backend/internal/models"
)

type UpdateRoleRequest struct {
	UserID uint           `json:"user_id" binding:"required"`
	Role   models.UserRole `json:"role" binding:"required,oneof=USER EDITOR ADMIN SUPER_ADMIN"`
	Reason string         `json:"reason"`
}

func UpdateUserRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err)})
			return
		}

		// Get current user from context (set by auth middleware)
		currentUser, exists := c.Get("user")
		if !exists {
			log.Print("User not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		cu, ok := currentUser.(*models.User)
		if !ok {
			log.Print("Failed to cast user from context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Get target user
		var targetUser models.User
		if err := db.First(&targetUser, req.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("User not found with ID: %d", req.UserID)
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			log.Printf("Database error while fetching user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Check if current user is first SUPER_ADMIN
		var firstSuperAdmin models.User
		if err := db.Where("role = ?", models.RoleSuperAdmin).Order("created_at asc").First(&firstSuperAdmin).Error; err != nil {
			log.Printf("Error finding first super admin: %v", err)
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
		}

		// Validation rules
		if targetUser.ID == firstSuperAdmin.ID && targetUser.Role == models.RoleSuperAdmin {
			log.Printf("Attempt to change first SUPER_ADMIN's role. Target ID: %d", targetUser.ID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot change first SUPER_ADMIN's role"})
			return
		}

		if cu.Role == models.RoleAdmin {
			// Admin can only modify between USER and EDITOR roles
			if req.Role != models.RoleUser && req.Role != models.RoleEditor {
				log.Printf("Admin attempted to assign invalid role: %s", req.Role)
				c.JSON(http.StatusForbidden, gin.H{"error": "Admin can only assign USER or EDITOR roles"})
				return
			}

			// Admin cannot modify SUPER_ADMIN or other ADMIN roles
			if targetUser.Role == models.RoleSuperAdmin || targetUser.Role == models.RoleAdmin {
				log.Printf("Admin attempted to modify ADMIN/SUPER_ADMIN role. Target role: %s", targetUser.Role)
				c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify ADMIN or SUPER_ADMIN roles"})
				return
			}

			// Admin must provide a reason with minimum 15 characters
			if len(req.Reason) < 15 {
				log.Printf("Admin provided insufficient reason length: %d", len(req.Reason))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Reason must be at least 15 characters long"})
				return
			}
		} else if cu.Role == models.RoleSuperAdmin {
			// Only first SUPER_ADMIN can grant SUPER_ADMIN role
			if req.Role == models.RoleSuperAdmin && cu.ID != firstSuperAdmin.ID {
				log.Printf("Non-first SUPER_ADMIN attempted to grant SUPER_ADMIN role. User ID: %d", cu.ID)
				c.JSON(http.StatusForbidden, gin.H{"error": "Only first SUPER_ADMIN can grant SUPER_ADMIN role"})
				return
			}
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			log.Printf("Failed to start transaction: %v", tx.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Create role history record
		roleHistory := models.RoleHistory{
			UserID:      targetUser.ID,
			ChangedByID: cu.ID,
			OldRole:     targetUser.Role,
			NewRole:     req.Role,
			Reason:      req.Reason,
		}

		if err := tx.Create(&roleHistory).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to create role history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role history"})
			return
		}

		// Update user role
		oldRole := targetUser.Role
		targetUser.Role = req.Role
		if err := tx.Save(&targetUser).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to update user role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
			return
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to commit transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		log.Printf("Role updated successfully. User ID: %d, Old Role: %s, New Role: %s", targetUser.ID, oldRole, targetUser.Role)
		c.JSON(http.StatusOK, gin.H{
			"message": "Role updated successfully",
			"user": gin.H{
				"id":         targetUser.ID,
				"username":   targetUser.Username,
				"role":       targetUser.Role,
				"updated_at": targetUser.UpdatedAt,
			},
			"role_history": gin.H{
				"old_role":      oldRole,
				"new_role":      targetUser.Role,
				"reason":        req.Reason,
				"changed_by_id": cu.ID,
				"changed_at":    roleHistory.CreatedAt,
			},
		})
	}
} 