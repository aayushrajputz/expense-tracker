package app

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"expense-tracker-backend/config"
	"expense-tracker-backend/controllers"
	"expense-tracker-backend/internal/core/services"
	"expense-tracker-backend/internal/http/handlers"
	"expense-tracker-backend/internal/http/middleware"
	"expense-tracker-backend/internal/repo"
	expenseServices "expense-tracker-backend/services"
)

// App represents the main application
type App struct {
	config       *config.Config
	repositories *repo.Repositories
	aaService    *services.AAService
	router       *gin.Engine
	server       *http.Server
	cron         *cron.Cron
	logger       *zap.Logger
	db           *gorm.DB
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config, repositories *repo.Repositories, db *gorm.DB, logger *zap.Logger) *App {
	logger.Info("Initializing application...")
	// Initialize services
	normalizer := services.NewNormalizer()
	deduplicator := services.NewDeduplicator()

	// Initialize AA client (mock for now)
	aaClient := services.NewMockAAClient()

	// Initialize AA service
	aaService := services.NewAAService(aaClient, repositories, normalizer, deduplicator, logger)

	// Initialize expense service and controller
	expenseService := expenseServices.NewExpenseService(db)
	expenseController := controllers.NewExpenseController(expenseService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(repositories, cfg)
	aaHandler := handlers.NewAAHandler(aaService, repositories, cfg, logger)

	// Setup router
	logger.Info("Setting up router...")
	router := setupRouter(cfg, authHandler, aaHandler, expenseController, logger)

	// Setup cron jobs
	logger.Info("Setting up cron jobs...")
	cronJobs := setupCronJobs(aaService, logger)

	app := &App{
		config:       cfg,
		repositories: repositories,
		aaService:    aaService,
		router:       router,
		cron:         cronJobs,
		logger:       logger,
		db:           db,
	}

	logger.Info("Application initialized successfully")
	return app
}

// Start starts the application
func (a *App) Start() error {
	// Start cron jobs
	a.cron.Start()
	a.logger.Info("Cron jobs started")

	// Get port from environment variable or config
	port := os.Getenv("PORT")
	if port == "" {
		port = a.config.App.Port
	}
	if port == "" {
		port = "8080" // Final fallback
	}

	// Create HTTP server
	a.server = &http.Server{
		Addr:         ":" + port,
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	a.logger.Info("Starting server", zap.String("port", port), zap.String("address", ":"+port))
	return a.server.ListenAndServe()
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down application...")

	// Stop cron jobs
	if a.cron != nil {
		ctx := a.cron.Stop()
		<-ctx.Done()
	}

	// Shutdown HTTP server
	if a.server != nil {
		return a.server.Shutdown(ctx)
	}

	return nil
}

// setupRouter configures the HTTP router
func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, aaHandler *handlers.AAHandler, expenseController *controllers.ExpenseController, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// Simple ping endpoint for basic connectivity
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Database test endpoint
	router.GET("/api/db-test", func(c *gin.Context) {
		// Simple database connectivity test
		c.JSON(200, gin.H{"message": "Database test endpoint available"})
	})

	// Database status endpoint
	router.GET("/api/db-status", func(c *gin.Context) {
		// Check database tables and status
		c.JSON(200, gin.H{
			"message": "Database status endpoint",
			"backend_url": "https://expense-tracker-tzun.onrender.com",
			"health_endpoint": "/health",
			"registration_endpoint": "/api/register",
			"login_endpoint": "/api/login",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public) - match frontend expectations
		api.POST("/register", authHandler.Register)
		api.POST("/signup", authHandler.Register) // Alias for register
		api.POST("/login", authHandler.Login)

		// Categories endpoint (public)
		api.GET("/categories", func(c *gin.Context) {
			categories := []string{
				"Food",
				"Transportation",
				"Entertainment",
				"Shopping",
				"Bills",
				"Healthcare",
				"Education",
				"Travel",
				"Income",
				"Other",
			}
			c.JSON(200, categories)
		})

		// Analytics endpoint (public for now)
		api.GET("/analytics", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"insights": []gin.H{
					{
						"type":    "trend",
						"title":   "Spending Trend",
						"content": "Your spending patterns are being analyzed.",
					},
					{
						"type":    "advice",
						"title":   "Budget Recommendation",
						"content": "Consider tracking your expenses for better insights.",
					},
				},
			})
		})

		// Expenses endpoint (protected) - using real controller
		expenses := api.Group("/expenses")
		expenses.Use(middleware.Auth(cfg.JWT.Secret))
		expenses.Use(func(c *gin.Context) {
			// Convert user_id string to userID uint for expense controller compatibility
			if _, exists := c.Get("user_id"); exists {
				// For now, use a simple hash of the user_id string to create a uint
				// In a real app, you'd want to look up the actual user ID from the database
				userIDHash := uint(1) // Default to user ID 1 for demo purposes
				c.Set("userID", userIDHash)
			}
			c.Next()
		})
		{
			expenses.GET("", expenseController.List)
			expenses.POST("", expenseController.Create)
			expenses.PUT("/:id", expenseController.Update)
			expenses.DELETE("/:id", expenseController.Delete)
		}

		// AA routes
		aa := api.Group("/aa")
		{
			// Public webhooks
			aa.POST("/consents/callback", aaHandler.ConsentCallback)
			aa.POST("/webhook", aaHandler.DataReadyWebhook)

			// Protected routes
			aaProtected := aa.Group("")
			aaProtected.Use(middleware.Auth(cfg.JWT.Secret))
			{
				aaProtected.POST("/consents/initiate", aaHandler.InitiateConsent)
				aaProtected.POST("/fetch", aaHandler.FetchTransactions)
				aaProtected.GET("/bank-links", aaHandler.GetBankLinks)
				aaProtected.POST("/consents/revoke", aaHandler.RevokeConsent)
			}
		}

		// User routes (protected)
		me := api.Group("/me")
		me.Use(middleware.Auth(cfg.JWT.Secret))
		{
			me.GET("/transactions", func(c *gin.Context) {
				// Get user ID from context
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}

				// Return empty transactions array for now (since we don't have real transaction data)
				c.JSON(200, []gin.H{})
			})
			me.GET("/summary", func(c *gin.Context) {
				// Get user ID from context
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}

				// Get userID for expense queries
				userID := uint(1) // Using the same user ID as in expense controller

				// Get all expenses for the user
				expenses, err := expenseController.S.List(userID)
				if err != nil {
					c.JSON(500, gin.H{"error": "Failed to fetch expenses"})
					return
				}

				// Calculate real summary data
				var totalIncome, totalExpenses float64
				var transactionCount int
				categoryBreakdown := make(map[string]float64)
				var recentTransactions []gin.H

				for _, expense := range expenses {
					transactionCount++
					if expense.Type == "income" {
						totalIncome += expense.Amount
					} else {
						totalExpenses += expense.Amount
					}

					// Category breakdown
					if expense.Category != "" {
						categoryBreakdown[expense.Category] += expense.Amount
					}

					// Add to recent transactions (limit to 5)
					if len(recentTransactions) < 5 {
						recentTransactions = append(recentTransactions, gin.H{
							"id":          expense.ID,
							"title":       expense.Title,
							"amount":      expense.Amount,
							"type":        expense.Type,
							"category":    expense.Category,
							"date":        expense.Date,
							"description": expense.Title,
						})
					}
				}

				netBalance := totalIncome - totalExpenses

				// Return real summary data
				summary := gin.H{
					"total_income":    totalIncome,
					"total_expenses":  totalExpenses,
					"net_balance":     netBalance,
					"transaction_count": transactionCount,
					"current_month": gin.H{
						"income":  totalIncome,
						"expenses": totalExpenses,
						"balance":  netBalance,
					},
					"category_breakdown": categoryBreakdown,
					"recent_transactions": recentTransactions,
				}

				c.JSON(200, summary)
			})
			me.POST("/categorize/override", func(c *gin.Context) {
				// TODO: Implement category override
				c.JSON(200, gin.H{"message": "Category override endpoint - to be implemented"})
			})
		}
	}

	return router
}

// setupCronJobs configures background cron jobs
func setupCronJobs(aaService *services.AAService, logger *zap.Logger) *cron.Cron {
	c := cron.New(cron.WithLocation(time.UTC))

	// Daily transaction fetch job (2:00 AM IST = 8:30 PM UTC)
	_, err := c.AddFunc("30 20 * * *", func() {
		logger.Info("Starting daily transaction fetch job")

		// TODO: Implement daily fetch logic using aaService
		// This would iterate through all active bank links and fetch recent transactions
		// For now, just log that we have access to the service
		_ = aaService // Use aaService parameter to avoid unused parameter warning
		// aaService.FetchDailyTransactions()

		logger.Info("Completed daily transaction fetch job")
	})

	if err != nil {
		logger.Error("Failed to schedule daily fetch job", zap.Error(err))
	}

	return c
}
