package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/your-github/expense-tracker-backend/internal/core/ports"
)

// Normalizer normalizes transaction data from AA providers
type Normalizer struct {
	merchantPatterns map[string]string
	upiPatterns      []*regexp.Regexp
}

// NewNormalizer creates a new transaction normalizer
func NewNormalizer() *Normalizer {
	return &Normalizer{
		merchantPatterns: map[string]string{
			"swiggy":     "Food Delivery",
			"zomato":     "Food Delivery",
			"amazon":     "Shopping",
			"flipkart":   "Shopping",
			"netflix":    "Entertainment",
			"spotify":    "Entertainment",
			"uber":       "Transport",
			"ola":        "Transport",
			"paytm":      "Digital Payments",
			"phonepe":    "Digital Payments",
			"google pay": "Digital Payments",
			"atm":        "Cash Withdrawal",
			"neft":       "Bank Transfer",
			"imps":       "Bank Transfer",
			"upi":        "Digital Payments",
			"cheque":     "Bank Transfer",
			"salary":     "Income",
			"interest":   "Income",
			"reimbursement": "Income",
		},
		upiPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)upi/([^@]+)@([^/\s]+)`),
			regexp.MustCompile(`(?i)([^@]+)@([^/\s]+)`),
		},
	}
}

// NormalizeTransaction normalizes a single transaction
func (n *Normalizer) NormalizeTransaction(txn ports.FITransaction) NormalizedTransaction {
	normalized := NormalizedTransaction{
		PostedAt:       txn.PostedAt,
		ValueDate:      txn.ValueDate,
		Amount:         txn.Amount,
		Currency:       txn.Currency,
		TxnType:        txn.Type,
		BalanceAfter:   txn.BalanceAfter,
		SourceMeta:     txn.SourceMeta,
	}

	// Normalize description and extract merchant
	normalized.DescriptionRaw = n.cleanDescription(txn.DescriptionRaw)
	normalized.MerchantName = n.extractMerchant(txn.DescriptionRaw)
	
	// Extract account reference
	normalized.AccountRef = n.extractAccountRef(txn.AccountRef, txn.DescriptionRaw)
	
	// Categorize transaction
	normalized.Category = n.categorizeTransaction(normalized.DescriptionRaw, normalized.MerchantName)
	
	return normalized
}

// NormalizedTransaction represents a normalized transaction
type NormalizedTransaction struct {
	PostedAt       string                 `json:"posted_at"`
	ValueDate      string                 `json:"value_date"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	TxnType        string                 `json:"txn_type"`
	BalanceAfter   *float64               `json:"balance_after"`
	DescriptionRaw string                 `json:"description_raw"`
	MerchantName   string                 `json:"merchant_name"`
	AccountRef     string                 `json:"account_ref"`
	Category       string                 `json:"category"`
	Subcategory    string                 `json:"subcategory"`
	SourceMeta     map[string]interface{} `json:"source_meta"`
}

