package port

import (
	"context"

	"github.com/google/uuid"
)

type WalletService interface {
	ProcessDeposit(ctx context.Context, userID uuid.UUID, amount float64) (string, error)
	ProcessWithdraw(ctx context.Context, userID uuid.UUID, amount float64) (string, error)
	CancelTransaction(ctx context.Context, reference string) error
}
