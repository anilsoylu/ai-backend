package models

import (
	"gorm.io/gorm"
)

// RoleHistory represents the history of role changes
type RoleHistory struct {
	gorm.Model
	UserID      uint      `gorm:"not null;index"`
	ChangedByID uint      `gorm:"not null;index"`
	OldRole     UserRole  `gorm:"type:varchar(50);not null"`
	NewRole     UserRole  `gorm:"type:varchar(50);not null"`
	Reason      string    `gorm:"type:text;not null"`
	
	// Relations
	User      User `gorm:"foreignKey:UserID"`
	ChangedBy User `gorm:"foreignKey:ChangedByID"`
} 