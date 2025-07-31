package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-github/expense-tracker-backend/database"
	"github.com/your-github/expense-tracker-backend/models"
	"github.com/your-github/expense-tracker-backend/services"
)

type SummaryController struct{ S *services.SummaryService }

// NewSummaryController creates a new summary controller with optimized service
func NewSummaryController(summaryService *services.SummaryService) *SummaryController {
	return &SummaryController{S: summaryService}
}

func (c *SummaryController) Get(ctx *gin.Context) {
	now := time.Now()
	uid := ctx.GetUint("userID")
	budget := ctx.DefaultQuery("budget", "")
	if budget == "" {
		var user models.User
		database.DB.First(&user, uid)
		budget = fmt.Sprintf("%f", user.Budget)
	}
	// parse budget
	bud, _ := strconv.ParseFloat(budget, 64)
	summary, _ := c.S.Monthly(uid, bud, now.Year(), now.Month())
	ctx.JSON(http.StatusOK, summary)
}

func (c *SummaryController) GetLifetime(ctx *gin.Context) {
	uid := ctx.GetUint("userID")
	summary, _ := c.S.Lifetime(uid)
	ctx.JSON(http.StatusOK, summary)
}

// GetCategoryBreakdown efficiently gets expense breakdown by category
func (c *SummaryController) GetCategoryBreakdown(ctx *gin.Context) {
	uid := ctx.GetUint("userID")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	if startDate == "" || endDate == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	breakdown, err := c.S.GetCategoryBreakdown(uid, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, breakdown)
}
