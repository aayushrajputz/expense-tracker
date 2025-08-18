package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BankVerificationService struct {
	APIKey   string
	APIURL   string
	Client   *http.Client
	Provider string // "karza", "signzy", "razorpay", "mock"
}

type BankVerificationRequest struct {
	BankID        string `json:"bank_id"`
	AccountNumber string `json:"account_number"`
	MobileNumber  string `json:"mobile_number"`
	AccountHolder string `json:"account_holder_name"`
}

type BankVerificationResponse struct {
	Success     bool   `json:"success"`
	Verified    bool   `json:"verified"`
	Message     string `json:"message"`
	AccountType string `json:"account_type,omitempty"`
	IFSCCode    string `json:"ifsc_code,omitempty"`
	BranchName  string `json:"branch_name,omitempty"`
	Provider    string `json:"provider,omitempty"`
}

type VerificationAPIRequest struct {
	BankID        string `json:"bank_id"`
	AccountNumber string `json:"account_number"`
	MobileNumber  string `json:"mobile_number"`
	AccountHolder string `json:"account_holder_name"`
	RequestID     string `json:"request_id"`
	Timestamp     string `json:"timestamp"`
}

type VerificationAPIResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	Verified    bool   `json:"verified"`
	AccountType string `json:"account_type,omitempty"`
	IFSCCode    string `json:"ifsc_code,omitempty"`
	BranchName  string `json:"branch_name,omitempty"`
}

// Karza API specific structures
type KarzaRequest struct {
	AccountNumber string `json:"accountNumber"`
	MobileNumber  string `json:"mobileNumber"`
	BankCode      string `json:"bankCode"`
}

type KarzaResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Verified    bool   `json:"verified"`
		AccountType string `json:"accountType"`
		IFSCCode    string `json:"ifscCode"`
		BranchName  string `json:"branchName"`
	} `json:"data"`
}

// Signzy API specific structures
type SignzyRequest struct {
	AccountNumber string `json:"account_number"`
	MobileNumber  string `json:"mobile_number"`
	BankID        string `json:"bank_id"`
}

type SignzyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Verified    bool   `json:"verified"`
		AccountType string `json:"account_type"`
		IFSCCode    string `json:"ifsc_code"`
		BranchName  string `json:"branch_name"`
	} `json:"result"`
}

func NewBankVerificationService(apiKey, apiURL, provider string) *BankVerificationService {
	return &BankVerificationService{
		APIKey:   apiKey,
		APIURL:   apiURL,
		Provider: provider,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VerifyBankAccount verifies if the account number is linked to the provided mobile number
func (s *BankVerificationService) VerifyBankAccount(req BankVerificationRequest) (*BankVerificationResponse, error) {
	switch s.Provider {
	case "karza":
		return s.verifyWithKarza(req)
	case "signzy":
		return s.verifyWithSignzy(req)
	case "razorpay":
		return s.verifyWithRazorpay(req)
	case "mock":
		fallthrough
	default:
		return s.MockVerifyBankAccount(req)
	}
}

// verifyWithKarza verifies using Karza API
func (s *BankVerificationService) verifyWithKarza(req BankVerificationRequest) (*BankVerificationResponse, error) {
	karzaReq := KarzaRequest{
		AccountNumber: req.AccountNumber,
		MobileNumber:  req.MobileNumber,
		BankCode:      req.BankID,
	}

	jsonData, err := json.Marshal(karzaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Karza request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.karza.in/v3/kyc/bank-account-verification", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Karza request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-karza-key", s.APIKey)

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make Karza request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Karza response: %v", err)
	}

	var karzaResp KarzaResponse
	if err := json.Unmarshal(body, &karzaResp); err != nil {
		return nil, fmt.Errorf("failed to parse Karza response: %v", err)
	}

	return &BankVerificationResponse{
		Success:     karzaResp.Status == "success",
		Verified:    karzaResp.Data.Verified,
		Message:     karzaResp.Message,
		AccountType: karzaResp.Data.AccountType,
		IFSCCode:    karzaResp.Data.IFSCCode,
		BranchName:  karzaResp.Data.BranchName,
		Provider:    "karza",
	}, nil
}

// verifyWithSignzy verifies using Signzy API
func (s *BankVerificationService) verifyWithSignzy(req BankVerificationRequest) (*BankVerificationResponse, error) {
	signzyReq := SignzyRequest{
		AccountNumber: req.AccountNumber,
		MobileNumber:  req.MobileNumber,
		BankID:        req.BankID,
	}

	jsonData, err := json.Marshal(signzyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Signzy request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.signzy.com/v2/bank-account-verification", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Signzy request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make Signzy request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Signzy response: %v", err)
	}

	var signzyResp SignzyResponse
	if err := json.Unmarshal(body, &signzyResp); err != nil {
		return nil, fmt.Errorf("failed to parse Signzy response: %v", err)
	}

	return &BankVerificationResponse{
		Success:     signzyResp.Status == "success",
		Verified:    signzyResp.Result.Verified,
		Message:     signzyResp.Message,
		AccountType: signzyResp.Result.AccountType,
		IFSCCode:    signzyResp.Result.IFSCCode,
		BranchName:  signzyResp.Result.BranchName,
		Provider:    "signzy",
	}, nil
}

