package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-github/expense-tracker-backend/config"
	"github.com/your-github/expense-tracker-backend/internal/core/domain"
	"github.com/your-github/expense-tracker-backend/internal/core/ports"
	"github.com/your-github/expense-tracker-backend/internal/core/services"
	"github.com/your-github/expense-tracker-backend/internal/repo"
	"go.uber.org/zap"
)

// AAHandler handles Account Aggregator operations
type AAHandler struct {
	aaService    *services.AAService
	repositories *repo.Repositories
	config       *config.Config
	logger       *zap.Logger
}

// NewAAHandler creates a new AA handler
func NewAAHandler(
	aaService *services.AAService,
	repositories *repo.Repositories,
	config *config.Config,
	logger *zap.Logger,
) *AAHandler {
	return &AAHandler{
		aaService:    aaService,
		repositories: repositories,
		config:       config,
		logger:       logger,
	}
}

// InitiateConsentRequest represents a consent initiation request
type InitiateConsentRequest struct {
	FIType    string    `json:"fi_type" binding:"required"`    // "SAVINGS", "CURRENT", etc.
	Purpose   string    `json:"purpose" binding:"required"`    // "EXPENSE_ANALYSIS"
	DateRange DateRange `json:"date_range" binding:"required"` // ISO dates
	Frequency string    `json:"frequency" binding:"required"`  // "DAILY"
}

// DateRange represents a date range
type DateRange struct {
	From string `json:"from" binding:"required"` // ISO date
	To   string `json:"to" binding:"required"`   // ISO date
}

// InitiateConsentResponse represents a consent initiation response
type InitiateConsentResponse struct {
	BankLinkID  string `json:"bank_link_id"`
	ConsentID   string `json:"consent_id"`
	RedirectURL string `json:"redirect_url"`
	Status      string `json:"status"`
}

