package http

import (
	"context"
	"kentech-project/internal/core/domain/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PlayerHandler struct {
	playerService *service.PlayerService
}

func NewPlayerHandler(playerService *service.PlayerService) *PlayerHandler {
	return &PlayerHandler{playerService: playerService}
}

func (h *PlayerHandler) GetProfileGin(c *gin.Context) {
	userID := getUserIDFromContext(c.Request.Context())
	user, err := h.playerService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *PlayerHandler) GetTransactionHistoryGin(c *gin.Context) {
	userID := getUserIDFromContext(c.Request.Context())
	transactions, err := h.playerService.GetTransactionHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

func (h *PlayerHandler) GetBalanceGin(c *gin.Context) {
	userID := getUserIDFromContext(c.Request.Context())
	balance, err := h.playerService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func getUserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}
