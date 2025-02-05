package admin

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ai-backend/internal/models"
)

type UnbanUserRequest struct {
	Reason string `json:"reason" binding:"required,min=15"`
}

func UnbanUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
		if err != nil {
			log.Printf("Invalid user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var req UnbanUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Get current user from context
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
		if err := db.First(&targetUser, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("User not found with ID: %d", userID)
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			log.Printf("Database error while fetching user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Check if user is banned
		if targetUser.Status != models.StatusBanned {
			log.Printf("User is not banned. User ID: %d, Status: %s", targetUser.ID, targetUser.Status)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User is not banned"})
			return
		}

		// Get active ban record
		var activeBan models.BanHistory
		if err := db.Preload("BannedBy").
			Where("user_id = ? AND is_active = ?", userID, true).
			First(&activeBan).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("No active ban found for user ID: %d", userID)
				c.JSON(http.StatusNotFound, gin.H{"error": "No active ban found"})
				return
			}
			log.Printf("Database error while fetching active ban: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Check if ADMIN is trying to unban a user banned by SUPER_ADMIN
		if cu.Role == models.RoleAdmin && activeBan.BannedBy.Role == models.RoleSuperAdmin {
			log.Printf("Admin attempted to unban user banned by SUPER_ADMIN. User ID: %d", targetUser.ID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot unban user banned by SUPER_ADMIN"})
			return
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			log.Printf("Failed to start transaction: %v", tx.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		now := time.Now()

		// Update old ban record
		activeBan.IsActive = false
		activeBan.UnbannedAt = &now
		activeBan.UnbannedBy = &cu.ID

		if err := tx.Save(&activeBan).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to update ban record: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ban record"})
			return
		}

		// Create new ban history record for unban action
		unbanHistory := models.BanHistory{
			UserID:      targetUser.ID,
			BannedByID:  cu.ID,
			Reason:      req.Reason,
			Duration:    "unban",
			StartDate:   now,
			IsActive:    false,
			UnbannedAt:  &now,
			UnbannedBy:  &cu.ID,
		}

		if err := tx.Create(&unbanHistory).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to create unban history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unban history"})
			return
		}

		// Update user status
		targetUser.Status = models.StatusActive
		if err := tx.Save(&targetUser).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to update user status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
			return
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to commit transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		log.Printf("User unbanned successfully. User ID: %d, Unbanned by: %d", targetUser.ID, cu.ID)
		c.JSON(http.StatusOK, gin.H{
			"message": "User unbanned successfully",
			"unban_details": gin.H{
				"user_id":     targetUser.ID,
				"username":    targetUser.Username,
				"unbanned_by": cu.Username,
				"reason":      req.Reason,
				"unbanned_at": now.Format("2006-01-02 15:04:05"),
			},
		})
	}
} 