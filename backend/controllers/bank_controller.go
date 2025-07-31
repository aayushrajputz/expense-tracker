package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-github/expense-tracker-backend/models"
	"github.com/your-github/expense-tracker-backend/services"
	"gorm.io/gorm"
)

type BankController struct {
	DB                  *gorm.DB
	VerificationService *services.BankVerificationService
	TransactionService  *services.TransactionService
}

type AddBankAccountRequest struct {
	BankID            string `json:"bankId" binding:"required"`
	AccountNumber     string `json:"accountNumber" binding:"required"`
	AccountHolderName string `json:"accountHolderName" binding:"required"`
	MobileNumber      string `json:"mobileNumber" binding:"required"`
}

type BankAccountResponse struct {
	ID                uint   `json:"id"`
	BankID            string `json:"bankId"`
	BankName          string `json:"bankName"`
	AccountNumber     string `json:"accountNumber"`
	AccountHolderName string `json:"accountHolderName"`
	MobileNumber      string `json:"mobileNumber"`
	Status            string `json:"status"`
	AccountType       string `json:"accountType"`
	IFSCCode          string `json:"ifscCode"`
	BranchName        string `json:"branchName"`
	VerifiedAt        string `json:"verifiedAt"`
}

// GetBankAccounts returns all bank accounts for the authenticated user
func (c *BankController) GetBankAccounts(ctx *gin.Context) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var bankAccounts []models.BankAccount
	if err := c.DB.Where("user_id = ?", userID).Find(&bankAccounts).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bank accounts"})
		return
	}

	// Convert to response format
	var response []BankAccountResponse
	for _, account := range bankAccounts {
		verifiedAt := ""
		if account.VerifiedAt != nil {
			verifiedAt = account.VerifiedAt.Format("2006-01-02T15:04:05Z")
		}

		response = append(response, BankAccountResponse{
			ID:                account.ID,
			BankID:            account.BankID,
			BankName:          getBankName(account.BankID),
			AccountNumber:     maskAccountNumber(account.AccountNumber),
			AccountHolderName: account.AccountHolderName,
			MobileNumber:      maskMobileNumber(account.MobileNumber),
			Status:            account.Status,
			AccountType:       account.AccountType,
			IFSCCode:          account.IFSCCode,
			BranchName:        account.BranchName,
			VerifiedAt:        verifiedAt,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

// AddBankAccount adds a new bank account for the authenticated user
func (c *BankController) AddBankAccount(ctx *gin.Context) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req AddBankAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Check if bank account already exists for this user
	var existingAccount models.BankAccount
	if err := c.DB.Where("user_id = ? AND bank_id = ? AND account_number = ?",
		userID, req.BankID, req.AccountNumber).First(&existingAccount).Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Bank account already exists"})
		return
	}

	// Validate mobile number and account number
	if !services.ValidateMobileNumber(req.MobileNumber) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mobile number format"})
		return
	}

	if !services.ValidateAccountNumber(req.AccountNumber) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account number format"})
		return
	}

	// Verify bank account with mobile number
	verificationReq := services.BankVerificationRequest{
		BankID:        req.BankID,
		AccountNumber: req.AccountNumber,
		MobileNumber:  req.MobileNumber,
		AccountHolder: req.AccountHolderName,
	}

	// Use mock verification for now (replace with real API call in production)
	verificationResp, err := c.VerificationService.MockVerifyBankAccount(verificationReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify bank account"})
		return
	}

	if !verificationResp.Verified {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": verificationResp.Message})
		return
	}

	// Create new bank account
	now := time.Now()
	bankAccount := models.BankAccount{
		UserID:            userID,
		BankID:            req.BankID,
		AccountNumber:     req.AccountNumber,
		AccountHolderName: req.AccountHolderName,
		MobileNumber:      req.MobileNumber,
		Status:            "ACTIVE",
		AccountType:       verificationResp.AccountType,
		IFSCCode:          verificationResp.IFSCCode,
		BranchName:        verificationResp.BranchName,
		VerifiedAt:        &now,
	}

	if err := c.DB.Create(&bankAccount).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add bank account"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Bank account added successfully",
		"id":      bankAccount.ID,
	})
}

// UpdateBankAccount updates an existing bank account
func (c *BankController) UpdateBankAccount(ctx *gin.Context) {
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

	var req AddBankAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Check if bank account exists and belongs to user
	var bankAccount models.BankAccount
	if err := c.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&bankAccount).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Bank account not found"})
		return
	}

	// Update bank account
	bankAccount.BankID = req.BankID
	bankAccount.AccountNumber = req.AccountNumber
	bankAccount.AccountHolderName = req.AccountHolderName

	if err := c.DB.Save(&bankAccount).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bank account"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bank account updated successfully"})
}

// DeleteBankAccount deletes a bank account
func (c *BankController) DeleteBankAccount(ctx *gin.Context) {
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

	// Check if bank account exists and belongs to user
	var bankAccount models.BankAccount
	if err := c.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&bankAccount).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Bank account not found"})
		return
	}

	if err := c.DB.Delete(&bankAccount).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bank account"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bank account deleted successfully"})
}

// FetchTransactions fetches transactions for a specific bank account
func (c *BankController) FetchTransactions(ctx *gin.Context) {
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

	// Check if bank account exists and belongs to user
	var bankAccount models.BankAccount
	if err := c.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&bankAccount).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Bank account not found"})
		return
	}

	// Generate mock transactions (20 transactions for demo)
	mockTransactions, err := c.TransactionService.GenerateMockTransactions(uint(accountID), userID, 20)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate transactions"})
		return
	}

	// Store transactions in database
	if err := c.TransactionService.StoreTransactions(mockTransactions, uint(accountID), userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store transactions"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Transactions fetched and stored successfully",
		"accountId":    accountID,
		"count":        len(mockTransactions),
		"transactions": mockTransactions,
	})
}

// Helper functions
func getBankName(bankID string) string {
	bankNames := map[string]string{
		"hdfc":   "HDFC Bank",
		"icici":  "ICICI Bank",
		"sbi":    "State Bank of India",
		"axis":   "Axis Bank",
		"kotak":  "Kotak Mahindra Bank",
		"yes":    "Yes Bank",
		"pnb":    "Punjab National Bank",
		"canara": "Canara Bank",
		"MANUAL": "Manual Entry",
	}

	if name, exists := bankNames[bankID]; exists {
		return name
	}
	return "Unknown Bank"
}

func maskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= 4 {
		return accountNumber
	}
	return "****" + accountNumber[len(accountNumber)-4:]
}

func maskMobileNumber(mobileNumber string) string {
	if len(mobileNumber) <= 4 {
		return mobileNumber
	}
	return "****" + mobileNumber[len(mobileNumber)-4:]
}

func getUserIDFromContext(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}

	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}
