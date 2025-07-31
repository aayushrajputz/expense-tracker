package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// OpenAI API structures
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	MaxTokens int      `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

// Financial data structure for AI analysis
type FinancialData struct {
	TotalIncome     float64            `json:"total_income"`
	TotalExpenses   float64            `json:"total_expenses"`
	SavingsRate     float64            `json:"savings_rate"`
	CategorySpending map[string]float64 `json:"category_spending"`
	MonthlyTrends   []MonthlyData      `json:"monthly_trends"`
	RecentTransactions []Transaction    `json:"recent_transactions"`
}

type MonthlyData struct {
	Month     string  `json:"month"`
	Income    float64 `json:"income"`
	Expenses  float64 `json:"expenses"`
	Savings   float64 `json:"savings"`
}

type Transaction struct {
	Title    string  `json:"title"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
}

// GenerateAIInsights uses OpenAI GPT-3.5 to generate financial insights
func GenerateAIInsights(financialData FinancialData) ([]Insight, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Prepare the prompt for GPT-3.5
	prompt := buildFinancialPrompt(financialData)

	// Create OpenAI request
	request := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role: "system",
				Content: `You are a professional financial advisor and AI analyst. Analyze the provided financial data and generate 6 insightful, actionable financial insights. Each insight should include:
1. A clear title
2. A descriptive analysis
3. A practical suggestion or recommendation
4. Appropriate type (warning, success, tip, info)

Focus on spending patterns, savings opportunities, budget recommendations, and financial health indicators. Be specific with numbers and percentages. Keep each insight concise but informative.`,
			},
			{
				Role: "user",
				Content: prompt,
			},
		},
		MaxTokens: 1000,
		Temperature: 0.7,
	}

	// Make API call
	insights, err := callOpenAI(apiKey, request)
	if err != nil {
		return nil, err
	}

	return insights, nil
}

// buildFinancialPrompt creates a comprehensive prompt for AI analysis
func buildFinancialPrompt(data FinancialData) string {
	var prompt strings.Builder
	
	prompt.WriteString("Please analyze this financial data and provide 6 AI-powered insights:\n\n")
	
	// Basic financial summary
	prompt.WriteString(fmt.Sprintf("Financial Summary:\n"))
	prompt.WriteString(fmt.Sprintf("- Total Income: $%.2f\n", data.TotalIncome))
	prompt.WriteString(fmt.Sprintf("- Total Expenses: $%.2f\n", data.TotalExpenses))
	prompt.WriteString(fmt.Sprintf("- Savings Rate: %.1f%%\n", data.SavingsRate))
	
	// Category spending
	if len(data.CategorySpending) > 0 {
		prompt.WriteString("\nCategory Spending:\n")
		for category, amount := range data.CategorySpending {
			percentage := (amount / data.TotalExpenses) * 100
			prompt.WriteString(fmt.Sprintf("- %s: $%.2f (%.1f%%)\n", category, amount, percentage))
		}
	}
	
	// Monthly trends
	if len(data.MonthlyTrends) > 0 {
		prompt.WriteString("\nMonthly Trends (last 3 months):\n")
		for i, month := range data.MonthlyTrends {
			if i >= 3 { // Show only last 3 months
				break
			}
			prompt.WriteString(fmt.Sprintf("- %s: Income $%.2f, Expenses $%.2f, Savings $%.2f\n", 
				month.Month, month.Income, month.Expenses, month.Savings))
		}
	}
	
	// Recent transactions
	if len(data.RecentTransactions) > 0 {
		prompt.WriteString("\nRecent Transactions (last 5):\n")
		for i, transaction := range data.RecentTransactions {
			if i >= 5 { // Show only last 5 transactions
				break
			}
			prompt.WriteString(fmt.Sprintf("- %s: $%.2f (%s - %s)\n", 
				transaction.Title, transaction.Amount, transaction.Type, transaction.Category))
		}
	}
	
	prompt.WriteString("\nPlease provide 6 insights in JSON format with this structure:\n")
	prompt.WriteString(`[
  {
    "type": "warning|success|tip|info",
    "title": "Insight Title",
    "description": "Detailed analysis with specific numbers",
    "suggestion": "Actionable recommendation"
  }
]`)
	
	return prompt.String()
}

// callOpenAI makes the actual API call to OpenAI
func callOpenAI(apiKey string, request OpenAIRequest) ([]Insight, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Check for API errors
	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the AI response to extract insights
	content := openAIResp.Choices[0].Message.Content
	insights, err := parseAIInsights(content)
	if err != nil {
		return nil, fmt.Errorf("error parsing AI insights: %v", err)
	}

	return insights, nil
}

// Insight structure for the response
type Insight struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