// cleanDescription cleans and standardizes transaction descriptions
func (n *Normalizer) cleanDescription(description string) string {
	if description == "" {
		return ""
	}

	// Convert to uppercase and trim spaces
	cleaned := strings.ToUpper(strings.TrimSpace(description))

	// Remove common prefixes and suffixes
	patterns := []string{
		`^UPI/`,
		`^NEFT/`,
		`^IMPS/`,
		`/REF\d+$`,
		`/TXN\d+$`,
		`/ORDER\d+$`,
		`/PAYMENT\d+$`,
		`/RIDE\d+$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		cleaned = re.ReplaceAllString(cleaned, "")
	}

	// Remove extra spaces and special characters
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = regexp.MustCompile(`[^\w\s@.-]`).ReplaceAllString(cleaned, "")
	
	return strings.TrimSpace(cleaned)
}

// extractMerchant extracts merchant name from description
func (n *Normalizer) extractMerchant(description string) string {
	if description == "" {
		return ""
	}

	desc := strings.ToLower(description)

	// Check for known merchant patterns
	for pattern, merchant := range n.merchantPatterns {
		if strings.Contains(desc, pattern) {
			return merchant
		}
	}

	// Extract from UPI patterns
	for _, pattern := range n.upiPatterns {
		matches := pattern.FindStringSubmatch(description)
		if len(matches) >= 3 {
			return strings.ToUpper(matches[2]) // Return the handle part
		}
	}

	// Try to extract from common patterns
	if strings.Contains(desc, "atm") {
		return "ATM"
	}
	if strings.Contains(desc, "salary") {
		return "Salary"
	}
	if strings.Contains(desc, "interest") {
		return "Interest"
	}

	return "Unknown"
}

// extractAccountRef extracts account reference from description or account ref
func (n *Normalizer) extractAccountRef(accountRef, description string) string {
	if accountRef != "" {
		return accountRef
	}

	// Try to extract from UPI patterns
	for _, pattern := range n.upiPatterns {
		matches := pattern.FindStringSubmatch(description)
		if len(matches) >= 3 {
			return fmt.Sprintf("%s@%s", matches[1], matches[2])
		}
	}

	// Try to extract masked account number
	accountPattern := regexp.MustCompile(`(\d{4}\*{4}\d{4})`)
	matches := accountPattern.FindStringSubmatch(description)
	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}

// categorizeTransaction categorizes transaction based on description and merchant
func (n *Normalizer) categorizeTransaction(description, merchant string) string {
	desc := strings.ToLower(description)
	merchantLower := strings.ToLower(merchant)

	// Check merchant-based categorization first
	if merchantLower != "unknown" {
		return merchant
	}

	// Check description-based categorization
	if strings.Contains(desc, "food") || strings.Contains(desc, "restaurant") {
		return "Food & Dining"
	}
	if strings.Contains(desc, "fuel") || strings.Contains(desc, "petrol") || strings.Contains(desc, "diesel") {
		return "Transport"
	}
	if strings.Contains(desc, "medical") || strings.Contains(desc, "hospital") || strings.Contains(desc, "pharmacy") {
		return "Healthcare"
	}
	if strings.Contains(desc, "education") || strings.Contains(desc, "school") || strings.Contains(desc, "college") {
		return "Education"
	}
	if strings.Contains(desc, "rent") || strings.Contains(desc, "electricity") || strings.Contains(desc, "water") {
		return "Utilities"
	}
	if strings.Contains(desc, "salary") || strings.Contains(desc, "income") {
		return "Income"
	}
	if strings.Contains(desc, "interest") {
		return "Income"
	}
	if strings.Contains(desc, "atm") || strings.Contains(desc, "withdrawal") {
		return "Cash Withdrawal"
	}
	if strings.Contains(desc, "transfer") || strings.Contains(desc, "neft") || strings.Contains(desc, "imps") {
		return "Bank Transfer"
	}

	return "Uncategorized"
}

// ApplyUserOverrides applies user-defined category overrides
func (n *Normalizer) ApplyUserOverrides(transaction *NormalizedTransaction, overrides []CategoryOverride) {
	for _, override := range overrides {
		if n.matchesOverride(transaction.DescriptionRaw, override.Matcher) {
			transaction.Category = override.Category
			transaction.Subcategory = override.Subcategory
			break
		}
	}
}

// CategoryOverride represents a user-defined category rule
type CategoryOverride struct {
	Matcher    string `json:"matcher"`
	Category   string `json:"category"`
	Subcategory string `json:"subcategory"`
}

// matchesOverride checks if transaction description matches the override pattern
func (n *Normalizer) matchesOverride(description, matcher string) bool {
	// Check if it's a regex pattern
	if strings.HasPrefix(matcher, "/") && strings.HasSuffix(matcher, "/") {
		pattern := matcher[1 : len(matcher)-1]
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(description)
	}

	// Simple contains match
	return strings.Contains(strings.ToLower(description), strings.ToLower(matcher))
} 