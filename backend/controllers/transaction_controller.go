package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-github/expense-tracker-backend/services"
)

type TransactionController struct {
	TransactionService *services.TransactionService
}

type TransactionResponse struct {
	ID              uint    `json:"id"`
	TransactionID   string  `json:"transaction_id"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Type            string  `json:"type"`
	Category        string  `json:"category"`
	Balance         float64 `json:"balance"`
	ReferenceNumber string  `json:"reference_number"`
	MerchantName    string  `json:"merchant_name"`
	Location        string  `json:"location"`
	Status          string  `json:"status"`
	BankAccount     struct {
		ID            uint   `json:"id"`
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
	} `json:"bank_account"`
}

// GetTransactionHistory returns all transactions for the authenticated user
func (c *TransactionController) GetTransactionHistory(ctx *gin.Context) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get query parameters for pagination
	limitStr := ctx.Query("limit")
	offsetStr := ctx.Query("offset")

	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	transactions, err := c.TransactionService.GetTransactions(userID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	// Convert to response format
	var response []TransactionResponse
	for _, txn := range transactions {
		response = append(response, TransactionResponse{
			ID:              txn.ID,
			TransactionID:   txn.TransactionID,
			TransactionDate: txn.TransactionDate.Format("2006-01-02T15:04:05Z"),
			Description:     txn.Description,
			Amount:          txn.Amount,
			Type:            txn.Type,
			Category:        txn.Category,
			Balance:         txn.Balance,
			ReferenceNumber: txn.ReferenceNumber,
			MerchantName:    txn.MerchantName,
			Location:        txn.Location,
			Status:          txn.Status,
			BankAccount: struct {
				ID            uint   `json:"id"`
				BankName      string `json:"bank_name"`
				AccountNumber string `json:"account_number"`
			}{
				ID:       txn.BankAccount.ID,
				BankName: getBankName(txn.BankAccount.BankID),
				AccountNumber: func() string {
					if txn.BankAccount.BankID == "MANUAL" {
						return "Manual Entry"
					}
					return maskAccountNumber(txn.BankAccount.AccountNumber)
				}(),
			},
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactions": response,
		"count":        len(response),
		"limit":        limit,
		"offset":       offset,
	})
}

// GetTransactionsByBankAccount returns transactions for a specific bank account
func (c *TransactionController) GetTransactionsByBankAccount(ctx *gin.Context) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	accountID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	// Get query parameters for pagination
	limitStr := ctx.Query("limit")
	offsetStr := ctx.Query("offset")

	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	transactions, err := c.TransactionService.GetTransactionsByBankAccount(userID, uint(accountID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	// Convert to response format
	var response []TransactionResponse
	for _, txn := range transactions {
		response = append(response, TransactionResponse{
			ID:              txn.ID,
			TransactionID:   txn.TransactionID,
			TransactionDate: txn.TransactionDate.Format("2006-01-02T15:04:05Z"),
			Description:     txn.Description,
			Amount:          txn.Amount,
			Type:            txn.Type,
			Category:        txn.Category,
			Balance:         txn.Balance,
			ReferenceNumber: txn.ReferenceNumber,
			MerchantName:    txn.MerchantName,
			Location:        txn.Location,
			Status:          txn.Status,
			BankAccount: struct {
				ID            uint   `json:"id"`
				BankName      string `json:"bank_name"`
				AccountNumber string `json:"account_number"`
			}{
				ID:       txn.BankAccount.ID,
				BankName: getBankName(txn.BankAccount.BankID),
				AccountNumber: func() string {
					if txn.BankAccount.BankID == "MANUAL" {
						return "Manual Entry"
					}
					return maskAccountNumber(txn.BankAccount.AccountNumber)
				}(),
			},
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactions": response,
		"count":        len(response),
		"accountId":    accountID,
		"limit":        limit,
		"offset":       offset,
	})
}
