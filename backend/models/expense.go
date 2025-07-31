package models

import "gorm.io/gorm"

type Expense struct {
	gorm.Model
	Title         string  `json:"title" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Category      string  `json:"category"`
	Date          string  `json:"date" binding:"required,datetime=2006-01-02"`
	Type          string  `json:"type" binding:"required,oneof=income expense"` // Add type field
	PaymentMethod string  `json:"payment_method"`
	Notes         string  `json:"notes"`
	UserID        uint    `json:"-"`
}
