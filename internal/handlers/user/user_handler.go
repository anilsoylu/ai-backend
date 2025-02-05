package user

import (
	"ai-backend/internal/database"
	"ai-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UpdateStatusRequest struct {
	UserID uint              `json:"user_id" binding:"required"`
	Status models.UserStatus `json:"status" binding:"required"`
}

// isUserAllowedStatus checks if the status is allowed for regular users
func isUserAllowedStatus(status models.UserStatus) bool {
	allowedStatuses := []models.UserStatus{
		models.StatusActive,
		models.StatusPassive,
		models.StatusFrozen,
	}

	for _, s := range allowedStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// UpdateUserStatus handles user status updates
func UpdateUserStatus(c *gin.Context) {
	// Get current user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	currentUser := userInterface.(models.User)

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find target user
	var targetUser models.User
	if err := database.DB.First(&targetUser, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check permissions and validate status
	isAdmin := currentUser.Role == models.RoleAdmin || currentUser.Role == models.RoleSuperAdmin
	isSelfUpdate := currentUser.ID == targetUser.ID

	// Regular users can only update their own status
	if !isAdmin && !isSelfUpdate {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own status"})
		return
	}

	// Regular users can't set banned status
	if !isAdmin && req.Status == models.StatusBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can set banned status"})
		return
	}

	// Regular users can only set allowed statuses
	if !isAdmin && !isUserAllowedStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status for user"})
		return
	}

	// Prevent status update of SUPER_ADMIN by ADMIN
	if targetUser.Role == models.RoleSuperAdmin && currentUser.Role == models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify SUPER_ADMIN status"})
		return
	}

	// Update user status
	targetUser.Status = req.Status
	if err := database.DB.Save(&targetUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User status updated successfully",
		"user": gin.H{
			"id":     targetUser.ID,
			"status": targetUser.Status,
		},
	})
} 