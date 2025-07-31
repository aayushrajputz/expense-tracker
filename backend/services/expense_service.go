package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/your-github/expense-tracker-backend/models"
	"github.com/your-github/expense-tracker-backend/utils"
)

type ExpenseService struct {
	DB    *gorm.DB
	Cache *utils.LRUCache
}

// NewExpenseService creates a new expense service with enhanced caching
func NewExpenseService(db *gorm.DB) *ExpenseService {
	// Increased cache size and optimized TTL for better performance
	cache := utils.NewLRUCache(5000, 30*time.Minute) // Larger cache, longer TTL
	cache.StartCleanup(10 * time.Minute)             // Less frequent cleanup

	return &ExpenseService{
		DB:    db,
		Cache: cache,
	}
}

func (s *ExpenseService) Create(e *models.Expense, uid uint) error {
	// Only auto-categorize expenses, not income
	if e.Type == "expense" && e.Category == "" {
		e.Category = utils.AutoCategory(e.Title)
	}
	e.UserID = uid

	// Use context with timeout for better performance
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start a transaction to ensure both expense and transaction are created
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Create the expense
	err := tx.Create(e).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Parse the date string to time.Time
	date, err := time.Parse("2006-01-02", e.Date)
	if err != nil {
		// If date parsing fails, use current time
		date = time.Now()
	}

	// Convert expense type to transaction type
	transactionType := "debit"
	if e.Type == "income" {
		transactionType = "credit"
	}

	// Create a corresponding transaction record
	transaction := models.Transaction{
		UserID:          uid,
		BankAccountID:   0, // No bank account for manual expenses
		TransactionID:   fmt.Sprintf("MANUAL_%d", e.ID),
		TransactionDate: date,
		Description:     e.Title,
		Amount:          e.Amount,
		Type:            transactionType,
		Category:        e.Category,
		Balance:         0, // No balance for manual expenses
		ReferenceNumber: "",
		MerchantName:    e.PaymentMethod,
		Location:        "Manual Entry",
		Status:          "completed",
	}

	err = tx.Create(&transaction).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err == nil {
		// Invalidate cache for this user with delay to prevent race conditions
		go func() {
			time.Sleep(100 * time.Millisecond)
			s.invalidateUserCache(uid)
		}()
	}
	return err
}

func (s *ExpenseService) Update(id uint, uid uint, in *models.Expense) error {
	var exp models.Expense

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.DB.WithContext(ctx).First(&exp, "id=? AND user_id=?", id, uid).Error; err != nil {
		return err
	}

	// Only auto-categorize expenses, not income
	if in.Type == "expense" && in.Category == "" {
		in.Category = utils.AutoCategory(in.Title)
	}

	// Start a transaction to ensure both expense and transaction are updated
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Update the expense
	err := tx.Model(&exp).Updates(in).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Parse the date string to time.Time
	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		// If date parsing fails, use current time
		date = time.Now()
	}

	// Convert expense type to transaction type
	transactionType := "debit"
	if in.Type == "income" {
		transactionType = "credit"
	}

	// Update the corresponding transaction record
	transactionID := fmt.Sprintf("MANUAL_%d", id)
	var transaction models.Transaction
	if err := tx.Where("transaction_id = ? AND user_id = ?", transactionID, uid).First(&transaction).Error; err != nil {
		// If transaction doesn't exist, create it
		transaction = models.Transaction{
			UserID:          uid,
			BankAccountID:   0,
			TransactionID:   transactionID,
			TransactionDate: date,
			Description:     in.Title,
			Amount:          in.Amount,
			Type:            transactionType,
			Category:        in.Category,
			Balance:         0,
			ReferenceNumber: "",
			MerchantName:    in.PaymentMethod,
			Location:        "Manual Entry",
			Status:          "completed",
		}
		err = tx.Create(&transaction).Error
	} else {
		// Update existing transaction
		err = tx.Model(&transaction).Updates(map[string]interface{}{
			"transaction_date": date,
			"description":      in.Title,
			"amount":           in.Amount,
			"type":             transactionType,
			"category":         in.Category,
			"merchant_name":    in.PaymentMethod,
		}).Error
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err == nil {
		// Invalidate cache for this user with delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			s.invalidateUserCache(uid)
		}()
	}
	return err
}

func (s *ExpenseService) Delete(id, uid uint) error {
	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start a transaction to ensure both expense and transaction are deleted
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Delete the expense
	err := tx.Delete(&models.Expense{}, "id=? AND user_id=?", id, uid).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the corresponding transaction record
	transactionID := fmt.Sprintf("MANUAL_%d", id)
	err = tx.Delete(&models.Transaction{}, "transaction_id=? AND user_id=?", transactionID, uid).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err == nil {
		// Invalidate cache for this user with delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			s.invalidateUserCache(uid)
		}()
	}
	return err
}

func (s *ExpenseService) List(uid uint, limit ...int) ([]models.Expense, error) {
	// Enhanced cache key with limit
	limitVal := 0
	if len(limit) > 0 {
		limitVal = limit[0]
	}
	cacheKey := fmt.Sprintf("expenses:%d:%d", uid, limitVal)

	// Try to get from cache first
	if cached, found := s.Cache.Get(cacheKey); found {
		if expenses, ok := cached.([]models.Expense); ok {
			return expenses, nil
		}
	}

	var ex []models.Expense
	query := s.DB.Where("user_id=?", uid).Order("date desc, created_at desc")

	if limitVal > 0 {
		query = query.Limit(limitVal)
	}

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := query.WithContext(ctx).Find(&ex).Error
	if err == nil {
		// Cache the result with longer TTL for frequently accessed data
		s.Cache.Set(cacheKey, ex)
	}
	return ex, err
}

