package models

import (
	"time"

	"gorm.io/gorm"
)

type BanDurationType string

const (
	BanDurationPermanent BanDurationType = "permanent"
)

// BanHistory represents the history of user bans
type BanHistory struct {
	gorm.Model
	UserID      uint           `gorm:"not null;index"`
	BannedByID  uint           `gorm:"not null;index"`
	Reason      string         `gorm:"type:text;not null"`
	Duration    BanDurationType `gorm:"type:varchar(20);not null"`
	DurationDays *int          `gorm:"default:null"` // null for permanent bans
	StartDate   time.Time      `gorm:"not null"`
	EndDate     *time.Time     `gorm:"default:null"` // null for permanent bans
	IsActive    bool           `gorm:"not null;default:true"`
	UnbannedAt  *time.Time     `gorm:"default:null"`
	UnbannedBy  *uint          `gorm:"default:null"`
	
	// Relations
	User      User `gorm:"foreignKey:UserID"`
	BannedBy  User `gorm:"foreignKey:BannedByID"`
	Unbanner  User `gorm:"foreignKey:UnbannedBy"`
} 