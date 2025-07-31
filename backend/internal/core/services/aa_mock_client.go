package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/your-github/expense-tracker-backend/internal/core/ports"
)

// MockAAClient implements the AAClient interface for testing
type MockAAClient struct {
	consents map[string]*MockConsent
	sessions map[string]*MockSession
	mu       sync.RWMutex
}

// MockConsent represents a mock consent in memory
type MockConsent struct {
	ConsentID   string
	UserID      string
	Status      ports.ConsentStatus
	FIType      string
	CreatedAt   time.Time
	ValidTill   time.Time
	RedirectURL string
}

// MockSession represents a mock data session in memory
type MockSession struct {
	SessionID string
	ConsentID string
	Status    ports.SessionStatus
	FromDate  string
	ToDate    string
	CreatedAt time.Time
}

// NewMockAAClient creates a new mock AA client
func NewMockAAClient() *MockAAClient {
	return &MockAAClient{
		consents: make(map[string]*MockConsent),
		sessions: make(map[string]*MockSession),
	}
}

// CreateConsent creates a mock consent
func (m *MockAAClient) CreateConsent(req ports.ConsentRequest) (ports.ConsentHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	consentID := fmt.Sprintf("consent_%s", uuid.New().String()[:8])
	redirectURL := fmt.Sprintf("https://mock-aa.com/consent/%s", consentID)

	consent := &MockConsent{
		ConsentID:   consentID,
		UserID:      req.UserID,
		Status:      ports.ConsentStatusPending,
		FIType:      req.FIType,
		CreatedAt:   time.Now(),
		ValidTill:   time.Now().AddDate(0, 1, 0), // 1 month validity
		RedirectURL: redirectURL,
	}

	m.consents[consentID] = consent

	return ports.ConsentHandle{
		ConsentID:   consentID,
		RedirectURL: redirectURL,
		Status:      string(consent.Status),
	}, nil
}

// GetConsentStatus retrieves the status of a mock consent
func (m *MockAAClient) GetConsentStatus(consentID string) (ports.ConsentStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	consent, exists := m.consents[consentID]
	if !exists {
		return "", fmt.Errorf("consent not found: %s", consentID)
	}

	return consent.Status, nil
}

// CreateDataSession creates a mock data session
func (m *MockAAClient) CreateDataSession(consentID string, fromISO, toISO string) (ports.DataSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if consent exists and is active
	consent, exists := m.consents[consentID]
	if !exists {
		return ports.DataSession{}, fmt.Errorf("consent not found: %s", consentID)
	}

	if consent.Status != ports.ConsentStatusActive {
		return ports.DataSession{}, fmt.Errorf("consent is not active: %s", consentID)
	}

	sessionID := fmt.Sprintf("session_%s", uuid.New().String()[:8])

	session := &MockSession{
		SessionID: sessionID,
		ConsentID: consentID,
		Status:    ports.SessionStatusPending,
		FromDate:  fromISO,
		ToDate:    toISO,
		CreatedAt: time.Now(),
	}

	m.sessions[sessionID] = session

	// Simulate async processing - change status to READY after a short delay
	go func() {
		time.Sleep(2 * time.Second)
		m.mu.Lock()
		if s, exists := m.sessions[sessionID]; exists {
			s.Status = ports.SessionStatusReady
		}
		m.mu.Unlock()
	}()

	return ports.DataSession{
		SessionID: sessionID,
		Status:    string(session.Status),
	}, nil
}

// GetSessionStatus retrieves the status of a mock session
func (m *MockAAClient) GetSessionStatus(sessionID string) (ports.SessionStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	return session.Status, nil
}

// FetchTransactions generates mock transactions for a session
func (m *MockAAClient) FetchTransactions(sessionID string) ([]ports.FITransaction, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != ports.SessionStatusReady {
		return nil, fmt.Errorf("session is not ready: %s", sessionID)
	}

	// Generate mock transactions
	transactions := generateMockTransactions(session.FromDate, session.ToDate)
	return transactions, nil
}

// RevokeConsent revokes a mock consent
func (m *MockAAClient) RevokeConsent(consentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	consent, exists := m.consents[consentID]
	if !exists {
		return fmt.Errorf("consent not found: %s", consentID)
	}

	consent.Status = ports.ConsentStatusRevoked
	return nil
}

// SimulateConsentApproval simulates user approving a consent (for testing)
func (m *MockAAClient) SimulateConsentApproval(consentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	consent, exists := m.consents[consentID]
	if !exists {
		return fmt.Errorf("consent not found: %s", consentID)
	}

	consent.Status = ports.ConsentStatusActive
	return nil
}

