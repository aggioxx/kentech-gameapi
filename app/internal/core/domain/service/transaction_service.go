package service

import (
	"context"
	"database/sql"
	"kentech-project/internal/core/domain/model"
	"kentech-project/pkg/logger"
	"strconv"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"kentech-project/internal/core/port"
)

type TransactionService struct {
	userRepo      port.UserRepository
	txRepo        port.TransactionRepository
	walletService port.WalletService
	db            *sql.DB
	logger        *logger.Logger
}

func NewTransactionService(userRepo port.UserRepository,
	txRepo port.TransactionRepository,
	walletService port.WalletService,
	db *sql.DB,
	log *logger.Logger) *TransactionService {
	return &TransactionService{
		userRepo:      userRepo,
		txRepo:        txRepo,
		walletService: walletService,
		db:            db,
		logger:        log,
	}
}

func (s *TransactionService) Deposit(ctx context.Context, userID uuid.UUID, currency string, amount float64, providerTxID, providerWithdrawnID string) (*model.TransactionResponse, error) {
	s.logger.Debugf("Deposit called: user_id=%s, amount=%f, currency=%s, providerTxID=%s, providerWithdrawnID=%s", userID.String(), amount, currency, providerTxID, providerWithdrawnID)

	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.Deposit", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Float64("amount", amount),
		attribute.String("currency", currency),
		attribute.String("provider_tx_id", providerTxID),
	))
	defer span.End()

	if amount < 0 {
		s.logger.Warnf("Deposit failed: invalid amount %f for user_id=%s", amount, userID.String())
		return nil, model.ErrInvalidAmount
	}

	s.logger.Debugf("Fetching user by ID: %s", userID.String())
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Deposit failed: user not found or repo error: " + err.Error())
		return nil, err
	}

	oldBalance := user.Balance
	s.logger.Debugf("User found. Old balance: %f", oldBalance)

	transaction := &model.Transaction{
		UserID:    userID,
		Type:      model.TransactionTypeDeposit,
		Amount:    amount,
		Status:    model.TransactionStatusPending,
		Reference: providerTxID,
	}
	s.logger.Debug("Creating deposit transaction record")
	if err := s.txRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Deposit failed: transaction creation error: " + err.Error())
		return nil, err
	}

	s.logger.Info("Calling wallet service for deposit")

	walletUserID, err := s.getWalletUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get wallet user ID: " + err.Error())
		return nil, err
	}
	walletResp, err := s.walletService.ProcessDeposit(ctx, walletUserID, amount, currency, 0, providerTxID)
	if err != nil {
		s.logger.Error("Wallet service deposit failed: " + err.Error())
		err := s.txRepo.UpdateStatus(ctx, transaction.ID, model.TransactionStatusFailed)
		if err != nil {
			s.logger.Error("Failed to update transaction status to failed: " + err.Error())
			return nil, err
		}
		return nil, err
	}

	newBalance, err := strconv.ParseFloat(walletResp.Balance, 64)
	if err != nil {
		s.logger.Error("Failed to parse wallet balance: " + err.Error())
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin DB transaction: " + err.Error())
		return nil, err
	}
	defer func(tx *sql.Tx) { _ = tx.Rollback() }(tx)

	transaction.Status = model.TransactionStatusCompleted
	s.logger.Debug("Updating transaction status to completed")
	if err := s.txRepo.Update(ctx, transaction); err != nil {
		s.logger.Error("Failed to update transaction status: " + err.Error())
		return nil, err
	}

	s.logger.Debugf("Updating user balance to: %f", newBalance)
	if err := s.userRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("Failed to update user balance: " + err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit DB transaction: " + err.Error())
		return nil, err
	}

	status := "LOST"
	if amount > 0 {
		status = "WON"
	}
	s.logger.Infof("Deposit successful: user_id=%s, transaction_id=%s, status=%s", userID.String(), transaction.ID.String(), status)

	return &model.TransactionResponse{
		TransactionID:         transaction.ID.String(),
		ProviderTransactionID: providerTxID,
		OldBalance:            oldBalance,
		NewBalance:            newBalance,
		Status:                status,
	}, nil
}

