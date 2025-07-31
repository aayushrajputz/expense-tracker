package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-github/expense-tracker-backend/models"
	"github.com/your-github/expense-tracker-backend/services"
)

type ExpenseController struct{ S *services.ExpenseService }

// NewExpenseController creates a new expense controller with optimized service
func NewExpenseController(expenseService *services.ExpenseService) *ExpenseController {
	return &ExpenseController{S: expenseService}
}

func (c *ExpenseController) Create(ctx *gin.Context) {
	var in models.Expense
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid := ctx.GetUint("userID")
	if err := c.S.Create(&in, uid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, in)
}

func (c *ExpenseController) List(ctx *gin.Context) {
	uid := ctx.GetUint("userID")
	limitStr := ctx.DefaultQuery("limit", "")

	var limit int
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	data, _ := c.S.List(uid, limit)
	ctx.JSON(http.StatusOK, data)
}

func (c *ExpenseController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var in models.Expense
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.S.Update(uint(id), ctx.GetUint("userID"), &in); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (c *ExpenseController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.S.Delete(uint(id), ctx.GetUint("userID")); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (c *ExpenseController) Recategorize(ctx *gin.Context) {
	uid := ctx.GetUint("userID")
	if err := c.S.RecategorizeAll(uid); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Transactions recategorized successfully"})
}

// GetByDateRange efficiently retrieves expenses within a date range
func (c *ExpenseController) GetByDateRange(ctx *gin.Context) {
	uid := ctx.GetUint("userID")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	if startDate == "" || endDate == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	expenses, err := c.S.GetExpensesByDateRange(uid, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, expenses)
}

// GetByCategory efficiently retrieves expenses by category
func (c *ExpenseController) GetByCategory(ctx *gin.Context) {
	uid := ctx.GetUint("userID")
	category := ctx.Param("category")

	if category == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "category is required"})
		return
	}

	expenses, err := c.S.GetExpensesByCategory(uid, category)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, expenses)
}

// MigrateExpensesToTransactions migrates existing expenses to transaction records
func (c *ExpenseController) MigrateExpensesToTransactions(ctx *gin.Context) {
	if err := c.S.MigrateExistingExpensesToTransactions(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Migration completed successfully"})
}

// MigrateAllExpensesToTransactions migrates ALL expenses to transaction records (no auth required for testing)
func (c *ExpenseController) MigrateAllExpensesToTransactions(ctx *gin.Context) {
	if err := c.S.MigrateExistingExpensesToTransactions(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Migration completed successfully"})
}
