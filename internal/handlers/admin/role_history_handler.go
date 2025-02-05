package admin

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ai-backend/internal/models"
)

type RoleHistoryResponse struct {
	ID          uint           `json:"id"`
	UserID      uint           `json:"user_id"`
	Username    string         `json:"username"`
	ChangedByID uint           `json:"changed_by_id"`
	ChangedBy   string         `json:"changed_by"`
	OldRole     models.UserRole `json:"old_role"`
	NewRole     models.UserRole `json:"new_role"`
	Reason      string         `json:"reason"`
	CreatedAt   string         `json:"created_at"`
}

// GetUserRoleHistory returns the role change history for a specific user
func GetUserRoleHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
		if err != nil {
			log.Printf("Invalid user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Check if user exists
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			log.Printf("Database error while fetching user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var histories []models.RoleHistory
		if err := db.Preload("User").Preload("ChangedBy").
			Where("user_id = ?", userID).
			Order("created_at desc").
			Find(&histories).Error; err != nil {
			log.Printf("Failed to fetch role histories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role histories"})
			return
		}

		response := make([]RoleHistoryResponse, len(histories))
		for i, history := range histories {
			response[i] = RoleHistoryResponse{
				ID:          history.ID,
				UserID:      history.UserID,
				Username:    *history.User.Username,
				ChangedByID: history.ChangedByID,
				ChangedBy:   *history.ChangedBy.Username,
				OldRole:     history.OldRole,
				NewRole:     history.NewRole,
				Reason:      history.Reason,
				CreatedAt:   history.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
			},
			"histories": response,
		})
	}
}

// GetAllRoleHistories returns all role change histories with pagination
func GetAllRoleHistories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 50 {
			limit = 10
		}

		offset := (page - 1) * limit

		var total int64
		if err := db.Model(&models.RoleHistory{}).Count(&total).Error; err != nil {
			log.Printf("Failed to count role histories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count role histories"})
			return
		}

		var histories []models.RoleHistory
		if err := db.Preload("User").Preload("ChangedBy").
			Order("created_at desc").
			Offset(offset).Limit(limit).
			Find(&histories).Error; err != nil {
			log.Printf("Failed to fetch role histories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role histories"})
			return
		}

		response := make([]RoleHistoryResponse, len(histories))
		for i, history := range histories {
			response[i] = RoleHistoryResponse{
				ID:          history.ID,
				UserID:      history.UserID,
				Username:    *history.User.Username,
				ChangedByID: history.ChangedByID,
				ChangedBy:   *history.ChangedBy.Username,
				OldRole:     history.OldRole,
				NewRole:     history.NewRole,
				Reason:      history.Reason,
				CreatedAt:   history.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		totalPages := (int(total) + limit - 1) / limit

		c.JSON(http.StatusOK, gin.H{
			"histories": response,
			"pagination": gin.H{
				"current_page": page,
				"total_pages": totalPages,
				"total_items": total,
				"per_page":    limit,
				"has_next":    page < totalPages,
				"has_prev":    page > 1,
			},
		})
	}
} 