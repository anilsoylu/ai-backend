package user

import (
	"ai-backend/internal/database"
	"ai-backend/internal/models"
	"net/http"

	"ai-backend/pkg/utils"
	"log"

	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type UpdateProfileRequest struct {
	Username  *string `json:"username" binding:"omitempty,min=3"`
	Email     *string `json:"email" binding:"omitempty,email"`
	FullName  *string `json:"full_name" binding:"omitempty,max=100"`
	Bio       *string `json:"bio" binding:"omitempty,max=500"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,url"`
}

// UpdateProfile güncelleme işlemini gerçekleştirir
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Kullanıcı kimliğini al
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mevcut kullanıcıyı bul
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Email veya kullanıcı adı değişikliği varsa, benzersizlik kontrolü yap
	if req.Email != nil && *req.Email != *user.Email {
		var count int64
		h.db.Model(&models.User{}).
			Where("email = ? AND deleted_at IS NULL", req.Email).
			Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
	}

	if req.Username != nil && *req.Username != *user.Username {
		var count int64
		h.db.Model(&models.User{}).
			Where("username = ? AND deleted_at IS NULL", req.Username).
			Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
	}

	// Güncelleme işlemi
	updates := map[string]interface{}{}
	
	if req.Username != nil {
		updates["username"] = req.Username
	}
	if req.Email != nil {
		updates["email"] = req.Email
	}
	if req.FullName != nil {
		updates["name"] = req.FullName
	}
	if req.Bio != nil {
		updates["bio"] = req.Bio
	}
	if req.AvatarURL != nil {
		updates["image"] = req.AvatarURL
	}

	// Güncelleme işlemini gerçekleştir
	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Güncellenmiş kullanıcı bilgilerini getir
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

// DeleteAccount kullanıcı hesabını soft delete yapar
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	// Kullanıcı kimliğini al
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mevcut kullanıcıyı bul
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Şifre kontrolü yap
	if !utils.CheckPasswordHash(req.Password, *user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Kullanıcıyı soft delete yap
	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	// Kullanıcının tüm oturumlarını sonlandır
	if err := h.db.Where("user_id = ?", userID).Delete(&models.Session{}).Error; err != nil {
		log.Printf("Failed to delete user sessions: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account deleted successfully",
	})
}

type FreezeAccountRequest struct {
	Duration int    `json:"duration" binding:"required,min=1,max=365"`
	Reason   string `json:"reason" binding:"required,max=500"`
}

// FreezeAccount kullanıcı hesabını geçici olarak dondurur
func (h *UserHandler) FreezeAccount(c *gin.Context) {
	// Kullanıcı kimliğini al
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req FreezeAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mevcut kullanıcıyı bul
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Aktif dondurma işlemi var mı kontrol et
	var activeFreeze models.FreezeHistory
	err := h.db.Where("user_id = ? AND is_active = ? AND end_date > ?", userID, true, time.Now()).
		First(&activeFreeze).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account is already frozen"})
		return
	}

	// Yeni dondurma işlemi oluştur
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, req.Duration)

	freezeHistory := models.FreezeHistory{
		UserID:    uint(userID.(uint)),
		Reason:    req.Reason,
		Duration:  req.Duration,
		StartDate: startDate,
		EndDate:   endDate,
		IsActive:  true,
	}

	// Transaction başlat
	tx := h.db.Begin()

	// Dondurma kaydını oluştur
	if err := tx.Create(&freezeHistory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create freeze record"})
		return
	}

	// Kullanıcı durumunu güncelle
	if err := tx.Model(&user).Update("status", models.StatusFrozen).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	// Transaction'ı tamamla
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete the freeze process"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account frozen successfully",
		"freeze_details": freezeHistory,
	})
}

// GetFreezeHistory kullanıcının dondurma geçmişini getirir
func (h *UserHandler) GetFreezeHistory(c *gin.Context) {
	// Kullanıcı kimliğini al
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var freezeHistory []models.FreezeHistory
	if err := h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&freezeHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch freeze history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"freeze_history": freezeHistory,
	})
} 