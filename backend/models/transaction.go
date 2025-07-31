package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	UserID          uint        `json:"user_id" gorm:"not null"`
	BankAccountID   uint        `json:"bank_account_id" gorm:"not null"`
	TransactionID   string      `json:"transaction_id" gorm:"uniqueIndex"`
	TransactionDate time.Time   `json:"transaction_date" gorm:"not null"`
	Description     string      `json:"description"`
	Amount          float64     `json:"amount" gorm:"not null"`
	Type            string      `json:"type" gorm:"not null"` // "credit" or "debit"
	Category        string      `json:"category"`
	Balance         float64     `json:"balance"`
	ReferenceNumber string      `json:"reference_number"`
	MerchantName    string      `json:"merchant_name"`
	Location        string      `json:"location"`
	Status          string      `json:"status" gorm:"default:'completed'"`
	BankAccount     BankAccount `gorm:"foreignKey:BankAccountID"`
	User            User        `gorm:"foreignKey:UserID"`
}
