package models

import (
	"time"

	"gorm.io/gorm"
)

type UserStatus string
type UserRole string

const (
	StatusActive  UserStatus = "active"
	StatusPassive UserStatus = "passive"
	StatusBanned  UserStatus = "banned"
	StatusFrozen  UserStatus = "frozen"

	RoleUser       UserRole = "USER"
	RoleEditor     UserRole = "EDITOR"
	RoleAdmin      UserRole = "ADMIN"
	RoleSuperAdmin UserRole = "SUPER_ADMIN"
)

type User struct {
	gorm.Model
	Name          *string    `gorm:"type:varchar(255)"`
	Username      *string    `gorm:"type:varchar(255);uniqueIndex"`
	Email         *string    `gorm:"type:varchar(255);uniqueIndex"`
	EmailVerified *time.Time `gorm:"column:emailVerified"`
	Password      *string    `gorm:"type:text"`
	Image         *string    `gorm:"type:text"`
	Role          UserRole   `gorm:"type:varchar(50);not null;default:'USER'"`
	Status        UserStatus `gorm:"type:varchar(50);not null;default:'active'"`
	
	// Relations
	Accounts  []Account  `gorm:"foreignKey:UserID"`
	Sessions  []Session  `gorm:"foreignKey:UserID"`
	Questions []Question `gorm:"foreignKey:UserID"`
	Answers   []Answer   `gorm:"foreignKey:UserID"`
	Votes     []Vote     `gorm:"foreignKey:UserID"`
} 