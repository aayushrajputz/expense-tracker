package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/your-github/expense-tracker-backend/models"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		"localhost", "postgres", "aayush001", "expense_tracker", "5432")

	log.Printf("Connecting to database with DSN: %s", dsn)

	// Enhanced database configuration for performance
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Error), // Only log errors for better performance
		PrepareStmt:            true,                                 // Enable prepared statements
		SkipDefaultTransaction: true,                                 // Skip default transaction for better performance
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure connection pool for better performance
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// Optimized connection pool settings
	sqlDB.SetMaxIdleConns(10)                  // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // Maximum idle time of a connection

	log.Println("Database connected successfully with optimized settings")
	DB = db
}

func Migrate(db *gorm.DB) {
	log.Println("Running database migrations...")

	// Drop the problematic unique index if it exists
	log.Println("Dropping GoogleID unique index if exists...")
	db.Exec("DROP INDEX IF EXISTS idx_users_google_id")

	// Update existing empty GoogleID values to NULL
	log.Println("Updating empty GoogleID values to NULL...")
	db.Exec("UPDATE users SET google_id = NULL WHERE google_id = '' OR google_id IS NULL")

	// Run the main migration
	log.Println("Migrating User model...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("User migration error: %v", err)
	}

	// Recreate the GoogleID unique index with proper NULL handling
	log.Println("Creating GoogleID unique index with NULL handling...")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id
                ON users(google_id) WHERE google_id IS NOT NULL`)

	log.Println("Migrating Expense model...")
	if err := db.AutoMigrate(&models.Expense{}); err != nil {
		log.Fatalf("Expense migration error: %v", err)
	}

	log.Println("Migrating OTP model...")
	if err := db.AutoMigrate(&models.OTP{}); err != nil {
		log.Fatalf("OTP migration error: %v", err)
	}

	log.Println("Migrating BankAccount model...")
	if err := db.AutoMigrate(&models.BankAccount{}); err != nil {
		log.Fatalf("BankAccount migration error: %v", err)
	}

	log.Println("Migrating Transaction model...")
	if err := db.AutoMigrate(&models.Transaction{}); err != nil {
		log.Fatalf("Transaction migration error: %v", err)
	}

	// Create performance indexes
	log.Println("Creating performance indexes...")

	// Enhanced expense indexes for better query performance
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_user_date_type
                ON expenses(user_id, date DESC, type)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_user_category
                ON expenses(user_id, category)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_date_range
                ON expenses(date)`)

	// User indexes
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email
                ON users(email)`)

	// OTP indexes for better performance
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_otp_email_created
                ON otps(email, created_at DESC)`)

	// Additional performance indexes
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_user_created
                ON expenses(user_id, created_at DESC)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_amount
                ON expenses(amount) WHERE amount > 0`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_type_category
                ON expenses(type, category)`)

	// Bank account indexes for better performance
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_bank_accounts_user_id
                ON bank_accounts(user_id)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_bank_accounts_bank_id
                ON bank_accounts(bank_id)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_bank_accounts_status
                ON bank_accounts(status)`)

	// Transaction indexes for better performance
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_user_date
                ON transactions(user_id, transaction_date DESC)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_bank_account
                ON transactions(bank_account_id)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_type_category
                ON transactions(type, category)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_amount
                ON transactions(amount) WHERE amount > 0`)

	log.Println("Performance indexes created successfully")
	log.Println("Database migrations completed successfully")
}
