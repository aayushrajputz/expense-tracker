package models

import (
	"time"

	"gorm.io/gorm"
)

type BankAccount struct {
	gorm.Model
	UserID            uint       `json:"user_id" gorm:"not null"`
	BankID            string     `json:"bank_id" gorm:"not null"`
	AccountNumber     string     `json:"account_number" gorm:"not null"`
	AccountHolderName string     `json:"account_holder_name" gorm:"not null"`
	MobileNumber      string     `json:"mobile_number" gorm:"not null"`
	Status            string     `json:"status" gorm:"default:'ACTIVE'"`
	AccountType       string     `json:"account_type"`
	IFSCCode          string     `json:"ifsc_code"`
	BranchName        string     `json:"branch_name"`
	VerifiedAt        *time.Time `json:"verified_at"`
	User              User       `gorm:"foreignKey:UserID"`
}
