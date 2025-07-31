package services

import (
	"crypto/sha256"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/your-github/expense-tracker-backend/internal/core/ports"
)

// Deduplicator handles transaction deduplication
type Deduplicator struct{}

// NewDeduplicator creates a new deduplicator
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

// GenerateHash generates a unique hash for transaction deduplication
func (d *Deduplicator) GenerateHash(txn ports.FITransaction) string {
	// Parse posted_at to get consistent format
	postedAt, err := time.Parse(time.RFC3339, txn.PostedAt)
	if err != nil {
		// Fallback to original string if parsing fails
		postedAt, _ = time.Parse("2006-01-02", txn.PostedAt)
	}

	// Format: YYYY-MM-DDTHH:MM
	formattedTime := postedAt.Format("2006-01-02T15:04")

	// Round amount to 2 decimal places to handle floating point precision issues
	roundedAmount := math.Round(txn.Amount*100) / 100

	// Clean description for consistent hashing
	cleanDesc := d.cleanDescriptionForHash(txn.DescriptionRaw)

	// Create hash components
	components := []string{
		strings.ToLower(txn.AccountRef),
		formattedTime,
		fmt.Sprintf("%.2f", roundedAmount),
		cleanDesc,
	}

	// Join components with separator
	hashString := strings.Join(components, "|")

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(hashString))
	return fmt.Sprintf("%x", hash)
}

// cleanDescriptionForHash cleans description for consistent hashing
func (d *Deduplicator) cleanDescriptionForHash(description string) string {
	if description == "" {
		return ""
	}

	// Convert to lowercase and trim
	cleaned := strings.ToLower(strings.TrimSpace(description))

	// Remove common dynamic parts that shouldn't affect deduplication
	patterns := []string{
		`\s+`,               // Multiple spaces
		`[^\w\s@.-]`,        // Special characters except @, ., -
		`/ref\d+`,           // Reference numbers
		`/txn\d+`,           // Transaction numbers
		`/order\d+`,         // Order numbers
		`/payment\d+`,       // Payment numbers
		`/ride\d+`,          // Ride numbers
		`^\d+/\d+/\d+`,      // Date patterns at start
		`\d{2}:\d{2}:\d{2}`, // Time patterns
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		cleaned = re.ReplaceAllString(cleaned, "")
	}

	return strings.TrimSpace(cleaned)
}

// IsDuplicate checks if a transaction is a duplicate based on hash
func (d *Deduplicator) IsDuplicate(txn ports.FITransaction, existingHashes map[string]bool) bool {
	hash := d.GenerateHash(txn)
	return existingHashes[hash]
}

// AddToHashSet adds a transaction hash to the set
func (d *Deduplicator) AddToHashSet(txn ports.FITransaction, hashSet map[string]bool) {
	hash := d.GenerateHash(txn)
	hashSet[hash] = true
}

// DeduplicateTransactions removes duplicate transactions from a list
func (d *Deduplicator) DeduplicateTransactions(transactions []ports.FITransaction) []ports.FITransaction {
	hashSet := make(map[string]bool)
	uniqueTransactions := make([]ports.FITransaction, 0)

	for _, txn := range transactions {
		if !d.IsDuplicate(txn, hashSet) {
			uniqueTransactions = append(uniqueTransactions, txn)
			d.AddToHashSet(txn, hashSet)
		}
	}

	return uniqueTransactions
}

// GenerateHashForNormalized generates hash for normalized transaction
func (d *Deduplicator) GenerateHashForNormalized(txn NormalizedTransaction) string {
	// Parse posted_at to get consistent format
	postedAt, err := time.Parse(time.RFC3339, txn.PostedAt)
	if err != nil {
		// Fallback to original string if parsing fails
		postedAt, _ = time.Parse("2006-01-02", txn.PostedAt)
	}

	// Format: YYYY-MM-DDTHH:MM
	formattedTime := postedAt.Format("2006-01-02T15:04")

	// Round amount to 2 decimal places
	roundedAmount := math.Round(txn.Amount*100) / 100

	// Clean description for consistent hashing
	cleanDesc := d.cleanDescriptionForHash(txn.DescriptionRaw)

	// Create hash components
	components := []string{
		strings.ToLower(txn.AccountRef),
		formattedTime,
		fmt.Sprintf("%.2f", roundedAmount),
		cleanDesc,
	}

	// Join components with separator
	hashString := strings.Join(components, "|")

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(hashString))
	return fmt.Sprintf("%x", hash)
}
