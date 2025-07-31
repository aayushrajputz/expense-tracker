package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/your-github/expense-tracker-backend/models"
	"gorm.io/gorm"
)

type TransactionService struct {
	DB *gorm.DB
}

type MockTransaction struct {
	TransactionID   string    `json:"transaction_id"`
	TransactionDate time.Time `json:"transaction_date"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"`
	Category        string    `json:"category"`
	Balance         float64   `json:"balance"`
	ReferenceNumber string    `json:"reference_number"`
	MerchantName    string    `json:"merchant_name"`
	Location        string    `json:"location"`
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{DB: db}
}

// GenerateMockTransactions generates realistic mock transactions for a bank account
func (s *TransactionService) GenerateMockTransactions(bankAccountID uint, userID uint, count int) ([]MockTransaction, error) {
	var transactions []MockTransaction

	// Categories for different transaction types
	categories := map[string][]string{
		"debit": {
			"Food & Dining", "Shopping", "Transportation", "Entertainment",
			"Utilities", "Healthcare", "Education", "Travel", "Insurance",
			"ATM Withdrawal", "Online Shopping", "Restaurant", "Fuel",
		},
		"credit": {
			"Salary", "Transfer", "Refund", "Interest", "Dividend",
			"Reimbursement", "Bonus", "Investment Return",
		},
	}

	// Merchant names for different categories
	merchants := map[string][]string{
		"Food & Dining":  {"McDonald's", "Domino's", "KFC", "Pizza Hut", "Subway", "Starbucks"},
		"Shopping":       {"Amazon", "Flipkart", "Myntra", "Reliance Digital", "Croma", "Big Bazaar"},
		"Transportation": {"Uber", "Ola", "Metro", "Bus", "Railway", "Airport"},
		"Entertainment":  {"Netflix", "Amazon Prime", "Hotstar", "Movie Theater", "Gym"},
		"Utilities":      {"Electricity Bill", "Water Bill", "Gas Bill", "Internet Bill"},
		"Healthcare":     {"Apollo Hospital", "Fortis", "Max Hospital", "Pharmacy"},
		"Education":      {"School Fee", "College Fee", "Course Fee", "Book Store"},
		"Travel":         {"MakeMyTrip", "Goibibo", "IRCTC", "Air India", "IndiGo"},
		"Insurance":      {"LIC", "HDFC Life", "ICICI Prudential", "Bajaj Allianz"},
		"Salary":         {"Company XYZ", "Tech Corp", "Startup Inc", "MNC Ltd"},
		"Transfer":       {"NEFT Transfer", "IMPS Transfer", "UPI Transfer", "RTGS Transfer"},
		"Refund":         {"Amazon Refund", "Flipkart Refund", "Restaurant Refund"},
		"Interest":       {"Bank Interest", "FD Interest", "Savings Interest"},
	}

	// Generate transactions for the last 30 days
	endDate := time.Now()

	currentBalance := 50000.0 // Starting balance

	for i := 0; i < count; i++ {
		// Random date within the last 30 days
		daysAgo := rand.Intn(30)
		transactionDate := endDate.AddDate(0, 0, -daysAgo)

		// Random transaction type (70% debit, 30% credit)
		transactionType := "debit"
		if rand.Float64() < 0.3 {
			transactionType = "credit"
		}

		// Select category
		categoryList := categories[transactionType]
		category := categoryList[rand.Intn(len(categoryList))]

		// Generate amount based on category and type
		amount := s.generateAmount(category, transactionType)

		// Update balance
		if transactionType == "debit" {
			currentBalance -= amount
		} else {
			currentBalance += amount
		}

		// Select merchant
		merchantList := merchants[category]
		merchantName := "Unknown"
		if len(merchantList) > 0 {
			merchantName = merchantList[rand.Intn(len(merchantList))]
		}

		// Generate description
		description := s.generateDescription(category, merchantName, transactionType)

		transaction := MockTransaction{
			TransactionID:   fmt.Sprintf("TXN%d%d", time.Now().Unix(), i),
			TransactionDate: transactionDate,
			Description:     description,
			Amount:          amount,
			Type:            transactionType,
			Category:        category,
			Balance:         currentBalance,
			ReferenceNumber: fmt.Sprintf("REF%d", rand.Intn(999999)),
			MerchantName:    merchantName,
			Location:        s.generateLocation(),
		}

		transactions = append(transactions, transaction)
	}

	// Sort transactions by date (newest first)
	for i := 0; i < len(transactions)-1; i++ {
		for j := i + 1; j < len(transactions); j++ {
			if transactions[i].TransactionDate.Before(transactions[j].TransactionDate) {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}

	return transactions, nil
}

// StoreTransactions stores transactions in the database
func (s *TransactionService) StoreTransactions(transactions []MockTransaction, bankAccountID uint, userID uint) error {
	for _, txn := range transactions {
		transaction := models.Transaction{
			UserID:          userID,
			BankAccountID:   bankAccountID,
			TransactionID:   txn.TransactionID,
			TransactionDate: txn.TransactionDate,
			Description:     txn.Description,
			Amount:          txn.Amount,
			Type:            txn.Type,
			Category:        txn.Category,
			Balance:         txn.Balance,
			ReferenceNumber: txn.ReferenceNumber,
			MerchantName:    txn.MerchantName,
			Location:        txn.Location,
			Status:          "completed",
		}

		// Check if transaction already exists
		var existing models.Transaction
		if err := s.DB.Where("transaction_id = ?", txn.TransactionID).First(&existing).Error; err == nil {
			// Transaction already exists, skip
			continue
		}

		if err := s.DB.Create(&transaction).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetTransactions retrieves transactions for a user, sorted by date
func (s *TransactionService) GetTransactions(userID uint, limit int, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	// Get all transactions (both bank and manual) from the transactions table
	query := s.DB.Where("user_id = ?", userID).
		Preload("BankAccount").
		Order("transaction_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&transactions).Error
	return transactions, err
}

// GetTransactionsByBankAccount retrieves transactions for a specific bank account
func (s *TransactionService) GetTransactionsByBankAccount(userID uint, bankAccountID uint, limit int, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	query := s.DB.Where("user_id = ? AND bank_account_id = ?", userID, bankAccountID).
		Preload("BankAccount").
		Order("transaction_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&transactions).Error
	return transactions, err
}

// Helper functions
func (s *TransactionService) generateAmount(category, transactionType string) float64 {
	switch category {
	case "Salary":
		return float64(rand.Intn(50000) + 30000) // 30k-80k
	case "Transfer":
		return float64(rand.Intn(10000) + 1000) // 1k-11k
	case "Food & Dining":
		return float64(rand.Intn(500) + 100) // 100-600
	case "Shopping":
		return float64(rand.Intn(2000) + 200) // 200-2200
	case "Transportation":
		return float64(rand.Intn(300) + 50) // 50-350
	case "Entertainment":
		return float64(rand.Intn(1000) + 100) // 100-1100
	case "Utilities":
		return float64(rand.Intn(2000) + 500) // 500-2500
	case "Healthcare":
		return float64(rand.Intn(5000) + 500) // 500-5500
	case "Education":
		return float64(rand.Intn(10000) + 1000) // 1k-11k
	case "Travel":
		return float64(rand.Intn(15000) + 2000) // 2k-17k
	case "Insurance":
		return float64(rand.Intn(5000) + 1000) // 1k-6k
	default:
		return float64(rand.Intn(1000) + 100) // 100-1100
	}
}

func (s *TransactionService) generateDescription(category, merchant, transactionType string) string {
	switch category {
	case "Salary":
		return fmt.Sprintf("Salary credit from %s", merchant)
	case "Transfer":
		return fmt.Sprintf("Money %s via %s", transactionType, merchant)
	case "Food & Dining":
		return fmt.Sprintf("Payment to %s", merchant)
	case "Shopping":
		return fmt.Sprintf("Online purchase from %s", merchant)
	case "Transportation":
		return fmt.Sprintf("Transport fare to %s", merchant)
	case "Entertainment":
		return fmt.Sprintf("Entertainment at %s", merchant)
	case "Utilities":
		return fmt.Sprintf("Utility bill payment to %s", merchant)
	case "Healthcare":
		return fmt.Sprintf("Medical expense at %s", merchant)
	case "Education":
		return fmt.Sprintf("Education fee to %s", merchant)
	case "Travel":
		return fmt.Sprintf("Travel booking with %s", merchant)
	case "Insurance":
		return fmt.Sprintf("Insurance premium to %s", merchant)
	default:
		return fmt.Sprintf("Transaction with %s", merchant)
	}
}

func (s *TransactionService) generateLocation() string {
	locations := []string{
		"Mumbai, Maharashtra", "Delhi, NCR", "Bangalore, Karnataka",
		"Chennai, Tamil Nadu", "Kolkata, West Bengal", "Hyderabad, Telangana",
		"Pune, Maharashtra", "Ahmedabad, Gujarat", "Jaipur, Rajasthan",
		"Lucknow, Uttar Pradesh", "Online", "N/A",
	}
	return locations[rand.Intn(len(locations))]
}