func (s *ExpenseService) Get(id, uid uint) (models.Expense, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("expense:%d:%d", uid, id)
	if cached, found := s.Cache.Get(cacheKey); found {
		if expense, ok := cached.(models.Expense); ok {
			return expense, nil
		}
	}

	var e models.Expense

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.DB.WithContext(ctx).Where("id=? AND user_id=?", id, uid).First(&e).Error
	if err == nil {
		// Cache the result
		s.Cache.Set(cacheKey, e)
	}
	return e, err
}

func (s *ExpenseService) RecategorizeAll(uid uint) error {
	var expenses []models.Expense

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.DB.WithContext(ctx).Where("user_id = ? AND (category = '' OR category = 'Other')", uid).Find(&expenses).Error; err != nil {
		return err
	}

	// Use batch update for better performance
	updates := make([]map[string]interface{}, 0, len(expenses))
	for _, expense := range expenses {
		if expense.Type == "expense" {
			newCategory := utils.AutoCategory(expense.Title)
			if newCategory != "Other" {
				updates = append(updates, map[string]interface{}{
					"id":       expense.ID,
					"category": newCategory,
				})
			}
		}
	}

	// Batch update for better performance
	if len(updates) > 0 {
		err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			for _, update := range updates {
				if err := tx.Model(&models.Expense{}).Where("id = ?", update["id"]).Update("category", update["category"]).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err == nil {
			// Invalidate cache for this user
			s.invalidateUserCache(uid)
		}
		return err
	}

	return nil
}

// GetExpensesByDateRange efficiently retrieves expenses within a date range
func (s *ExpenseService) GetExpensesByDateRange(uid uint, startDate, endDate string) ([]models.Expense, error) {
	cacheKey := fmt.Sprintf("expenses_range:%d:%s:%s", uid, startDate, endDate)
	if cached, found := s.Cache.Get(cacheKey); found {
		if expenses, ok := cached.([]models.Expense); ok {
			return expenses, nil
		}
	}

	var expenses []models.Expense

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.DB.WithContext(ctx).
		Where("user_id = ? AND date >= ? AND date <= ?", uid, startDate, endDate).
		Order("date DESC, created_at DESC").
		Find(&expenses).Error

	if err == nil {
		s.Cache.Set(cacheKey, expenses)
	}

	return expenses, err
}

// GetExpensesByCategory efficiently retrieves expenses by category
func (s *ExpenseService) GetExpensesByCategory(uid uint, category string) ([]models.Expense, error) {
	cacheKey := fmt.Sprintf("expenses_category:%d:%s", uid, category)
	if cached, found := s.Cache.Get(cacheKey); found {
		if expenses, ok := cached.([]models.Expense); ok {
			return expenses, nil
		}
	}

	var expenses []models.Expense

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.DB.WithContext(ctx).
		Where("user_id = ? AND category = ?", uid, category).
		Order("date DESC, created_at DESC").
		Find(&expenses).Error

	if err == nil {
		s.Cache.Set(cacheKey, expenses)
	}

	return expenses, err
}

// invalidateUserCache removes all cache entries for a specific user
func (s *ExpenseService) invalidateUserCache(uid uint) {
	// Get all cache keys and remove those belonging to this user
	// This is a simplified approach - in production, you might want a more sophisticated cache invalidation strategy

	// Clear cache entries that contain this user ID
	// Note: This is a simplified approach. In a real implementation, you might want to maintain a separate index
	// TODO: Implement selective cache invalidation based on user ID
	// For now, clear all cache when user data changes for simplicity
	_ = uid // Use uid parameter (will be used in future implementation)
	s.Cache.Clear()
}

// MigrateExistingExpensesToTransactions creates transaction records for existing expenses
// This should be called once to migrate existing data
func (s *ExpenseService) MigrateExistingExpensesToTransactions() error {
	var expenses []models.Expense

	// Get all expenses that don't have corresponding transaction records
	err := s.DB.Find(&expenses).Error
	if err != nil {
		return err
	}

	for _, expense := range expenses {
		// Check if transaction already exists
		transactionID := fmt.Sprintf("MANUAL_%d", expense.ID)
		var existingTransaction models.Transaction
		if err := s.DB.Where("transaction_id = ?", transactionID).First(&existingTransaction).Error; err == nil {
			// Transaction already exists, skip
			continue
		}

		// Parse the date string to time.Time
		date, err := time.Parse("2006-01-02", expense.Date)
		if err != nil {
			// If date parsing fails, use current time
			date = time.Now()
		}

		// Convert expense type to transaction type
		transactionType := "debit"
		if expense.Type == "income" {
			transactionType = "credit"
		}

		// Create a corresponding transaction record
		transaction := models.Transaction{
			UserID:          expense.UserID,
			BankAccountID:   0, // No bank account for manual expenses
			TransactionID:   transactionID,
			TransactionDate: date,
			Description:     expense.Title,
			Amount:          expense.Amount,
			Type:            transactionType,
			Category:        expense.Category,
			Balance:         0, // No balance for manual expenses
			ReferenceNumber: "",
			MerchantName:    expense.PaymentMethod,
			Location:        "Manual Entry",
			Status:          "completed",
		}

		if err := s.DB.Create(&transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction for expense %d: %v", expense.ID, err)
		}
	}

	return nil
}