// generateMockTransactions creates realistic mock transactions
func generateMockTransactions(fromDate, toDate string) []ports.FITransaction {
	transactions := []ports.FITransaction{}

	// Parse date range
	from, _ := time.Parse("2006-01-02", fromDate)
	to, _ := time.Parse("2006-01-02", toDate)

	// Generate transactions for each day in the range
	current := from
	for current.Before(to) || current.Equal(to) {
		// Generate 1-5 transactions per day
		numTxns := rand.Intn(5) + 1

		for i := 0; i < numTxns; i++ {
			// Random time during the day
			hour := rand.Intn(24)
			minute := rand.Intn(60)
			postedAt := time.Date(current.Year(), current.Month(), current.Day(), hour, minute, 0, 0, time.UTC)

			// Random amount between 100 and 10000
			amount := float64(rand.Intn(9900) + 100)

			// Random transaction type
			txnType := "DEBIT"
			if rand.Float32() < 0.3 { // 30% chance of credit
				txnType = "CREDIT"
			}

			// Generate realistic descriptions
			description, merchant := generateMockDescription(txnType)

			// Generate account reference
			accountRef := generateMockAccountRef()

			transaction := ports.FITransaction{
				PostedAt:       postedAt.Format(time.RFC3339),
				ValueDate:      postedAt.Format("2006-01-02"),
				Amount:         amount,
				Currency:       "INR",
				Type:           txnType,
				DescriptionRaw: description,
				Merchant:       merchant,
				AccountRef:     accountRef,
				BalanceAfter:   nil, // Mock doesn't track balance
				SourceMeta: map[string]interface{}{
					"source":       "mock_aa",
					"generated_at": time.Now().Format(time.RFC3339),
				},
			}

			transactions = append(transactions, transaction)
		}

		current = current.AddDate(0, 0, 1)
	}

	return transactions
}

// generateMockDescription creates realistic transaction descriptions
func generateMockDescription(txnType string) (string, string) {
	var merchants []string

	// Select merchants based on transaction type
	if txnType == "CREDIT" {
		merchants = []string{
			"INTEREST CREDIT", "SALARY CREDIT", "REIMBURSEMENT", "NEFT", "IMPS",
			"REFUND", "CASH DEPOSIT", "CHEQUE", "UPI",
		}
	} else {
		merchants = []string{
			"SWIGGY", "ZOMATO", "AMAZON", "FLIPKART", "NETFLIX", "SPOTIFY",
			"UBER", "OLA", "PAYTM", "PHONEPE", "GOOGLE PAY", "BHARAT QR",
			"ATM WITHDRAWAL", "NEFT", "IMPS", "UPI", "CHEQUE",
		}
	}

	merchant := merchants[rand.Intn(len(merchants))]

	var description string
	switch merchant {
	case "SWIGGY", "ZOMATO":
		description = fmt.Sprintf("%s/ORDER/%s", merchant, strings.ToUpper(generateRandomString(8)))
	case "AMAZON", "FLIPKART":
		description = fmt.Sprintf("%s/PAYMENT/%s", merchant, strings.ToUpper(generateRandomString(8)))
	case "UBER", "OLA":
		description = fmt.Sprintf("%s/RIDE/%s", merchant, strings.ToUpper(generateRandomString(8)))
	case "PAYTM", "PHONEPE", "GOOGLE PAY":
		description = fmt.Sprintf("%s/UPI/%s@%s", merchant, generateRandomString(6), generateRandomString(3))
	case "ATM WITHDRAWAL":
		description = "ATM WITHDRAWAL/XXXX1234/BRANCH"
	case "NEFT", "IMPS":
		description = fmt.Sprintf("%s/TRANSFER/%s", merchant, generateRandomString(8))
	case "UPI":
		description = fmt.Sprintf("UPI/%s@%s", generateRandomString(6), generateRandomString(3))
	case "INTEREST CREDIT":
		description = "INTEREST CREDIT/SAVINGS ACCOUNT"
	case "SALARY CREDIT":
		description = "SALARY CREDIT/COMPANY NAME"
	case "REFUND":
		description = fmt.Sprintf("REFUND/%s", strings.ToUpper(generateRandomString(8)))
	default:
		description = fmt.Sprintf("%s/%s", merchant, strings.ToUpper(generateRandomString(8)))
	}

	return description, merchant
}

// generateMockAccountRef creates realistic account references
func generateMockAccountRef() string {
	patterns := []string{
		"XXXX1234",     // Masked account number
		"user@upi",     // UPI VPA
		"user@okicici", // UPI VPA
		"user@paytm",   // UPI VPA
		"user@ybl",     // UPI VPA
	}

	return patterns[rand.Intn(len(patterns))]
}

// generateRandomString creates a random string of given length
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// VerifySignature verifies webhook signature (mock implementation)
func (m *MockAAClient) VerifySignature(payload []byte, signature string) bool {
	// Mock signature verification - always returns true for testing
	return true
}

// GenerateSignature generates a mock signature for webhooks
func (m *MockAAClient) GenerateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
