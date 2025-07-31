package ports

import "time"

// ConsentRequest represents a request to create an AA consent
type ConsentRequest struct {
	UserID      string    `json:"user_id"`
	FIType      string    `json:"fi_type"`      // account types: "SAVINGS", "CURRENT", etc.
	Purpose     string    `json:"purpose"`      // e.g. "EXPENSE_ANALYSIS"
	DateRange   DateRange `json:"date_range"`   // ISO dates
	Frequency   string    `json:"frequency"`    // e.g. "DAILY"
	RedirectURL string    `json:"redirect_url"`
	WebhookURL  string    `json:"webhook_url"`
}

// DateRange represents a date range for consent
type DateRange struct {
	From string `json:"from"` // ISO date
	To   string `json:"to"`   // ISO date
}

// ConsentHandle represents the response from creating a consent
type ConsentHandle struct {
	ConsentID   string `json:"consent_id"`
	RedirectURL string `json:"redirect_url"`
	Status      string `json:"status"`
}

// DataSession represents a data fetch session
type DataSession struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"` // "PENDING", "READY", "FAILED"
}

// FITransaction represents a financial institution transaction
type FITransaction struct {
	PostedAt       string                 `json:"posted_at"`
	ValueDate      string                 `json:"value_date"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Type           string                 `json:"type"` // "DEBIT" | "CREDIT"
	DescriptionRaw string                 `json:"description_raw"`
	Merchant       string                 `json:"merchant"`
	AccountRef     string                 `json:"account_ref"`
	BalanceAfter   *float64               `json:"balance_after"`
	SourceMeta     map[string]interface{} `json:"source_meta"`
}

// ConsentStatus represents the status of a consent
type ConsentStatus string

const (
	ConsentStatusPending ConsentStatus = "PENDING"
	ConsentStatusActive  ConsentStatus = "ACTIVE"
	ConsentStatusRevoked ConsentStatus = "REVOKED"
	ConsentStatusExpired ConsentStatus = "EXPIRED"
)

// SessionStatus represents the status of a data session
type SessionStatus string

const (
	SessionStatusPending SessionStatus = "PENDING"
	SessionStatusReady   SessionStatus = "READY"
	SessionStatusFailed  SessionStatus = "FAILED"
)

// AAClient defines the interface for Account Aggregator providers
type AAClient interface {
	// CreateConsent creates a new consent request
	CreateConsent(req ConsentRequest) (ConsentHandle, error)
	
	// GetConsentStatus retrieves the current status of a consent
	GetConsentStatus(consentID string) (ConsentStatus, error)
	
	// CreateDataSession creates a new data fetch session
	CreateDataSession(consentID string, fromISO, toISO string) (DataSession, error)
	
	// GetSessionStatus retrieves the current status of a data session
	GetSessionStatus(sessionID string) (SessionStatus, error)
	
	// FetchTransactions fetches transactions for a given session
	FetchTransactions(sessionID string) ([]FITransaction, error)
	
	// RevokeConsent revokes an active consent
	RevokeConsent(consentID string) error
}

// WebhookEvent represents an incoming webhook event from AA
type WebhookEvent struct {
	EventType   string                 `json:"event_type"`   // "CONSENT_STATUS_UPDATE", "DATA_READY", etc.
	ConsentID   string                 `json:"consent_id"`
	SessionID   string                 `json:"session_id,omitempty"`
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
	Signature   string                 `json:"signature"`
}

// WebhookHandler defines the interface for handling webhook events
type WebhookHandler interface {
	// HandleWebhook processes incoming webhook events
	HandleWebhook(event WebhookEvent) error
	
	// VerifySignature verifies the webhook signature
	VerifySignature(payload []byte, signature string) bool
} 