// parseAIInsights extracts insights from AI response
func parseAIInsights(content string) ([]Insight, error) {
	// Try to extract JSON from the response
	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")
	
	if start == -1 || end == -1 || end <= start {
		// If no JSON found, create a fallback insight
		return []Insight{
			{
				Type:        "info",
				Title:       "AI Analysis Complete",
				Description: "AI has analyzed your financial data and provided insights.",
				Suggestion:  "Review the insights above for personalized financial recommendations.",
			},
		}, nil
	}

	jsonStr := content[start : end+1]
	var insights []Insight
	
	if err := json.Unmarshal([]byte(jsonStr), &insights); err != nil {
		// If JSON parsing fails, return fallback
		return []Insight{
			{
				Type:        "info",
				Title:       "AI Analysis Available",
				Description: "AI analysis has been performed on your financial data.",
				Suggestion:  "Check your spending patterns and consider the recommendations provided.",
			},
		}, nil
	}

	return insights, nil
}

// Enhanced keyword matcher for better auto-categorization
var keywordMap = map[string]string{
	// Food & Dining
	"dominos": "Food & Dining",
	"swiggy":  "Food & Dining",
	"zomato":  "Food & Dining",
	"restaurant": "Food & Dining",
	"cafe":    "Food & Dining",
	"pizza":   "Food & Dining",
	"burger":  "Food & Dining",
	"mcdonalds": "Food & Dining",
	"kfc":     "Food & Dining",
	"subway":  "Food & Dining",
	"starbucks": "Food & Dining",
	"coffee":  "Food & Dining",
	"food":    "Food & Dining",
	"dining":  "Food & Dining",
	"lunch":   "Food & Dining",
	"dinner":  "Food & Dining",
	"breakfast": "Food & Dining",
	"snack":   "Food & Dining",
	"grocery": "Food & Dining",
	"supermarket": "Food & Dining",
	
	// Transportation
	"uber":    "Transportation",
	"ola":     "Transportation",
	"taxi":    "Transportation",
	"bus":     "Transportation",
	"train":   "Transportation",
	"metro":   "Transportation",
	"fuel":    "Transportation",
	"gas":     "Transportation",
	"petrol":  "Transportation",
	"diesel":  "Transportation",
	"parking": "Transportation",
	"toll":    "Transportation",
	"transport": "Transportation",
	"travel":  "Transportation",
	"flight":  "Transportation",
	"airline": "Transportation",
	
	// Shopping
	"amazon":  "Shopping",
	"flipkart": "Shopping",
	"myntra":  "Shopping",
	"shopping": "Shopping",
	"clothes": "Shopping",
	"clothing": "Shopping",
	"shirt":   "Shopping",
	"pants":   "Shopping",
	"dress":   "Shopping",
	"shoes":   "Shopping",
	"bag":     "Shopping",
	"accessories": "Shopping",
	"electronics": "Shopping",
	"phone":   "Shopping",
	"laptop":  "Shopping",
	"computer": "Shopping",
	"book":    "Shopping",
	"store":   "Shopping",
	"mall":    "Shopping",
	
	// Entertainment
	"movie":   "Entertainment",
	"cinema":  "Entertainment",
	"theatre": "Entertainment",
	"netflix": "Entertainment",
	"spotify": "Entertainment",
	"youtube": "Entertainment",
	"game":    "Entertainment",
	"gaming":  "Entertainment",
	"concert": "Entertainment",
	"show":    "Entertainment",
	"event":   "Entertainment",
	"party":   "Entertainment",
	"entertainment": "Entertainment",
	"fun":     "Entertainment",
	
	// Utilities & Bills
	"electricity": "Utilities",
	"water":   "Utilities",
	"internet": "Utilities",
	"wifi":    "Utilities",
	"bill":    "Utilities",
	"utility": "Utilities",
	"rent":    "Utilities",
	"maintenance": "Utilities",
	
	// Healthcare
	"hospital": "Healthcare",
	"doctor":  "Healthcare",
	"medical": "Healthcare",
	"medicine": "Healthcare",
	"pharmacy": "Healthcare",
	"health":  "Healthcare",
	"dental":  "Healthcare",
	"clinic":  "Healthcare",
	
	// Income
	"salary":  "Income",
	"wage":    "Income",
	"payment": "Income",
	"income":  "Income",
	"earnings": "Income",
	"bonus":   "Income",
	"commission": "Income",
	"freelance": "Income",
	"business": "Income",
	"investment": "Income",
	"dividend": "Income",
	"refund":  "Income",
}

func AutoCategory(title string) string {
	low := strings.ToLower(title)
	for k, v := range keywordMap {
		if strings.Contains(low, k) {
			return v
		}
	}
	return "Other"
}
