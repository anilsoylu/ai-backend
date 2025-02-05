package models

import (
	"time"

	"gorm.io/gorm"
)

type VerificationToken struct {
	Identifier string    `gorm:"primaryKey"`
	Token      string    `gorm:"primaryKey"`
	Expires    time.Time `gorm:"not null"`
}

type Account struct {
	gorm.Model
	UserID           uint   `gorm:"not null"`
	Type             string `gorm:"type:varchar(255);not null"`
	Provider         string `gorm:"type:varchar(255);not null"`
	ProviderAccountID string `gorm:"type:varchar(255);not null;column:providerAccountId"`
	RefreshToken     *string `gorm:"type:text;column:refresh_token"`
	AccessToken      *string `gorm:"type:text;column:access_token"`
	ExpiresAt        *int64  `gorm:"column:expires_at"`
	IDToken          *string `gorm:"type:text;column:id_token"`
	Scope            *string `gorm:"type:text"`
	SessionState     *string `gorm:"type:text;column:session_state"`
	TokenType        *string `gorm:"type:text;column:token_type"`
	User             User    `gorm:"foreignKey:UserID"`
}

type Session struct {
	gorm.Model
	UserID       uint      `gorm:"not null"`
	Expires      time.Time `gorm:"not null"`
	SessionToken string    `gorm:"type:varchar(255);not null;column:sessionToken"`
	User         User      `gorm:"foreignKey:UserID"`
} 