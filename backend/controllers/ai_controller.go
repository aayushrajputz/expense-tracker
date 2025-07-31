package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-github/expense-tracker-backend/database"
	"github.com/your-github/expense-tracker-backend/models"
	"github.com/your-github/expense-tracker-backend/utils"
)

type AIController struct{}

// GetAIInsights generates AI-powered insights using OpenAI GPT-3.5
func (c *AIController) GetAIInsights(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	// Get user's expenses and income
	var expenses []models.Expense
	database.DB.Where("user_id = ?", userID).Order("date desc, created_at desc").Find(&expenses)

	if len(expenses) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"insights": []gin.H{
				{
					"type":        "info",
					"title":       "No Data Available",
					"description": "Add some transactions to get AI-powered insights.",
					"suggestion":  "Start by adding your income and expenses to receive personalized financial advice.",
				},
			},
		})
		return
	}

	// Calculate financial data
	financialData := c.calculateFinancialData(expenses)

	// Generate AI insights
	insights, err := utils.GenerateAIInsights(financialData)
	if err != nil {
		// If AI fails, return fallback insights
		ctx.JSON(http.StatusOK, gin.H{
			"insights": c.generateFallbackInsights(financialData),
			"ai_error": err.Error(),
		})
		return
	}

	// Convert insights to the expected format
	var responseInsights []gin.H
	for _, insight := range insights {
		responseInsights = append(responseInsights, gin.H{
			"type":        insight.Type,
			"title":       insight.Title,
			"description": insight.Description,
			"suggestion":  insight.Suggestion,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"insights":     responseInsights,
		"ai_generated": true,
	})
}

// calculateFinancialData prepares financial data for AI analysis
func (c *AIController) calculateFinancialData(expenses []models.Expense) utils.FinancialData {
	var totalIncome, totalExpenses float64
	categorySpending := make(map[string]float64)
	monthlyData := make(map[string]utils.MonthlyData)

	// Calculate totals and category spending
	for _, expense := range expenses {
		if expense.Type == "income" {
			totalIncome += expense.Amount
		} else {
			totalExpenses += expense.Amount
			categorySpending[expense.Category] += expense.Amount
		}

		// Group by month
		date, _ := time.Parse("2006-01-02", expense.Date)
		monthKey := date.Format("2006-01")

		if monthData, exists := monthlyData[monthKey]; exists {
			if expense.Type == "income" {
				monthData.Income += expense.Amount
			} else {
				monthData.Expenses += expense.Amount
			}
			monthData.Savings = monthData.Income - monthData.Expenses
			monthlyData[monthKey] = monthData
		} else {
			monthData := utils.MonthlyData{
				Month: date.Format("January 2006"),
			}
			if expense.Type == "income" {
				monthData.Income = expense.Amount
			} else {
				monthData.Expenses = expense.Amount
			}
			monthData.Savings = monthData.Income - monthData.Expenses
			monthlyData[monthKey] = monthData
		}
	}

	// Calculate savings rate
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = ((totalIncome - totalExpenses) / totalIncome) * 100
	}

	// Convert monthly data to slice
	var monthlyTrends []utils.MonthlyData
	for _, monthData := range monthlyData {
		monthlyTrends = append(monthlyTrends, monthData)
	}

	// Get recent transactions (last 5)
	var recentTransactions []utils.Transaction
	for i, expense := range expenses {
		if i >= 5 {
			break
		}
		recentTransactions = append(recentTransactions, utils.Transaction{
			Title:    expense.Title,
			Amount:   expense.Amount,
			Type:     expense.Type,
			Category: expense.Category,
			Date:     expense.Date,
		})
	}

	return utils.FinancialData{
		TotalIncome:        totalIncome,
		TotalExpenses:      totalExpenses,
		SavingsRate:        savingsRate,
		CategorySpending:   categorySpending,
		MonthlyTrends:      monthlyTrends,
		RecentTransactions: recentTransactions,
	}
}

// generateFallbackInsights creates basic insights when AI is not available
func (c *AIController) generateFallbackInsights(data utils.FinancialData) []gin.H {
	var insights []gin.H

	// Savings rate insight
	if data.SavingsRate >= 20 {
		insights = append(insights, gin.H{
			"type":        "success",
			"title":       "Excellent Savings Rate",
			"description": "Your savings rate is " + fmt.Sprintf("%.1f", data.SavingsRate) + "%.",
			"suggestion":  "Keep up the great work! Consider investing your savings for long-term growth.",
		})
	} else {
		insights = append(insights, gin.H{
			"type":        "warning",
			"title":       "Savings Rate Alert",
			"description": "Your savings rate is " + fmt.Sprintf("%.1f", data.SavingsRate) + "%.",
			"suggestion":  "Consider reducing expenses or increasing income to improve your savings rate.",
		})
	}

	// Category spending insight
	if len(data.CategorySpending) > 0 {
		var topCategory string
		var topAmount float64
		for category, amount := range data.CategorySpending {
			if amount > topAmount {
				topAmount = amount
				topCategory = category
			}
		}

		percentage := (topAmount / data.TotalExpenses) * 100
		if percentage > 40 {
			insights = append(insights, gin.H{
				"type":        "warning",
				"title":       "High Category Concentration",
				"description": topCategory + " accounts for " + fmt.Sprintf("%.1f", percentage) + "% of your spending.",
				"suggestion":  "Consider diversifying your spending across different categories.",
			})
		}
	}

	// Basic financial health insight
	if data.TotalIncome > 0 {
		expenseRatio := (data.TotalExpenses / data.TotalIncome) * 100
		if expenseRatio < 80 {
			insights = append(insights, gin.H{
				"type":        "success",
				"title":       "Good Financial Health",
				"description": "Your expenses are " + fmt.Sprintf("%.1f", expenseRatio) + "% of your income.",
				"suggestion":  "You have a healthy balance between income and expenses.",
			})
		}
	}

	return insights
}
