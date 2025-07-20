package port

import (
	"context"
	"kentech-project/internal/adapters/repository/wallet"
)

type WalletService interface {
	ProcessDeposit(ctx context.Context, userID int, amount float64, currency string, betID int, reference string) (wallet.OperationResponse, error)
	ProcessWithdraw(ctx context.Context, userID int, amount float64, currency string, betID int, reference string) (wallet.OperationResponse, error)
	CancelTransaction(ctx context.Context, reference string) error
}