func (s *TransactionService) Withdraw(ctx context.Context, userID uuid.UUID, currency string, amount float64, providerTxID string) (*model.TransactionResponse, error) {
	s.logger.Debugf("Withdraw called: user_id=%s, amount=%f, currency=%s, providerTxID=%s", userID.String(), amount, currency, providerTxID)

	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.Withdraw", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Float64("amount", amount),
		attribute.String("currency", currency),
		attribute.String("provider_tx_id", providerTxID),
	))
	defer span.End()

	if amount <= 0 {
		s.logger.Warnf("Withdraw failed: invalid amount %f for user_id=%s", amount, userID.String())
		return nil, model.ErrInvalidAmount
	}

	s.logger.Debugf("Fetching user by ID: %s", userID.String())
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Withdraw failed: user not found or repo error: " + err.Error())
		return nil, err
	}

	oldBalance := user.Balance
	s.logger.Debugf("User found. Old balance: %f", oldBalance)
	if oldBalance < amount {
		s.logger.Warnf("Withdraw failed: insufficient balance for user_id=%s, requested=%f, available=%f", userID.String(), amount, oldBalance)
		return nil, model.ErrInsufficientBalance
	}

	transaction := &model.Transaction{
		UserID:    userID,
		Type:      model.TransactionTypeWithdraw,
		Amount:    amount,
		Status:    model.TransactionStatusPending,
		Reference: providerTxID,
	}
	s.logger.Debug("Creating withdraw transaction record")
	if err := s.txRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Withdraw failed: transaction creation error: " + err.Error())
		return nil, err
	}

	s.logger.Info("Calling wallet service for withdraw")

	walletUserID, err := s.getWalletUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get wallet user ID: " + err.Error())
		return nil, err
	}
	walletResp, err := s.walletService.ProcessWithdraw(ctx, walletUserID, amount, currency, 0, providerTxID)
	if err != nil {
		s.logger.Error("Wallet service withdraw failed: " + err.Error())
		err := s.txRepo.UpdateStatus(ctx, transaction.ID, model.TransactionStatusFailed)
		if err != nil {
			s.logger.Error("Failed to update transaction status to failed: " + err.Error())
			return nil, err
		}
		return nil, err
	}

	newBalance, err := strconv.ParseFloat(walletResp.Balance, 64)
	if err != nil {
		s.logger.Error("Failed to parse wallet balance: " + err.Error())
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin DB transaction: " + err.Error())
		return nil, err
	}
	defer func(tx *sql.Tx) { _ = tx.Rollback() }(tx)

	transaction.Status = model.TransactionStatusCompleted
	s.logger.Debug("Updating transaction status to completed")
	if err := s.txRepo.Update(ctx, transaction); err != nil {
		s.logger.Error("Failed to update transaction status: " + err.Error())
		return nil, err
	}

	s.logger.Debugf("Updating user balance to: %f", newBalance)
	if err := s.userRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("Failed to update user balance: " + err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit DB transaction: " + err.Error())
		return nil, err
	}

	s.logger.Infof("Withdraw successful: user_id=%s, transaction_id=%s", userID.String(), transaction.ID.String())

	return &model.TransactionResponse{
		TransactionID:         transaction.ID.String(),
		ProviderTransactionID: providerTxID,
		OldBalance:            oldBalance,
		NewBalance:            newBalance,
		Status:                "COMPLETED",
	}, nil
}

func (s *TransactionService) CancelTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*model.TransactionResponse, error) {
	s.logger.Debugf("CancelTransaction called: user_id=%s, transaction_id=%s", userID.String(), transactionID.String())
	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.CancelTransaction", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("transaction_id", transactionID.String()),
	))
	defer span.End()

	transaction, err := s.txRepo.GetByID(ctx, transactionID)
	if err != nil {
		s.logger.Error("CancelTransaction failed: could not fetch transaction: " + err.Error())
		return nil, model.ErrTransactionNotFound
	}

	if transaction.UserID != userID {
		s.logger.Warnf("CancelTransaction failed: unauthorized user_id=%s for transaction_id=%s", userID.String(), transactionID.String())
		return nil, model.ErrUnauthorized
	}

	if transaction.Status != model.TransactionStatusPending {
		s.logger.Warnf("CancelTransaction failed: transaction not pending for transaction_id=%s", transactionID.String())
		return nil, model.ErrTransactionNotPending
	}

	oldBalance := 0.0
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		oldBalance = user.Balance
	}

	if transaction.Reference != "" {
		s.logger.Debugf("Calling walletService.CancelTransaction for reference=%s", transaction.Reference)
		if err := s.walletService.CancelTransaction(ctx, transaction.Reference); err != nil {
			s.logger.Error("CancelTransaction failed: walletService error: " + err.Error())
			return nil, err
		}
	}

	s.logger.Debugf("Updating transaction status to canceled for transaction_id=%s", transactionID.String())
	err = s.txRepo.UpdateStatus(ctx, transactionID, model.TransactionStatusCanceled)
	if err != nil {
		s.logger.Error("CancelTransaction failed: could not update transaction status: " + err.Error())
		return nil, err
	}

	newBalance := oldBalance
	user, err = s.userRepo.GetByID(ctx, userID)
	if err == nil {
		newBalance = user.Balance
	}

	s.logger.Infof("CancelTransaction successful: transaction_id=%s canceled by user_id=%s", transactionID.String(), userID.String())
	return &model.TransactionResponse{
		TransactionID:         transaction.ID.String(),
		ProviderTransactionID: transaction.Reference,
		OldBalance:            oldBalance,
		NewBalance:            newBalance,
		Status:                string(model.TransactionStatusCanceled),
	}, nil
}

func (s *TransactionService) getWalletUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.WalletUserID, nil
}
