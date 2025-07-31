package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// BankLink represents a user's bank account link via AA
type BankLink struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	AAConsentID  string    `gorm:"not null;index" json:"aa_consent_id"`
	FIType       string    `gorm:"not null" json:"fi_type"` // "SAVINGS", "CURRENT", etc.
	Status       string    `gorm:"not null;index" json:"status"` // "PENDING", "ACTIVE", "REVOKED"
	ValidTill    *time.Time `json:"valid_till"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User         User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Transactions []Transaction  `gorm:"foreignKey:BankLinkID" json:"transactions,omitempty"`
}

// Transaction represents a bank transaction
type Transaction struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	BankLinkID    *uuid.UUID `gorm:"type:uuid;index" json:"bank_link_id"`
	PostedAt      time.Time `gorm:"not null;index" json:"posted_at"`
	ValueDate     *time.Time `json:"value_date"`
	Amount        float64   `gorm:"type:numeric(14,2);not null" json:"amount"`
	Currency      string    `gorm:"default:'INR'" json:"currency"`
	TxnType       string    `gorm:"not null;index" json:"txn_type"` // "DEBIT" | "CREDIT"
	BalanceAfter  *float64  `gorm:"type:numeric(14,2)" json:"balance_after"`
	DescriptionRaw string   `json:"description_raw"`
	MerchantName  string    `json:"merchant_name"`
	AccountRef    string    `json:"account_ref"` // masked account / VPA
	Category      string    `json:"category"`
	Subcategory   string    `json:"subcategory"`
	HashDedupe    string    `gorm:"uniqueIndex;not null" json:"hash_dedupe"`
	SourceMeta    JSONB     `gorm:"type:jsonb;default:'{}'::jsonb" json:"source_meta"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BankLink  *BankLink `gorm:"foreignKey:BankLinkID" json:"bank_link,omitempty"`
}

// CategoryOverride represents user-defined category rules
type CategoryOverride struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Matcher    string    `gorm:"not null" json:"matcher"` // regex or contains
	Category   string    `gorm:"not null" json:"category"`
	Subcategory string  `json:"subcategory"`
	CreatedAt  time.Time `gorm:"default:now()" json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// JSONB is a custom type for PostgreSQL JSONB
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (interface{}, error) {
	return j, nil
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
}

// MarshalJSON implements json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		*j = make(JSONB)
	}
	return json.Unmarshal(data, (*map[string]interface{})(j))
} 