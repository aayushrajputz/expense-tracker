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

	"github.com/your-github/expense-tracker-backend/config"
	"github.com/your-github/expense-tracker-backend/internal/core/services"
	"github.com/your-github/expense-tracker-backend/internal/http/handlers"
	"github.com/your-github/expense-tracker-backend/internal/http/middleware"
	"github.com/your-github/expense-tracker-backend/internal/repo"
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
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config, repositories *repo.Repositories, logger *zap.Logger) *App {
	logger.Info("Initializing application...")
	// Initialize services
	normalizer := services.NewNormalizer()
	deduplicator := services.NewDeduplicator()

	// Initialize AA client (mock for now)
	aaClient := services.NewMockAAClient()

	// Initialize AA service
	aaService := services.NewAAService(aaClient, repositories, normalizer, deduplicator, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(repositories, cfg)
	aaHandler := handlers.NewAAHandler(aaService, repositories, cfg, logger)

	// Setup router
	logger.Info("Setting up router...")
	router := setupRouter(cfg, authHandler, aaHandler, logger)

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
func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, aaHandler *handlers.AAHandler, logger *zap.Logger) *gin.Engine {
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

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public) - match frontend expectations
		api.POST("/register", authHandler.Register)
		api.POST("/signup", authHandler.Register) // Alias for register
		api.POST("/login", authHandler.Login)

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
				// TODO: Implement transaction listing
				c.JSON(200, gin.H{"message": "Transactions endpoint - to be implemented"})
			})
			me.GET("/summary", func(c *gin.Context) {
				// TODO: Implement summary endpoint
				c.JSON(200, gin.H{"message": "Summary endpoint - to be implemented"})
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
