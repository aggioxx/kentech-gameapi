package model

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusCanceled  TransactionStatus = "canceled"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	ID        uuid.UUID         `json:"id"`
	UserID    uuid.UUID         `json:"user_id"`
	Type      TransactionType   `json:"type"`
	Amount    float64           `json:"amount"`
	Status    TransactionStatus `json:"status"`
	Reference string            `json:"reference,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type TransactionRequest struct {
	Amount float64 `json:"amount"`
}
type TransactionResponse struct {
	Transaction Transaction `json:"transaction"`
	Balance     float64     `json:"balance"`
}
