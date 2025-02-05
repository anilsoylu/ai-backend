package models

import (
	"gorm.io/gorm"
)

type Question struct {
	gorm.Model
	Title       string `gorm:"type:varchar(255);not null"`
	Content     string `gorm:"type:text;not null"`
	UserID      uint   `gorm:"not null"`
	ViewCount   uint   `gorm:"default:0"`
	VoteCount   int    `gorm:"default:0"`
	IsResolved  bool   `gorm:"default:false"`
	
	// Relations
	User    User     `gorm:"foreignKey:UserID"`
	Answers []Answer `gorm:"foreignKey:QuestionID"`
	Votes   []Vote   `gorm:"foreignKey:QuestionID"`
}

type Answer struct {
	gorm.Model
	Content    string `gorm:"type:text;not null"`
	UserID     uint   `gorm:"not null"`
	QuestionID uint   `gorm:"not null"`
	VoteCount  int    `gorm:"default:0"`
	IsAccepted bool   `gorm:"default:false"`
	
	// Relations
	User     User   `gorm:"foreignKey:UserID"`
	Question Question `gorm:"foreignKey:QuestionID"`
	Votes    []Vote  `gorm:"foreignKey:AnswerID"`
}

type VoteType string

const (
	VoteUp   VoteType = "up"
	VoteDown VoteType = "down"
)

type Vote struct {
	gorm.Model
	UserID     uint     `gorm:"not null"`
	QuestionID *uint    `gorm:"default:null"`
	AnswerID   *uint    `gorm:"default:null"`
	VoteType   VoteType `gorm:"type:varchar(10);not null"`
	
	// Relations
	User     User      `gorm:"foreignKey:UserID"`
	Question *Question `gorm:"foreignKey:QuestionID"`
	Answer   *Answer   `gorm:"foreignKey:AnswerID"`
} 