package models

import (
	"time"

	"gorm.io/gorm"
)

type FreezeHistory struct {
	gorm.Model
	UserID      uint      `gorm:"not null;index"`
	User        User      `gorm:"foreignKey:UserID"`
	Reason      string    `gorm:"type:text;not null"`
	Duration    int       `gorm:"not null"` // Gün cinsinden dondurma süresi
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
	IsActive    bool      `gorm:"not null;default:true"` // Dondurma işlemi hala aktif mi?
	UnfrozenAt  *time.Time // Kullanıcı manuel olarak çözdüyse bu tarih dolar
} 