// InitiateConsent initiates a new consent request
// @Summary Initiate AA consent
// @Description Initiate a new Account Aggregator consent request
// @Tags aa
// @Accept json
// @Produce json
// @Param request body InitiateConsentRequest true "Consent details"
// @Success 200 {object} InitiateConsentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /aa/consents/initiate [post]
func (h *AAHandler) InitiateConsent(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req InitiateConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Create consent request
	consentReq := ports.ConsentRequest{
		UserID:      userID.String(),
		FIType:      req.FIType,
		Purpose:     req.Purpose,
		DateRange:   ports.DateRange{From: req.DateRange.From, To: req.DateRange.To},
		Frequency:   req.Frequency,
		RedirectURL: h.config.AA.BaseURL + "/consent/callback",
		WebhookURL:  h.config.AA.BaseURL + "/webhook",
	}

	// Initiate consent
	bankLink, err := h.aaService.InitiateConsent(c.Request.Context(), userID, consentReq)
	if err != nil {
		h.logger.Error("Failed to initiate consent", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to initiate consent"})
		return
	}

	c.JSON(http.StatusOK, InitiateConsentResponse{
		BankLinkID:  bankLink.ID.String(),
		ConsentID:   bankLink.AAConsentID,
		RedirectURL: h.config.AA.BaseURL + "/consent/" + bankLink.AAConsentID,
		Status:      bankLink.Status,
	})
}

// ConsentCallbackRequest represents a consent callback request
type ConsentCallbackRequest struct {
	ConsentID string `json:"consent_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// ConsentCallback handles consent status updates from AA
// @Summary Handle consent callback
// @Description Handle consent status updates from Account Aggregator
// @Tags aa
// @Accept json
// @Produce json
// @Param request body ConsentCallbackRequest true "Callback data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /aa/consents/callback [post]
func (h *AAHandler) ConsentCallback(c *gin.Context) {
	var req ConsentCallbackRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Read request body for signature verification
	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to read request body"})
		return
	}

	// Verify webhook signature
	if !h.verifyWebhookSignature(bodyBytes, req.Signature) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid signature"})
		return
	}

	// Handle consent status update
	err = h.aaService.HandleConsentCallback(c.Request.Context(), req.ConsentID, req.Status)
	if err != nil {
		h.logger.Error("Failed to handle consent callback", zap.Error(err), zap.String("consent_id", req.ConsentID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to handle callback"})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// FetchTransactionsRequest represents a transaction fetch request
type FetchTransactionsRequest struct {
	BankLinkID string `json:"bank_link_id" binding:"required"`
	FromDate   string `json:"from_date" binding:"required"` // ISO date
	ToDate     string `json:"to_date" binding:"required"`   // ISO date
}

// FetchTransactionsResponse represents a transaction fetch response
type FetchTransactionsResponse struct {
	SessionID    string                `json:"session_id"`
	Status       string                `json:"status"`
	Processed    bool                  `json:"processed"`
	Transactions []*domain.Transaction `json:"transactions,omitempty"`
}

// FetchTransactions fetches transactions for a bank link
// @Summary Fetch transactions
// @Description Fetch transactions for a bank link via AA
// @Tags aa
// @Accept json
// @Produce json
// @Param request body FetchTransactionsRequest true "Fetch details"
// @Success 200 {object} FetchTransactionsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /aa/fetch [post]
func (h *AAHandler) FetchTransactions(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req FetchTransactionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Parse bank link ID
	bankLinkID, err := uuid.Parse(req.BankLinkID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid bank link ID"})
		return
	}

	// Fetch transactions
	result, err := h.aaService.FetchTransactions(c.Request.Context(), userID, bankLinkID, req.FromDate, req.ToDate)
	if err != nil {
		h.logger.Error("Failed to fetch transactions", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, FetchTransactionsResponse{
		SessionID:    result.SessionID,
		Status:       result.Status,
		Processed:    result.Processed,
		Transactions: result.Transactions,
	})
}

// DataReadyWebhookRequest represents a data ready webhook request
type DataReadyWebhookRequest struct {
	EventType string `json:"event_type" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// DataReadyWebhook handles data ready webhook from AA
// @Summary Handle data ready webhook
// @Description Handle data ready webhook from Account Aggregator
// @Tags aa
// @Accept json
// @Produce json
// @Param request body DataReadyWebhookRequest true "Webhook data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /aa/webhook [post]
func (h *AAHandler) DataReadyWebhook(c *gin.Context) {
	var req DataReadyWebhookRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Read request body for signature verification
	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to read request body"})
		return
	}

	// Verify webhook signature
	if !h.verifyWebhookSignature(bodyBytes, req.Signature) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid signature"})
		return
	}

	// Handle data ready webhook
	err = h.aaService.HandleDataReadyWebhook(c.Request.Context(), req.SessionID)
	if err != nil {
		h.logger.Error("Failed to handle data ready webhook", zap.Error(err), zap.String("session_id", req.SessionID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to handle webhook"})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// GetBankLinksResponse represents a bank links response
type GetBankLinksResponse struct {
	BankLinks []*domain.BankLink `json:"bank_links"`
}

// GetBankLinks returns active bank links for the user
// @Summary Get bank links
// @Description Get active bank links for the authenticated user
// @Tags aa
// @Produce json
// @Success 200 {object} GetBankLinksResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /aa/bank-links [get]
func (h *AAHandler) GetBankLinks(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	bankLinks, err := h.aaService.GetActiveBankLinks(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get bank links", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get bank links"})
		return
	}

	c.JSON(http.StatusOK, GetBankLinksResponse{BankLinks: bankLinks})
}

// RevokeConsentRequest represents a consent revocation request
type RevokeConsentRequest struct {
	BankLinkID string `json:"bank_link_id" binding:"required"`
}

// RevokeConsent revokes an active consent
// @Summary Revoke consent
// @Description Revoke an active bank link consent
// @Tags aa
// @Accept json
// @Produce json
// @Param request body RevokeConsentRequest true "Revocation details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /aa/consents/revoke [post]
func (h *AAHandler) RevokeConsent(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req RevokeConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Parse bank link ID
	bankLinkID, err := uuid.Parse(req.BankLinkID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid bank link ID"})
		return
	}

	// Revoke consent
	err = h.aaService.RevokeConsent(c.Request.Context(), userID, bankLinkID)
	if err != nil {
		h.logger.Error("Failed to revoke consent", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to revoke consent"})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// verifyWebhookSignature verifies webhook signature
func (h *AAHandler) verifyWebhookSignature(payload []byte, signature string) bool {
	// In production, implement proper HMAC verification
	// For now, return true for mock implementation
	// TODO: Implement proper signature verification using h.config.Webhook.Secret
	_ = payload   // Use payload parameter
	_ = signature // Use signature parameter
	return true
}

// getUserIDFromContext extracts user ID from JWT context
func getUserIDFromContext(c *gin.Context) uuid.UUID {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil
	}

	return userID
}
