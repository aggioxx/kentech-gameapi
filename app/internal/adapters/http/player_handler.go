package http

import (
	"context"
	"kentech-project/internal/core/domain/service"
	"kentech-project/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PlayerHandler struct {
	playerService *service.PlayerService
	logger        *logger.Logger
}

func NewPlayerHandler(playerService *service.PlayerService, log *logger.Logger) *PlayerHandler {
	return &PlayerHandler{
		playerService: playerService,
		logger:        log,
	}
}

func (h *PlayerHandler) GetProfileGin(c *gin.Context) {
	h.logger.Debug("GetProfile endpoint called")

	userID := getUserIDFromContext(c.Request.Context())
	h.logger.Infof("Fetching profile for user_id=%s", userID.String())
	user, err := h.playerService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Warnf("User not found: user_id=%s", userID.String())
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	h.logger.Info("Profile fetched successfully")
	c.JSON(http.StatusOK, user)
}

func (h *PlayerHandler) GetTransactionHistoryGin(c *gin.Context) {
	h.logger.Debug("GetTransactionHistory endpoint called")

	userID := getUserIDFromContext(c.Request.Context())
	h.logger.Infof("Fetching transaction history for user_id=%s", userID.String())
	transactions, err := h.playerService.GetTransactionHistory(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch transaction history: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	h.logger.Info("Transaction history fetched successfully")
	c.JSON(http.StatusOK, transactions)
}

func (h *PlayerHandler) GetBalanceGin(c *gin.Context) {
	h.logger.Debug("GetBalance endpoint called")

	userID := getUserIDFromContext(c.Request.Context())
	h.logger.Infof("Fetching balance for user_id=%s", userID.String())
	balance, err := h.playerService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch balance: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	h.logger.Info("Balance fetched successfully")
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func getUserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}
