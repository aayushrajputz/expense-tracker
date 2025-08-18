package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string  `json:"name" binding:"required"`
	Email    string  `gorm:"uniqueIndex" json:"email" binding:"required,email"`
	Password string  `json:"password,omitempty" binding:"required"`
	GoogleID *string `json:"google_id,omitempty"`
	Budget   float64
	Expenses []Expense `gorm:"constraint:OnDelete:CASCADE;"`
}
