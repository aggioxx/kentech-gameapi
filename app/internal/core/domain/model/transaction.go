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
	Currency              string  `json:"currency"`
	Amount                float64 `json:"amount"`
	ProviderTransactionID string  `json:"provider_transaction_id"`
	ProviderWithdrawnID   string  `json:"provider_withdrawn_id,omitempty"` // Only for deposit
}

type TransactionResponse struct {
	TransactionID         string  `json:"transaction_id"`
	ProviderTransactionID string  `json:"provider_transaction_id"`
	OldBalance            float64 `json:"old_balance"`
	NewBalance            float64 `json:"new_balance"`
	Status                string  `json:"status"` // WON/LOST for deposit, COMPLETED for withdraw
}
