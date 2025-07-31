package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/your-github/expense-tracker-backend/utils"
)

type Summary struct {
	TotalExpenses   float64            `json:"total_expenses"`
	TotalIncome     float64            `json:"total_income"`
	NetBalance      float64            `json:"net_balance"`
	TopCategories   map[string]float64 `json:"top_categories"`
	AverageDaily    float64            `json:"average_daily"`
	RemainingBudget float64            `json:"remaining_budget"`
}

type SummaryService struct {
	DB    *gorm.DB
	Cache *utils.LRUCache
}

// NewSummaryService creates a new summary service with caching
func NewSummaryService(db *gorm.DB) *SummaryService {
	cache := utils.NewLRUCache(500, 10*time.Minute) // Cache for 10 minutes
	cache.StartCleanup(5 * time.Minute)             // Cleanup every 5 minutes

	return &SummaryService{
		DB:    db,
		Cache: cache,
	}
}

func (s *SummaryService) Monthly(uid uint, budget float64, year int, month time.Month) (Summary, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("summary_monthly:%d:%d:%d:%f", uid, year, month, budget)
	if cached, found := s.Cache.Get(cacheKey); found {
		if summary, ok := cached.(Summary); ok {
			return summary, nil
		}
	}

	var sum Summary
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	// Convert to string format to match database date field
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use efficient single query with aggregation for better performance
	type AggResult struct {
		Type     string  `json:"type"`
		Total    float64 `json:"total"`
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
	}

	var results []AggResult

	// Single query to get all aggregated data
	err := s.DB.WithContext(ctx).Raw(`
		SELECT 
			type,
			SUM(amount) as total,
			category,
			amount
		FROM expenses 
		WHERE user_id = ? AND date >= ? AND date < ?
		GROUP BY type, category, amount
		ORDER BY type, total DESC
	`, uid, startStr, endStr).Scan(&results).Error

	if err != nil {
		return sum, err
	}

	// Process results efficiently
	sum.TopCategories = make(map[string]float64)
	for _, result := range results {
		switch result.Type {
		case "expense":
			sum.TotalExpenses += result.Total
			// Track top categories for expenses only
			if result.Category != "" {
				sum.TopCategories[result.Category] += result.Amount
			}
		case "income":
			sum.TotalIncome += result.Total
		}
	}

	// Calculate net balance
	sum.NetBalance = sum.TotalIncome - sum.TotalExpenses

	// Get top 3 categories efficiently
	topCategories := make([]struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
	}, 0, 3)

	err = s.DB.WithContext(ctx).Raw(`
		SELECT category, SUM(amount) as total
		FROM expenses 
		WHERE user_id = ? AND date >= ? AND date < ? AND type = 'expense'
		GROUP BY category 
		ORDER BY total DESC 
		LIMIT 3
	`, uid, startStr, endStr).Scan(&topCategories).Error

	if err == nil {
		sum.TopCategories = make(map[string]float64)
		for _, cat := range topCategories {
			if cat.Category != "" {
				sum.TopCategories[cat.Category] = cat.Total
			}
		}
	} else {
		sum.TopCategories = make(map[string]float64)
	}

	days := time.Now().Day()
	if days > 0 {
		sum.AverageDaily = sum.TotalExpenses / float64(days)
	}

	// Calculate remaining budget (budget - expenses + income)
	sum.RemainingBudget = budget - sum.TotalExpenses + sum.TotalIncome

	// Cache the result
	s.Cache.Set(cacheKey, sum)

	return sum, nil
}

// Lifetime totals for profile page
func (s *SummaryService) Lifetime(uid uint) (Summary, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("summary_lifetime:%d", uid)
	if cached, found := s.Cache.Get(cacheKey); found {
		if summary, ok := cached.(Summary); ok {
			return summary, nil
		}
	}

	var sum Summary

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use efficient single query with aggregation
	type AggResult struct {
		Type     string  `json:"type"`
		Total    float64 `json:"total"`
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
	}

	var results []AggResult

	// Single query to get all aggregated data
	err := s.DB.WithContext(ctx).Raw(`
		SELECT 
			type,
			SUM(amount) as total,
			category,
			amount
		FROM expenses 
		WHERE user_id = ?
		GROUP BY type, category, amount
		ORDER BY type, total DESC
	`, uid).Scan(&results).Error

	if err != nil {
		return sum, err
	}

	// Process results efficiently
	sum.TopCategories = make(map[string]float64)
	for _, result := range results {
		switch result.Type {
		case "expense":
			sum.TotalExpenses += result.Total
			// Track top categories for expenses only
			if result.Category != "" {
				sum.TopCategories[result.Category] += result.Amount
			}
		case "income":
			sum.TotalIncome += result.Total
		}
	}

	// Calculate net balance
	sum.NetBalance = sum.TotalIncome - sum.TotalExpenses

	// Get top 5 categories efficiently
	topCategories := make([]struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
	}, 0, 5)

	err = s.DB.WithContext(ctx).Raw(`
		SELECT category, SUM(amount) as total
		FROM expenses 
		WHERE user_id = ? AND type = 'expense'
		GROUP BY category 
		ORDER BY total DESC 
		LIMIT 5
	`, uid).Scan(&topCategories).Error

	if err == nil {
		sum.TopCategories = make(map[string]float64)
		for _, cat := range topCategories {
			if cat.Category != "" {
				sum.TopCategories[cat.Category] = cat.Total
			}
		}
	} else {
		sum.TopCategories = make(map[string]float64)
	}

	// Cache the result
	s.Cache.Set(cacheKey, sum)

	return sum, nil
}

// GetCategoryBreakdown efficiently gets expense breakdown by category
func (s *SummaryService) GetCategoryBreakdown(uid uint, startDate, endDate string) (map[string]float64, error) {
	cacheKey := fmt.Sprintf("category_breakdown:%d:%s:%s", uid, startDate, endDate)
	if cached, found := s.Cache.Get(cacheKey); found {
		if breakdown, ok := cached.(map[string]float64); ok {
			return breakdown, nil
		}
	}

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var results []struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
	}

	err := s.DB.WithContext(ctx).Raw(`
		SELECT category, SUM(amount) as total
		FROM expenses 
		WHERE user_id = ? AND date >= ? AND date <= ? AND type = 'expense'
		GROUP BY category 
		ORDER BY total DESC
	`, uid, startDate, endDate).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	breakdown := make(map[string]float64)
	for _, result := range results {
		if result.Category != "" {
			breakdown[result.Category] = result.Total
		}
	}

	// Cache the result
	s.Cache.Set(cacheKey, breakdown)

	return breakdown, nil
}

// InvalidateUserCache removes cache entries for a specific user
func (s *SummaryService) InvalidateUserCache(uid uint) {
	// Clear all cache entries for this user
	// TODO: Implement selective cache invalidation based on user ID
	// For now, clear all cache when user data changes for simplicity
	_ = uid // Use uid parameter (will be used in future implementation)
	s.Cache.Clear()
}
