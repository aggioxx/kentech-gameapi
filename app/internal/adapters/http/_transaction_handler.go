package http

//
//import (
//	"errors"
//	"kentech-project/internal/core/domain/model"
//	"kentech-project/internal/core/domain/service"
//	"net/http"
//
//	"github.com/gin-gonic/gin"
//	"github.com/google/uuid"
//)
//
//type TransactionHandler struct {
//	transactionService *service.TransactionService
//}
//
//func NewTransactionHandler(transactionService *service.TransactionService) *TransactionHandler {
//	return &TransactionHandler{transactionService: transactionService}
//}
//
//func (h *TransactionHandler) DepositGin(c *gin.Context) {
//	userID := getUserIDFromContext(c.Request.Context())
//	var req model.TransactionRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	response, err := h.transactionService.Deposit(c.Request.Context(), userID, req.Amount)
//	if err != nil {
//		switch {
//		case errors.Is(err, model.ErrInvalidAmount):
//			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		default:
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
//		}
//		return
//	}
//	c.JSON(http.StatusCreated, response)
//}
//
//func (h *TransactionHandler) WithdrawGin(c *gin.Context) {
//	userID := getUserIDFromContext(c.Request.Context())
//	var req model.TransactionRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	response, err := h.transactionService.Withdraw(c.Request.Context(), userID, req.Amount)
//
//	if err != nil {
//		switch err {
//		case model.ErrInvalidAmount, model.ErrInsufficientBalance:
//			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		default:
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
//		}
//		return
//	}
//	c.JSON(http.StatusCreated, response)
//}
//
//func (h *TransactionHandler) CancelGin(c *gin.Context) {
//	userID := getUserIDFromContext(c.Request.Context())
//	transactionIDStr := c.Param("id")
//	transactionID, err := uuid.Parse(transactionIDStr)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
//		return
//	}
//	err = h.transactionService.CancelTransaction(c.Request.Context(), userID, transactionID)
//	if err != nil {
//		switch err {
//		case model.ErrTransactionNotFound:
//			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
//		case model.ErrUnauthorized:
//			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
//		case model.ErrTransactionNotPending:
//			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		default:
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
//		}
//		return
//	}
//	c.Status(http.StatusNoContent)
//}
