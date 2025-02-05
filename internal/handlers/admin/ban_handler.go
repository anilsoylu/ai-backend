package admin

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ai-backend/internal/models"
)

type BanUserRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Reason   string `json:"reason" binding:"required,min=15"`
	Duration string `json:"duration" binding:"required"` // number of days or "permanent"
}

// calculateBanEndDate calculates the end date based on duration
func calculateBanEndDate(duration string) (*time.Time, *int, error) {
	if duration == string(models.BanDurationPermanent) {
		return nil, nil, nil
	}

	// Parse duration as number of days
	var days int
	if _, err := fmt.Sscanf(duration, "%d", &days); err != nil {
		return nil, nil, fmt.Errorf("invalid duration format: must be a number or 'permanent'")
	}

	if days < 1 {
		return nil, nil, fmt.Errorf("duration must be at least 1 day")
	}

	endDate := time.Now().AddDate(0, 0, days)
	return &endDate, &days, nil
}

func BanUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BanUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err)})
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
		if targetUser.ID == firstSuperAdmin.ID {
			log.Printf("Attempt to ban first SUPER_ADMIN. Target ID: %d", targetUser.ID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot ban first SUPER_ADMIN"})
			return
		}

		if cu.Role == models.RoleAdmin {
			// Admin cannot ban SUPER_ADMIN
			if targetUser.Role == models.RoleSuperAdmin {
				log.Printf("Admin attempted to ban SUPER_ADMIN. Target ID: %d", targetUser.ID)
				c.JSON(http.StatusForbidden, gin.H{"error": "Admin cannot ban SUPER_ADMIN"})
				return
			}
		}

		// Calculate ban end date
		endDate, durationDays, err := calculateBanEndDate(req.Duration)
		if err != nil {
			log.Printf("Invalid duration: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			log.Printf("Failed to start transaction: %v", tx.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Create ban history record
		now := time.Now()
		banHistory := models.BanHistory{
			UserID:       targetUser.ID,
			BannedByID:   cu.ID,
			Reason:       req.Reason,
			Duration:     models.BanDurationType(req.Duration),
			DurationDays: durationDays,
			StartDate:    now,
			EndDate:      endDate,
			IsActive:     true,
		}

		if err := tx.Create(&banHistory).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to create ban history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ban history"})
			return
		}

		// Update user status
		targetUser.Status = models.StatusBanned
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

		durationText := "permanent"
		if durationDays != nil {
			durationText = fmt.Sprintf("%d days", *durationDays)
		}

		log.Printf("User banned successfully. User ID: %d, Banned by: %d, Duration: %s", targetUser.ID, cu.ID, durationText)
		c.JSON(http.StatusOK, gin.H{
			"message": "User banned successfully",
			"ban_details": gin.H{
				"user_id":       targetUser.ID,
				"username":      targetUser.Username,
				"banned_by":     cu.Username,
				"reason":        req.Reason,
				"duration":      durationText,
				"duration_days": durationDays,
				"start_date":    banHistory.StartDate,
				"end_date":      banHistory.EndDate,
				"created_at":    banHistory.CreatedAt,
			},
		})
	}
} 