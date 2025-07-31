package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-github/expense-tracker-backend/internal/core/domain"
	"github.com/your-github/expense-tracker-backend/internal/core/ports"
	"github.com/your-github/expense-tracker-backend/internal/repo"
	"go.uber.org/zap"
)

// AAService orchestrates Account Aggregator operations
type AAService struct {
	aaClient     ports.AAClient
	repositories *repo.Repositories
	normalizer   *Normalizer
	deduplicator *Deduplicator
	logger       *zap.Logger
}

// NewAAService creates a new AA service
func NewAAService(
	aaClient ports.AAClient,
	repositories *repo.Repositories,
	normalizer *Normalizer,
	deduplicator *Deduplicator,
	logger *zap.Logger,
) *AAService {
	return &AAService{
		aaClient:     aaClient,
		repositories: repositories,
		normalizer:   normalizer,
		deduplicator: deduplicator,
		logger:       logger,
	}
}

// InitiateConsent initiates a new consent request
func (s *AAService) InitiateConsent(ctx context.Context, userID uuid.UUID, req ports.ConsentRequest) (*domain.BankLink, error) {
	// Create consent via AA client
	consentHandle, err := s.aaClient.CreateConsent(req)
	if err != nil {
		s.logger.Error("Failed to create consent", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to create consent: %w", err)
	}

	// Create bank link record
	bankLink := &domain.BankLink{
		ID:          uuid.New(),
		UserID:      userID,
		AAConsentID: consentHandle.ConsentID,
		FIType:      req.FIType,
		Status:      "PENDING",
		ValidTill:   nil, // Will be set when consent is approved
	}

	err = s.repositories.BankLink.Create(ctx, bankLink)
	if err != nil {
		s.logger.Error("Failed to create bank link", zap.Error(err), zap.String("consent_id", consentHandle.ConsentID))
		return nil, fmt.Errorf("failed to create bank link: %w", err)
	}

	s.logger.Info("Consent initiated successfully",
		zap.String("consent_id", consentHandle.ConsentID),
		zap.String("user_id", userID.String()),
		zap.String("redirect_url", consentHandle.RedirectURL))

	return bankLink, nil
}

// HandleConsentCallback handles consent status updates from AA
func (s *AAService) HandleConsentCallback(ctx context.Context, consentID, status string) error {
	// Find bank link by consent ID
	bankLink, err := s.repositories.BankLink.GetByConsentID(ctx, consentID)
	if err != nil {
		s.logger.Error("Failed to find bank link for consent", zap.Error(err), zap.String("consent_id", consentID))
		return fmt.Errorf("failed to find bank link: %w", err)
	}

	// Update status
	err = s.repositories.BankLink.UpdateStatus(ctx, bankLink.ID, status)
	if err != nil {
		s.logger.Error("Failed to update bank link status", zap.Error(err), zap.String("consent_id", consentID))
		return fmt.Errorf("failed to update bank link status: %w", err)
	}

	// Set valid till date if consent is active
	if status == "ACTIVE" {
		validTill := time.Now().AddDate(0, 1, 0) // 1 month validity
		bankLink.ValidTill = &validTill
		bankLink.Status = status
		err = s.repositories.BankLink.Update(ctx, bankLink)
		if err != nil {
			s.logger.Error("Failed to update bank link validity", zap.Error(err), zap.String("consent_id", consentID))
		}
	}

	s.logger.Info("Consent status updated",
		zap.String("consent_id", consentID),
		zap.String("status", status),
		zap.String("user_id", bankLink.UserID.String()))

	return nil
}

// FetchTransactions fetches transactions for a bank link
func (s *AAService) FetchTransactions(ctx context.Context, userID uuid.UUID, bankLinkID uuid.UUID, fromDate, toDate string) (*DataFetchResult, error) {
	// Get bank link
	bankLink, err := s.repositories.BankLink.GetByID(ctx, bankLinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank link: %w", err)
	}

	// Verify ownership
	if bankLink.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to bank link")
	}

	// Verify consent is active
	if bankLink.Status != "ACTIVE" {
		return nil, fmt.Errorf("consent is not active: %s", bankLink.Status)
	}

	// Create data session
	dataSession, err := s.aaClient.CreateDataSession(bankLink.AAConsentID, fromDate, toDate)
	if err != nil {
		s.logger.Error("Failed to create data session", zap.Error(err), zap.String("consent_id", bankLink.AAConsentID))
		return nil, fmt.Errorf("failed to create data session: %w", err)
	}

	result := &DataFetchResult{
		SessionID: dataSession.SessionID,
		Status:    dataSession.Status,
	}

	// If session is ready, fetch transactions immediately
	if dataSession.Status == "READY" {
		transactions, err := s.fetchAndProcessTransactions(ctx, dataSession.SessionID, userID, bankLinkID)
		if err != nil {
			return nil, err
		}
		result.Transactions = transactions
		result.Processed = true
	}

	return result, nil
}

// HandleDataReadyWebhook handles data ready webhook from AA
func (s *AAService) HandleDataReadyWebhook(ctx context.Context, sessionID string) error {
	s.logger.Info("Processing data ready webhook", zap.String("session_id", sessionID))

	// Get session status to verify it's ready
	status, err := s.aaClient.GetSessionStatus(sessionID)
	if err != nil {
		s.logger.Error("Failed to get session status", zap.Error(err), zap.String("session_id", sessionID))
		return fmt.Errorf("failed to get session status: %w", err)
	}

	if status != ports.SessionStatusReady {
		s.logger.Warn("Session is not ready", zap.String("session_id", sessionID), zap.String("status", string(status)))
		return fmt.Errorf("session is not ready: %s", status)
	}

	// Fetch and process transactions
	// Note: We need to determine user_id and bank_link_id from session
	// This is a simplified implementation - in production, you'd store session metadata
	_, err = s.fetchAndProcessTransactions(ctx, sessionID, uuid.Nil, uuid.Nil)
	return err
}

