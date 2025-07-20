package service

import (
	"context"
	"database/sql"
	"kentech-project/internal/core/domain/model"
	"kentech-project/pkg/logger"

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

func (s *TransactionService) Deposit(ctx context.Context, userID uuid.UUID, amount float64) (*model.TransactionResponse, error) {
	s.logger.Debugf("Deposit called: user_id=%s, amount=%f", userID.String(), amount)

	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.Deposit", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Float64("amount", amount),
	))
	defer span.End()

	if amount <= 0 {
		s.logger.Warnf("Deposit failed: invalid amount %f for user_id=%s", amount, userID.String())
		return nil, model.ErrInvalidAmount
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Deposit failed: user not found or repo error: " + err.Error())
		return nil, err
	}

	transaction := &model.Transaction{
		UserID: userID,
		Type:   model.TransactionTypeDeposit,
		Amount: amount,
		Status: model.TransactionStatusPending,
	}
	s.logger.Debugf("Creating deposit transaction for user_id=%s, amount=%f", userID.String(), amount)
	if err := s.txRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Deposit failed: transaction creation error: " + err.Error())
		return nil, err
	}

	s.logger.Debugf("Calling walletService.ProcessDeposit for user_id=%s, amount=%f", userID.String(), amount)
	reference, err := s.walletService.ProcessDeposit(ctx, userID, amount)
	if err != nil {
		s.logger.Warnf("Deposit walletService failed for user_id=%s: %s", userID.String(), err.Error())
		updateErr := s.txRepo.UpdateStatus(ctx, transaction.ID, model.TransactionStatusFailed)
		if updateErr != nil {
			s.logger.Error("Deposit failed: could not update transaction status to failed: " + updateErr.Error())
			return nil, updateErr
		}
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Deposit failed: could not begin DB transaction: " + err.Error())
		return nil, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	transaction.Reference = reference
	transaction.Status = model.TransactionStatusCompleted
	s.logger.Debugf("Updating transaction status to completed for transaction_id=%s", transaction.ID.String())
	if err := s.txRepo.Update(ctx, transaction); err != nil {
		s.logger.Error("Deposit failed: could not update transaction: " + err.Error())
		return nil, err
	}

	newBalance := user.Balance + amount
	s.logger.Debugf("Updating user balance for user_id=%s: new_balance=%f", userID.String(), newBalance)
	if err := s.userRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("Deposit failed: could not update user balance: " + err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Deposit failed: could not commit DB transaction: " + err.Error())
		return nil, err
	}

	s.logger.Infof("Deposit successful: user_id=%s, transaction_id=%s, amount=%f, new_balance=%f", userID.String(), transaction.ID.String(), amount, newBalance)
	return &model.TransactionResponse{
		Transaction: *transaction,
		Balance:     newBalance,
	}, nil
}

func (s *TransactionService) Withdraw(ctx context.Context, userID uuid.UUID, amount float64) (*model.TransactionResponse, error) {
	s.logger.Debugf("Withdraw called: user_id=%s, amount=%f", userID.String(), amount)

	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.Withdraw", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Float64("amount", amount),
	))
	defer span.End()

	if amount <= 0 {
		s.logger.Warnf("Withdraw failed: invalid amount %f for user_id=%s", amount, userID.String())
		return nil, model.ErrInvalidAmount
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Withdraw failed: user not found or repo error: " + err.Error())
		return nil, err
	}

	if user.Balance < amount {
		s.logger.Warnf("Withdraw failed: insufficient balance for user_id=%s, balance=%f, amount=%f", userID.String(), user.Balance, amount)
		return nil, model.ErrInsufficientBalance
	}

	transaction := &model.Transaction{
		UserID: userID,
		Type:   model.TransactionTypeWithdraw,
		Amount: amount,
		Status: model.TransactionStatusPending,
	}
	s.logger.Debugf("Creating withdraw transaction for user_id=%s, amount=%f", userID.String(), amount)
	if err := s.txRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Withdraw failed: transaction creation error: " + err.Error())
		return nil, err
	}

	s.logger.Debugf("Calling walletService.ProcessWithdraw for user_id=%s, amount=%f", userID.String(), amount)
	reference, err := s.walletService.ProcessWithdraw(ctx, userID, amount)
	if err != nil {
		s.logger.Warnf("Withdraw walletService failed for user_id=%s: %s", userID.String(), err.Error())
		updateErr := s.txRepo.UpdateStatus(ctx, transaction.ID, model.TransactionStatusFailed)
		if updateErr != nil {
			s.logger.Error("Withdraw failed: could not update transaction status to failed: " + updateErr.Error())
			return nil, updateErr
		}
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Withdraw failed: could not begin DB transaction: " + err.Error())
		return nil, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	transaction.Reference = reference
	transaction.Status = model.TransactionStatusCompleted
	s.logger.Debugf("Updating transaction status to completed for transaction_id=%s", transaction.ID.String())
	if err := s.txRepo.Update(ctx, transaction); err != nil {
		s.logger.Error("Withdraw failed: could not update transaction: " + err.Error())
		return nil, err
	}

	newBalance := user.Balance - amount
	s.logger.Debugf("Updating user balance for user_id=%s: new_balance=%f", userID.String(), newBalance)
	if err := s.userRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("Withdraw failed: could not update user balance: " + err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Withdraw failed: could not commit DB transaction: " + err.Error())
		return nil, err
	}

	s.logger.Infof("Withdraw successful: user_id=%s, transaction_id=%s, amount=%f, new_balance=%f", userID.String(), transaction.ID.String(), amount, newBalance)
	return &model.TransactionResponse{
		Transaction: *transaction,
		Balance:     newBalance,
	}, nil
}

func (s *TransactionService) CancelTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	s.logger.Debugf("CancelTransaction called: user_id=%s, transaction_id=%s", userID.String(), transactionID.String())
	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.CancelTransaction", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("transaction_id", transactionID.String()),
	))
	defer span.End()

	transaction, err := s.txRepo.GetByID(ctx, transactionID)
	if err != nil {
		s.logger.Error("CancelTransaction failed: could not fetch transaction: " + err.Error())
		return err
	}

	if transaction.UserID != userID {
		s.logger.Warnf("CancelTransaction failed: unauthorized user_id=%s for transaction_id=%s", userID.String(), transactionID.String())
		return model.ErrUnauthorized
	}

	if transaction.Status != model.TransactionStatusPending {
		s.logger.Warnf("CancelTransaction failed: transaction not pending for transaction_id=%s", transactionID.String())
		return model.ErrTransactionNotPending
	}

	if transaction.Reference != "" {
		s.logger.Debugf("Calling walletService.CancelTransaction for reference=%s", transaction.Reference)
		if err := s.walletService.CancelTransaction(ctx, transaction.Reference); err != nil {
			s.logger.Error("CancelTransaction failed: walletService error: " + err.Error())
			return err
		}
	}

	s.logger.Debugf("Updating transaction status to canceled for transaction_id=%s", transactionID.String())
	err = s.txRepo.UpdateStatus(ctx, transactionID, model.TransactionStatusCanceled)
	if err != nil {
		s.logger.Error("CancelTransaction failed: could not update transaction status: " + err.Error())
		return err
	}

	s.logger.Infof("CancelTransaction successful: transaction_id=%s canceled by user_id=%s", transactionID.String(), userID.String())
	return nil
}
