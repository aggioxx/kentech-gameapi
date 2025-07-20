package http

import (
	"errors"
	"kentech-project/internal/adapters/repository/wallet"
	"kentech-project/internal/core/domain/model"
	"kentech-project/internal/core/domain/service"
	"kentech-project/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
	logger             *logger.Logger
}

func NewTransactionHandler(transactionService *service.TransactionService, log *logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:             log,
	}
}

func (h *TransactionHandler) DepositGin(c *gin.Context) {
	h.logger.Debug("Deposit endpoint called")

	userID := getUserIDFromContext(c.Request.Context())
	var req model.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body for deposit")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "code": "INVALID_BODY"})
		return
	}
	h.logger.Infof("Processing deposit: user_id=%s, amount=%f", userID.String(), req.Amount)

	response, err := h.transactionService.Deposit(
		c.Request.Context(),
		userID,
		req.Currency,
		req.Amount,
		req.ProviderTransactionID,
		req.ProviderWithdrawnID,
	)
	if err != nil {
		var walletErr *wallet.WalletError
		if errors.As(err, &walletErr) {
			h.logger.Warnf("Wallet error during deposit: %s", walletErr.Message)
			c.JSON(walletErr.StatusCode, gin.H{
				"error": walletErr.Message,
				"code":  "WALLET_ERROR",
			})
			return
		}
		switch {
		case errors.Is(err, model.ErrInvalidAmount):
			h.logger.Warnf("Invalid deposit amount: %f", req.Amount)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_AMOUNT",
			})
		default:
			h.logger.Error("Internal error during deposit: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
		}
		return
	}
	h.logger.Info("Deposit successful")
	c.JSON(http.StatusCreated, response)
}

func (h *TransactionHandler) WithdrawGin(c *gin.Context) {
	h.logger.Debug("Withdraw endpoint called")

	userID := getUserIDFromContext(c.Request.Context())
	var req model.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body for withdraw")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "code": "INVALID_BODY"})
		return
	}
	h.logger.Infof("Processing withdraw: user_id=%s, amount=%f", userID.String(), req.Amount)

	response, err := h.transactionService.Withdraw(
		c.Request.Context(),
		userID,
		req.Currency,
		req.Amount,
		req.ProviderTransactionID,
	)
	if err != nil {
		var walletErr *wallet.WalletError
		if errors.As(err, &walletErr) {
			h.logger.Warnf("Wallet error during withdraw: %s", walletErr.Message)
			c.JSON(walletErr.StatusCode, gin.H{
				"error": walletErr.Message,
				"code":  "WALLET_ERROR",
			})
			return
		}
		switch {
		case errors.Is(err, model.ErrInvalidAmount):
			h.logger.Warnf("Invalid withdraw amount: %f", req.Amount)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_AMOUNT",
			})
		case errors.Is(err, model.ErrInsufficientBalance):
			h.logger.Warnf("Insufficient balance for user_id=%s, amount=%f", userID.String(), req.Amount)
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
				"code":  "INSUFFICIENT_BALANCE",
			})
		default:
			h.logger.Error("Internal error during withdraw: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
		}
		return
	}
	h.logger.Info("Withdraw successful")
	c.JSON(http.StatusCreated, response)
}

func (h *TransactionHandler) CancelGin(c *gin.Context) {
	h.logger.Debug("CancelTransaction endpoint called")

	userID := getUserIDFromContext(c.Request.Context())
	transactionIDStr := c.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		h.logger.Warnf("Invalid transaction ID: %s", transactionIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}
	h.logger.Infof("Processing cancel: user_id=%s, transaction_id=%s", userID.String(), transactionID.String())

	response, err := h.transactionService.CancelTransaction(c.Request.Context(), userID, transactionID)
	if err != nil {
		switch err {
		case model.ErrTransactionNotFound:
			h.logger.Warnf("Transaction not found: %s", transactionID.String())
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case model.ErrUnauthorized:
			h.logger.Warnf("Unauthorized cancel attempt: user_id=%s, transaction_id=%s", userID.String(), transactionID.String())
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case model.ErrTransactionNotPending:
			h.logger.Warnf("Transaction not pending: %s", transactionID.String())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			h.logger.Error("Internal error during cancel: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	h.logger.Info("Transaction canceled successfully")
	c.JSON(http.StatusOK, response)
}