// fetchAndProcessTransactions fetches transactions and processes them
func (s *AAService) fetchAndProcessTransactions(ctx context.Context, sessionID string, userID uuid.UUID, bankLinkID uuid.UUID) ([]*domain.Transaction, error) {
	// Fetch transactions from AA
	fiTransactions, err := s.aaClient.FetchTransactions(sessionID)
	if err != nil {
		s.logger.Error("Failed to fetch transactions", zap.Error(err), zap.String("session_id", sessionID))
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	s.logger.Info("Fetched transactions from AA",
		zap.String("session_id", sessionID),
		zap.Int("count", len(fiTransactions)))

	// Deduplicate transactions
	uniqueTransactions := s.deduplicator.DeduplicateTransactions(fiTransactions)
	s.logger.Info("Deduplicated transactions",
		zap.Int("original", len(fiTransactions)),
		zap.Int("unique", len(uniqueTransactions)))

	// Get existing transaction hashes for this user
	existingHashes := make(map[string]bool)
	existingTransactions, _, err := s.repositories.Transaction.GetByUserID(ctx, userID, nil, nil, 1000, 0)
	if err != nil {
		s.logger.Error("Failed to get existing transactions", zap.Error(err))
		return nil, fmt.Errorf("failed to get existing transactions: %w", err)
	}

	for _, txn := range existingTransactions {
		existingHashes[txn.HashDedupe] = true
	}

	// Process and store new transactions
	var newTransactions []*domain.Transaction
	for _, fiTxn := range uniqueTransactions {
		// Generate hash for deduplication
		hash := s.deduplicator.GenerateHash(fiTxn)

		// Skip if already exists
		if existingHashes[hash] {
			continue
		}

		// Normalize transaction
		normalized := s.normalizer.NormalizeTransaction(fiTxn)

		// Parse posted_at
		postedAt, err := time.Parse(time.RFC3339, fiTxn.PostedAt)
		if err != nil {
			s.logger.Warn("Failed to parse posted_at", zap.Error(err), zap.String("posted_at", fiTxn.PostedAt))
			postedAt = time.Now()
		}

		// Parse value_date if available
		var valueDate *time.Time
		if fiTxn.ValueDate != "" {
			if parsed, err := time.Parse("2006-01-02", fiTxn.ValueDate); err == nil {
				valueDate = &parsed
			}
		}

		// Create domain transaction
		transaction := &domain.Transaction{
			ID:             uuid.New(),
			UserID:         userID,
			BankLinkID:     &bankLinkID,
			PostedAt:       postedAt,
			ValueDate:      valueDate,
			Amount:         fiTxn.Amount,
			Currency:       fiTxn.Currency,
			TxnType:        fiTxn.Type,
			BalanceAfter:   fiTxn.BalanceAfter,
			DescriptionRaw: normalized.DescriptionRaw,
			MerchantName:   normalized.MerchantName,
			AccountRef:     normalized.AccountRef,
			Category:       normalized.Category,
			Subcategory:    normalized.Subcategory,
			HashDedupe:     hash,
			SourceMeta:     domain.JSONB(fiTxn.SourceMeta),
		}

		// Store transaction
		err = s.repositories.Transaction.Create(ctx, transaction)
		if err != nil {
			s.logger.Error("Failed to create transaction", zap.Error(err), zap.String("hash", hash))
			continue // Continue with other transactions
		}

		newTransactions = append(newTransactions, transaction)
	}

	s.logger.Info("Processed transactions",
		zap.String("session_id", sessionID),
		zap.Int("new_transactions", len(newTransactions)))

	return newTransactions, nil
}

// DataFetchResult represents the result of a data fetch operation
type DataFetchResult struct {
	SessionID    string                `json:"session_id"`
	Status       string                `json:"status"`
	Processed    bool                  `json:"processed"`
	Transactions []*domain.Transaction `json:"transactions,omitempty"`
}

// GetActiveBankLinks returns active bank links for a user
func (s *AAService) GetActiveBankLinks(ctx context.Context, userID uuid.UUID) ([]*domain.BankLink, error) {
	return s.repositories.BankLink.GetActiveByUserID(ctx, userID)
}

// RevokeConsent revokes an active consent
func (s *AAService) RevokeConsent(ctx context.Context, userID uuid.UUID, bankLinkID uuid.UUID) error {
	// Get bank link
	bankLink, err := s.repositories.BankLink.GetByID(ctx, bankLinkID)
	if err != nil {
		return fmt.Errorf("failed to get bank link: %w", err)
	}

	// Verify ownership
	if bankLink.UserID != userID {
		return fmt.Errorf("unauthorized access to bank link")
	}

	// Revoke consent via AA client
	err = s.aaClient.RevokeConsent(bankLink.AAConsentID)
	if err != nil {
		s.logger.Error("Failed to revoke consent", zap.Error(err), zap.String("consent_id", bankLink.AAConsentID))
		return fmt.Errorf("failed to revoke consent: %w", err)
	}

	// Update status locally
	err = s.repositories.BankLink.UpdateStatus(ctx, bankLinkID, "REVOKED")
	if err != nil {
		s.logger.Error("Failed to update bank link status", zap.Error(err), zap.String("consent_id", bankLink.AAConsentID))
		return fmt.Errorf("failed to update bank link status: %w", err)
	}

	s.logger.Info("Consent revoked successfully",
		zap.String("consent_id", bankLink.AAConsentID),
		zap.String("user_id", userID.String()))

	return nil
}
