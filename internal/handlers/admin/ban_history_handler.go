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

type BanHistoryResponse struct {
	ID           uint                  `json:"id"`
	UserID       uint                  `json:"user_id"`
	Username     string                `json:"username"`
	BannedByID   uint                  `json:"banned_by_id"`
	BannedBy     string                `json:"banned_by"`
	Reason       string                `json:"reason"`
	Duration     string                `json:"duration"`
	DurationDays *int                  `json:"duration_days"`
	StartDate    string                `json:"start_date"`
	EndDate      *string               `json:"end_date"`
	IsActive     bool                  `json:"is_active"`
	UnbannedAt   *string               `json:"unbanned_at"`
	UnbannedBy   *string               `json:"unbanned_by"`
	CreatedAt    string                `json:"created_at"`
}

// GetUserBanHistory returns the ban history for a specific user
func GetUserBanHistory(db *gorm.DB) gin.HandlerFunc {
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

		var histories []models.BanHistory
		if err := db.Preload("User").Preload("BannedBy").Preload("Unbanner").
			Where("user_id = ?", userID).
			Order("created_at desc").
			Find(&histories).Error; err != nil {
			log.Printf("Failed to fetch ban histories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ban histories"})
			return
		}

		response := make([]BanHistoryResponse, len(histories))
		for i, history := range histories {
			var endDate, unbannedAt *string
			if history.EndDate != nil {
				formatted := history.EndDate.Format("2006-01-02 15:04:05")
				endDate = &formatted
			}
			if history.UnbannedAt != nil {
				formatted := history.UnbannedAt.Format("2006-01-02 15:04:05")
				unbannedAt = &formatted
			}

			var unbannedBy *string
			if history.UnbannedBy != nil {
				var unbanner models.User
				if err := db.First(&unbanner, *history.UnbannedBy).Error; err == nil {
					username := *unbanner.Username
					unbannedBy = &username
				}
			}

			durationText := "permanent"
			if history.DurationDays != nil {
				durationText = strconv.Itoa(*history.DurationDays) + " days"
			}

			response[i] = BanHistoryResponse{
				ID:           history.ID,
				UserID:       history.UserID,
				Username:     *history.User.Username,
				BannedByID:   history.BannedByID,
				BannedBy:     *history.BannedBy.Username,
				Reason:       history.Reason,
				Duration:     durationText,
				DurationDays: history.DurationDays,
				StartDate:    history.StartDate.Format("2006-01-02 15:04:05"),
				EndDate:      endDate,
				IsActive:     history.IsActive,
				UnbannedAt:   unbannedAt,
				UnbannedBy:   unbannedBy,
				CreatedAt:    history.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"status":   user.Status,
			},
			"histories": response,
		})
	}
}

// GetAllBanHistories returns all ban histories with pagination
func GetAllBanHistories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		status := c.Query("status")
		durationType := c.Query("duration")

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 50 {
			limit = 10
		}

		offset := (page - 1) * limit

		// Build query
		query := db.Model(&models.BanHistory{})

		// Apply filters
		if status == "active" {
			query = query.Where("is_active = ?", true)
		} else if status == "inactive" {
			query = query.Where("is_active = ?", false)
		}

		if durationType == "permanent" {
			query = query.Where("duration_days IS NULL")
		} else if durationType == "temporary" {
			query = query.Where("duration_days IS NOT NULL")
		}

		// Count total records
		var total int64
		if err := query.Count(&total).Error; err != nil {
			log.Printf("Failed to count ban histories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count ban histories"})
			return
		}

		// Get records
		var histories []models.BanHistory
		if err := query.Preload("User").Preload("BannedBy").Preload("Unbanner").
			Order("created_at desc").
			Offset(offset).Limit(limit).
			Find(&histories).Error; err != nil {
			log.Printf("Failed to fetch ban histories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ban histories"})
			return
		}

		response := make([]BanHistoryResponse, len(histories))
		for i, history := range histories {
			var endDate, unbannedAt *string
			if history.EndDate != nil {
				formatted := history.EndDate.Format("2006-01-02 15:04:05")
				endDate = &formatted
			}
			if history.UnbannedAt != nil {
				formatted := history.UnbannedAt.Format("2006-01-02 15:04:05")
				unbannedAt = &formatted
			}

			var unbannedBy *string
			if history.UnbannedBy != nil {
				var unbanner models.User
				if err := db.First(&unbanner, *history.UnbannedBy).Error; err == nil {
					username := *unbanner.Username
					unbannedBy = &username
				}
			}

			durationText := "permanent"
			if history.DurationDays != nil {
				durationText = strconv.Itoa(*history.DurationDays) + " days"
			}

			response[i] = BanHistoryResponse{
				ID:           history.ID,
				UserID:       history.UserID,
				Username:     *history.User.Username,
				BannedByID:   history.BannedByID,
				BannedBy:     *history.BannedBy.Username,
				Reason:       history.Reason,
				Duration:     durationText,
				DurationDays: history.DurationDays,
				StartDate:    history.StartDate.Format("2006-01-02 15:04:05"),
				EndDate:      endDate,
				IsActive:     history.IsActive,
				UnbannedAt:   unbannedAt,
				UnbannedBy:   unbannedBy,
				CreatedAt:    history.CreatedAt.Format("2006-01-02 15:04:05"),
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