package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-github/expense-tracker-backend/internal/core/domain"
	"gorm.io/gorm"
)

// Repositories holds all repository interfaces
type Repositories struct {
	User             UserRepository
	BankLink         BankLinkRepository
	Transaction      TransactionRepository
	CategoryOverride CategoryOverrideRepository
}

// NewRepositories creates new repository instances
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:             NewUserRepository(db),
		BankLink:         NewBankLinkRepository(db),
		Transaction:      NewTransactionRepository(db),
		CategoryOverride: NewCategoryOverrideRepository(db),
	}
}

// UserRepository defines user data access methods
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// BankLinkRepository defines bank link data access methods
type BankLinkRepository interface {
	Create(ctx context.Context, bankLink *domain.BankLink) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BankLink, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BankLink, error)
	GetByConsentID(ctx context.Context, consentID string) (*domain.BankLink, error)
	Update(ctx context.Context, bankLink *domain.BankLink) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BankLink, error)
}

// TransactionRepository defines transaction data access methods
type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, from, to *time.Time, limit, offset int) ([]*domain.Transaction, int64, error)
	GetByHashDedupe(ctx context.Context, hashDedupe string) (*domain.Transaction, error)
	Update(ctx context.Context, transaction *domain.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetSummary(ctx context.Context, userID uuid.UUID, from, to *time.Time) (*TransactionSummary, error)
}

// CategoryOverrideRepository defines category override data access methods
type CategoryOverrideRepository interface {
	Create(ctx context.Context, override *domain.CategoryOverride) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CategoryOverride, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.CategoryOverride, error)
	Update(ctx context.Context, override *domain.CategoryOverride) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TransactionSummary represents transaction summary data
type TransactionSummary struct {
	TotalDebit        float64                    `json:"total_debit"`
	TotalCredit       float64                    `json:"total_credit"`
	NetAmount         float64                    `json:"net_amount"`
	CategoryBreakdown map[string]CategorySummary `json:"category_breakdown"`
}

// CategorySummary represents category-level summary
type CategorySummary struct {
	TotalDebit  float64 `json:"total_debit"`
	TotalCredit float64 `json:"total_credit"`
	NetAmount   float64 `json:"net_amount"`
	Count       int64   `json:"count"`
}

// userRepository implements UserRepository
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

// bankLinkRepository implements BankLinkRepository
type bankLinkRepository struct {
	db *gorm.DB
}

func NewBankLinkRepository(db *gorm.DB) BankLinkRepository {
	return &bankLinkRepository{db: db}
}

func (r *bankLinkRepository) Create(ctx context.Context, bankLink *domain.BankLink) error {
	return r.db.WithContext(ctx).Create(bankLink).Error
}

func (r *bankLinkRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BankLink, error) {
	var bankLink domain.BankLink
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&bankLink).Error
	if err != nil {
		return nil, err
	}
	return &bankLink, nil
}

func (r *bankLinkRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BankLink, error) {
	var bankLinks []*domain.BankLink
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&bankLinks).Error
	return bankLinks, err
}

func (r *bankLinkRepository) GetByConsentID(ctx context.Context, consentID string) (*domain.BankLink, error) {
	var bankLink domain.BankLink
	err := r.db.WithContext(ctx).Where("aa_consent_id = ?", consentID).First(&bankLink).Error
	if err != nil {
		return nil, err
	}
	return &bankLink, nil
}

func (r *bankLinkRepository) Update(ctx context.Context, bankLink *domain.BankLink) error {
	return r.db.WithContext(ctx).Save(bankLink).Error
}

func (r *bankLinkRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&domain.BankLink{}).Where("id = ?", id).Update("status", status).Error
}

func (r *bankLinkRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BankLink, error) {
	var bankLinks []*domain.BankLink
	err := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, "ACTIVE").Find(&bankLinks).Error
	return bankLinks, err
}

// transactionRepository implements TransactionRepository
type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, from, to *time.Time, limit, offset int) ([]*domain.Transaction, int64, error) {
	var transactions []*domain.Transaction
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if from != nil {
		query = query.Where("posted_at >= ?", from)
	}
	if to != nil {
		query = query.Where("posted_at <= ?", to)
	}

	// Get total count
	err := query.Model(&domain.Transaction{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = query.Order("posted_at DESC").Limit(limit).Offset(offset).Find(&transactions).Error
	return transactions, total, err
}

func (r *transactionRepository) GetByHashDedupe(ctx context.Context, hashDedupe string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.db.WithContext(ctx).Where("hash_dedupe = ?", hashDedupe).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

func (r *transactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Transaction{}, "id = ?", id).Error
}

func (r *transactionRepository) GetSummary(ctx context.Context, userID uuid.UUID, from, to *time.Time) (*TransactionSummary, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if from != nil {
		query = query.Where("posted_at >= ?", from)
	}
	if to != nil {
		query = query.Where("posted_at <= ?", to)
	}

	var summary TransactionSummary
	var categoryBreakdown []struct {
		Category    string  `json:"category"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
		Count       int64   `json:"count"`
	}

	// Get overall totals
	err := query.Select(`
		SUM(CASE WHEN txn_type = 'DEBIT' THEN amount ELSE 0 END) as total_debit,
		SUM(CASE WHEN txn_type = 'CREDIT' THEN amount ELSE 0 END) as total_credit
	`).Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	summary.NetAmount = summary.TotalCredit - summary.TotalDebit

	// Get category breakdown
	err = query.Select(`
		COALESCE(category, 'Uncategorized') as category,
		SUM(CASE WHEN txn_type = 'DEBIT' THEN amount ELSE 0 END) as total_debit,
		SUM(CASE WHEN txn_type = 'CREDIT' THEN amount ELSE 0 END) as total_credit,
		COUNT(*) as count
	`).Group("category").Scan(&categoryBreakdown).Error
	if err != nil {
		return nil, err
	}

	summary.CategoryBreakdown = make(map[string]CategorySummary)
	for _, cat := range categoryBreakdown {
		summary.CategoryBreakdown[cat.Category] = CategorySummary{
			TotalDebit:  cat.TotalDebit,
			TotalCredit: cat.TotalCredit,
			NetAmount:   cat.TotalCredit - cat.TotalDebit,
			Count:       cat.Count,
		}
	}

	return &summary, nil
}

// categoryOverrideRepository implements CategoryOverrideRepository
type categoryOverrideRepository struct {
	db *gorm.DB
}

func NewCategoryOverrideRepository(db *gorm.DB) CategoryOverrideRepository {
	return &categoryOverrideRepository{db: db}
}

func (r *categoryOverrideRepository) Create(ctx context.Context, override *domain.CategoryOverride) error {
	return r.db.WithContext(ctx).Create(override).Error
}

func (r *categoryOverrideRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CategoryOverride, error) {
	var override domain.CategoryOverride
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&override).Error
	if err != nil {
		return nil, err
	}
	return &override, nil
}

func (r *categoryOverrideRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.CategoryOverride, error) {
	var overrides []*domain.CategoryOverride
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&overrides).Error
	return overrides, err
}

func (r *categoryOverrideRepository) Update(ctx context.Context, override *domain.CategoryOverride) error {
	return r.db.WithContext(ctx).Save(override).Error
}

func (r *categoryOverrideRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CategoryOverride{}, "id = ?", id).Error
}
