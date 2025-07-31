package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-github/expense-tracker-backend/database"
	"github.com/your-github/expense-tracker-backend/models"
)

type ProfileController struct{}

func (c *ProfileController) Get(ctx *gin.Context) {
	uid := ctx.GetUint("userID")

	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's total transactions count
	var totalTransactions int64
	database.DB.Model(&models.Expense{}).Where("user_id = ?", uid).Count(&totalTransactions)

	// Get user's first transaction date
	var firstExpense models.Expense
	var memberSince time.Time
	if err := database.DB.Where("user_id = ?", uid).Order("date asc").First(&firstExpense).Error; err == nil {
		// Parse the date string to time.Time
		if parsedDate, err := time.Parse("2006-01-02", firstExpense.Date); err == nil {
			memberSince = parsedDate
		} else {
			memberSince = user.CreatedAt
		}
	} else {
		memberSince = user.CreatedAt
	}

	profileData := gin.H{
		"id":                 user.ID,
		"name":               user.Name,
		"email":              user.Email,
		"budget":             user.Budget,
		"created_at":         user.CreatedAt,
		"member_since":       memberSince,
		"total_transactions": totalTransactions,
		"account_status":     "Active",
	}

	ctx.JSON(http.StatusOK, profileData)
}

func (c *ProfileController) Delete(ctx *gin.Context) {
	uid := ctx.GetUint("userID")

	// Delete the user and cascade delete all related data
	if err := database.DB.Unscoped().Delete(&models.User{}, uid).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Account and all data deleted successfully"})
}
