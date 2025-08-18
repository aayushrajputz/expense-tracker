package models

import (
	"time"

	"gorm.io/gorm"
)

type OTP struct {
	gorm.Model
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used" gorm:"default:false"`
}