// verifyWithRazorpay verifies using Razorpay API
func (s *BankVerificationService) verifyWithRazorpay(req BankVerificationRequest) (*BankVerificationResponse, error) {
	razorpayReq := map[string]string{
		"account_number": req.AccountNumber,
		"mobile_number":  req.MobileNumber,
		"bank_code":      req.BankID,
	}

	jsonData, err := json.Marshal(razorpayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Razorpay request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.razorpay.com/v1/bank-accounts/verify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+s.APIKey)

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make Razorpay request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Razorpay response: %v", err)
	}

	var razorpayResp map[string]interface{}
	if err := json.Unmarshal(body, &razorpayResp); err != nil {
		return nil, fmt.Errorf("failed to parse Razorpay response: %v", err)
	}

	verified := false
	if status, ok := razorpayResp["status"].(string); ok {
		verified = status == "verified"
	}

	return &BankVerificationResponse{
		Success:  true,
		Verified: verified,
		Message:  "Razorpay verification completed",
		Provider: "razorpay",
	}, nil
}

// Enhanced Mock verification for testing and development
func (s *BankVerificationService) MockVerifyBankAccount(req BankVerificationRequest) (*BankVerificationResponse, error) {
	// Simulate API delay
	time.Sleep(1 * time.Second)

	// Enhanced mock verification logic
	verified := s.validateMockVerification(req)

	var message string
	var accountType string
	var ifscCode string
	var branchName string

	if verified {
		message = "Account verification successful"
		accountType = "Savings"
		ifscCode = s.generateMockIFSC(req.BankID)
		branchName = "Main Branch"
	} else {
		message = "Account verification failed - invalid account number or mobile number"
		accountType = ""
		ifscCode = ""
		branchName = ""
	}

	return &BankVerificationResponse{
		Success:     true,
		Verified:    verified,
		Message:     message,
		AccountType: accountType,
		IFSCCode:    ifscCode,
		BranchName:  branchName,
		Provider:    "mock",
	}, nil
}

// validateMockVerification provides realistic mock validation
func (s *BankVerificationService) validateMockVerification(req BankVerificationRequest) bool {
	// Basic format validation
	if !ValidateAccountNumber(req.AccountNumber) || !ValidateMobileNumber(req.MobileNumber) {
		return false
	}

	// Mock bank-specific validation
	switch req.BankID {
	case "sbi":
		// SBI account numbers are typically 11 digits
		return len(req.AccountNumber) == 11
	case "hdfc":
		// HDFC account numbers are typically 12-14 digits
		return len(req.AccountNumber) >= 12 && len(req.AccountNumber) <= 14
	case "icici":
		// ICICI account numbers are typically 12 digits
		return len(req.AccountNumber) == 12
	case "axis":
		// Axis account numbers are typically 15 digits
		return len(req.AccountNumber) == 15
	case "airtel-payment":
		// Airtel Payment Bank account numbers are typically 10-12 digits
		return len(req.AccountNumber) >= 10 && len(req.AccountNumber) <= 12
	default:
		// For other banks, accept 9-18 digits
		return len(req.AccountNumber) >= 9 && len(req.AccountNumber) <= 18
	}
}

// generateMockIFSC generates realistic IFSC codes
func (s *BankVerificationService) generateMockIFSC(bankID string) string {
	bankCodes := map[string]string{
		"sbi":            "SBIN",
		"hdfc":           "HDFC",
		"icici":          "ICIC",
		"axis":           "UTIB",
		"kotak":          "KKBK",
		"yes":            "YESB",
		"airtel-payment": "AIRP",
	}

	bankCode := bankCodes[bankID]
	if bankCode == "" {
		bankCode = "BANK"
	}

	return fmt.Sprintf("%s0001234", bankCode)
}

// ValidateMobileNumber validates Indian mobile number format
func ValidateMobileNumber(mobile string) bool {
	if len(mobile) != 10 {
		return false
	}

	// Check if it starts with valid Indian mobile prefixes
	validPrefixes := []string{"6", "7", "8", "9"}
	for _, prefix := range validPrefixes {
		if mobile[0] == prefix[0] {
			return true
		}
	}
	return false
}

// ValidateAccountNumber validates account number format
func ValidateAccountNumber(accountNumber string) bool {
	// Basic validation - account number should be between 9-18 digits
	if len(accountNumber) < 9 || len(accountNumber) > 18 {
		return false
	}

	// Check if it contains only digits
	for _, char := range accountNumber {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
