package routes

import (
	"fmt"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/your-github/expense-tracker-backend/config"
	"github.com/your-github/expense-tracker-backend/controllers"
	"github.com/your-github/expense-tracker-backend/middleware"
	"github.com/your-github/expense-tracker-backend/services"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// Set Gin to release mode for better performance
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Use optimized middleware stack
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom logging format for better performance
		return fmt.Sprintf("%s | %d | %s | %s | %s | %d | %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Method,
			param.Path,
			param.ClientIP,
			param.Latency.Milliseconds(),
			param.ErrorMessage,
		)
	}))

	// Enable Gzip compression for better performance
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// Apply security middleware with optimized settings
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.InputValidation())
	r.Use(middleware.RequestTimeout(15 * time.Second)) // Reduced timeout for better responsiveness
	r.Use(middleware.RateLimit(rate.Limit(200), 500))  // Increased rate limits for better scalability
	r.Use(middleware.CORSSecurity())

	// Initialize optimized services with enhanced caching
	authSvc := services.NewAuthService(db, cfg)
	expSvc := services.NewExpenseService(db)
	sumSvc := services.NewSummaryService(db)

	// Initialize controllers
	authCtl := &controllers.AuthController{S: authSvc, Config: cfg}
	expCtl := &controllers.ExpenseController{S: expSvc}
	sumCtl := &controllers.SummaryController{S: sumSvc}
	profCtl := &controllers.ProfileController{}
	aiCtl := &controllers.AIController{}
	// Initialize bank verification service
	bankVerificationSvc := services.NewBankVerificationService(
		cfg.BankVerification.APIKey,
		cfg.BankVerification.APIURL,
		"mock", // Use "mock" for development, change to "karza", "signzy", or "razorpay" for production
	)

	// Initialize services
	transactionSvc := services.NewTransactionService(db)

	bankCtl := &controllers.BankController{
		DB:                  db,
		VerificationService: bankVerificationSvc,
		TransactionService:  transactionSvc,
	}
	txnCtl := &controllers.TransactionController{
		TransactionService: transactionSvc,
	}

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "timestamp": time.Now().Unix()})
	})

	// Migration endpoint (no auth required for testing)
	r.POST("/api/migrate-expenses", expCtl.MigrateAllExpensesToTransactions)

	// Public routes (no authentication required)
	r.GET("/api/categories", func(c *gin.Context) {
		categories := []string{"Food", "Transport", "Shopping", "Entertainment", "Bills", "Income", "Other"}
		c.JSON(200, categories)
	})

	r.GET("/api/analytics", func(c *gin.Context) {
		c.JSON(200, gin.H{"analytics": "available"})
	})

	// Public bank information
	r.GET("/api/banks", func(c *gin.Context) {
		banks := []map[string]interface{}{
			// Public Sector Banks
			{"id": "sbi", "name": "State Bank of India", "logo": "sbi.png", "type": "public"},
			{"id": "pnb", "name": "Punjab National Bank", "logo": "pnb.png", "type": "public"},
			{"id": "canara", "name": "Canara Bank", "logo": "canara.png", "type": "public"},
			{"id": "bank-of-baroda", "name": "Bank of Baroda", "logo": "bob.png", "type": "public"},
			{"id": "union-bank", "name": "Union Bank of India", "logo": "union.png", "type": "public"},
			{"id": "bank-of-india", "name": "Bank of India", "logo": "boi.png", "type": "public"},
			{"id": "central-bank", "name": "Central Bank of India", "logo": "cbi.png", "type": "public"},
			{"id": "indian-bank", "name": "Indian Bank", "logo": "indian.png", "type": "public"},
			{"id": "uco-bank", "name": "UCO Bank", "logo": "uco.png", "type": "public"},
			{"id": "bank-of-maharashtra", "name": "Bank of Maharashtra", "logo": "bom.png", "type": "public"},
			{"id": "punjab-sind-bank", "name": "Punjab & Sind Bank", "logo": "psb.png", "type": "public"},

			// Private Sector Banks
			{"id": "hdfc", "name": "HDFC Bank", "logo": "hdfc.png", "type": "private"},
			{"id": "icici", "name": "ICICI Bank", "logo": "icici.png", "type": "private"},
			{"id": "axis", "name": "Axis Bank", "logo": "axis.png", "type": "private"},
			{"id": "kotak", "name": "Kotak Mahindra Bank", "logo": "kotak.png", "type": "private"},
			{"id": "yes", "name": "Yes Bank", "logo": "yes.png", "type": "private"},
			{"id": "indusind", "name": "IndusInd Bank", "logo": "indusind.png", "type": "private"},
			{"id": "idfc", "name": "IDFC First Bank", "logo": "idfc.png", "type": "private"},
			{"id": "bandhan", "name": "Bandhan Bank", "logo": "bandhan.png", "type": "private"},
			{"id": "csb", "name": "CSB Bank", "logo": "csb.png", "type": "private"},
			{"id": "dcb", "name": "DCB Bank", "logo": "dcb.png", "type": "private"},
			{"id": "federal", "name": "Federal Bank", "logo": "federal.png", "type": "private"},
			{"id": "karnataka", "name": "Karnataka Bank", "logo": "karnataka.png", "type": "private"},
			{"id": "karur-vysya", "name": "Karur Vysya Bank", "logo": "kvb.png", "type": "private"},
			{"id": "nainital", "name": "Nainital Bank", "logo": "nainital.png", "type": "private"},
			{"id": "rbl", "name": "RBL Bank", "logo": "rbl.png", "type": "private"},
			{"id": "south-indian", "name": "South Indian Bank", "logo": "sib.png", "type": "private"},
			{"id": "tamilnad", "name": "Tamilnad Mercantile Bank", "logo": "tmb.png", "type": "private"},

			// Foreign Banks
			{"id": "citibank", "name": "Citibank", "logo": "citi.png", "type": "foreign"},
			{"id": "hsbc", "name": "HSBC Bank", "logo": "hsbc.png", "type": "foreign"},
			{"id": "standard-chartered", "name": "Standard Chartered Bank", "logo": "scb.png", "type": "foreign"},
			{"id": "deutsche", "name": "Deutsche Bank", "logo": "deutsche.png", "type": "foreign"},
			{"id": "barclays", "name": "Barclays Bank", "logo": "barclays.png", "type": "foreign"},
			{"id": "dbs", "name": "DBS Bank", "logo": "dbs.png", "type": "foreign"},
			{"id": "rbs", "name": "Royal Bank of Scotland", "logo": "rbs.png", "type": "foreign"},
			{"id": "bnp-paribas", "name": "BNP Paribas", "logo": "bnp.png", "type": "foreign"},
			{"id": "societe-generale", "name": "Societe Generale", "logo": "sg.png", "type": "foreign"},

			// Regional Rural Banks
			{"id": "andhra-pradesh-grameena", "name": "Andhra Pradesh Grameena Vikas Bank", "logo": "apgvb.png", "type": "rrb"},
			{"id": "karnataka-gramin", "name": "Karnataka Gramin Bank", "logo": "kgb.png", "type": "rrb"},
			{"id": "madhya-pradesh-gramin", "name": "Madhya Pradesh Gramin Bank", "logo": "mpgb.png", "type": "rrb"},
			{"id": "rajasthan-marudhara", "name": "Rajasthan Marudhara Gramin Bank", "logo": "rmgb.png", "type": "rrb"},
			{"id": "uttar-bihar-gramin", "name": "Uttar Bihar Gramin Bank", "logo": "ubgb.png", "type": "rrb"},

			// Small Finance Banks
			{"id": "au-small-finance", "name": "AU Small Finance Bank", "logo": "au.png", "type": "sfb"},
			{"id": "equitas-small-finance", "name": "Equitas Small Finance Bank", "logo": "equitas.png", "type": "sfb"},
			{"id": "fino-payments", "name": "Fino Payments Bank", "logo": "fino.png", "type": "sfb"},
			{"id": "jammu-kashmir", "name": "Jammu & Kashmir Bank", "logo": "jkb.png", "type": "sfb"},
			{"id": "karnataka-vikas", "name": "Karnataka Vikas Grameena Bank", "logo": "kvg.png", "type": "sfb"},
			{"id": "maharashtra-gramin", "name": "Maharashtra Gramin Bank", "logo": "mgb.png", "type": "sfb"},
			{"id": "odisha-gramya", "name": "Odisha Gramya Bank", "logo": "ogb.png", "type": "sfb"},
			{"id": "puduvai-bharathiar", "name": "Puduvai Bharathiar Grama Bank", "logo": "pbg.png", "type": "sfb"},
			{"id": "saurashtra-gramin", "name": "Saurashtra Gramin Bank", "logo": "sgb.png", "type": "sfb"},
			{"id": "tamil-nadu-grama", "name": "Tamil Nadu Grama Bank", "logo": "tngb.png", "type": "sfb"},
			{"id": "telangana-gramin", "name": "Telangana Grameena Bank", "logo": "tgb.png", "type": "sfb"},
			{"id": "uttar-pradesh-gramin", "name": "Uttar Pradesh Gramin Bank", "logo": "upgb.png", "type": "sfb"},
			{"id": "uttarakhand-gramin", "name": "Uttarakhand Gramin Bank", "logo": "ukgb.png", "type": "sfb"},
			{"id": "west-bengal-gramin", "name": "West Bengal Gramin Bank", "logo": "wbgb.png", "type": "sfb"},
		}
		c.JSON(200, banks)
	})

	// Auth routes
	r.POST("/api/register", authCtl.Register)
	r.POST("/api/signup", authCtl.Register) // Alias for register to match frontend
	r.POST("/api/login", authCtl.Login)
	r.POST("/api/verify-otp", authCtl.VerifyOTP)
	r.POST("/api/resend-otp", authCtl.ResendOTP)

	// Protected routes (authentication required)
	protected := r.Group("/api")
	protected.Use(middleware.Auth(cfg.JWT.Secret))
	{
		// Profile routes
		protected.GET("/profile", profCtl.Get)
		protected.DELETE("/user", profCtl.Delete)

		// Expense routes with optimized endpoints
		protected.POST("/expenses", expCtl.Create)
		protected.GET("/expenses", expCtl.List)
		protected.PUT("/expenses/:id", expCtl.Update)
		protected.DELETE("/expenses/:id", expCtl.Delete)
		protected.POST("/expenses/recategorize", expCtl.Recategorize)
		protected.GET("/expenses/range", expCtl.GetByDateRange)
		protected.GET("/expenses/category/:category", expCtl.GetByCategory)
		protected.POST("/expenses/migrate", expCtl.MigrateExpensesToTransactions)

		// Summary routes
		protected.GET("/summary", sumCtl.Get)
		protected.GET("/summary/lifetime", sumCtl.GetLifetime)
		protected.GET("/summary/category-breakdown", sumCtl.GetCategoryBreakdown)

		// AI insights route
		protected.GET("/ai-insights", aiCtl.GetAIInsights)

		// Bank account management routes
		protected.GET("/bank-accounts", bankCtl.GetBankAccounts)
		protected.POST("/bank-accounts", bankCtl.AddBankAccount)
		protected.PUT("/bank-accounts/:id", bankCtl.UpdateBankAccount)
		protected.DELETE("/bank-accounts/:id", bankCtl.DeleteBankAccount)
		protected.POST("/bank-accounts/:id/fetch", bankCtl.FetchTransactions)

		// Transaction history routes
		protected.GET("/transactions", txnCtl.GetTransactionHistory)
		protected.GET("/transactions/bank-account/:id", txnCtl.GetTransactionsByBankAccount)
	}

	return r
}
