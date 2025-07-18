package service

//
//import (
//	"context"
//	"database/sql"
//	"kentech-project/internal/core/domain/model"
//
//	"github.com/google/uuid"
//	"go.opentelemetry.io/otel"
//	"go.opentelemetry.io/otel/attribute"
//	"go.opentelemetry.io/otel/trace"
//	"kentech-project/internal/core/port"
//)
//
//type TransactionService struct {
//	userRepo      port.UserRepository
//	txRepo        port.TransactionRepository
//	walletService port.WalletService
//	db            *sql.DB
//}
//
//func NewTransactionService(userRepo port.UserRepository, txRepo port.TransactionRepository, walletService port.WalletService, db *sql.DB) *TransactionService {
//	return &TransactionService{
//		userRepo:      userRepo,
//		txRepo:        txRepo,
//		walletService: walletService,
//		db:            db,
//	}
//}
//
//func (s *TransactionService) Deposit(ctx context.Context, userID uuid.UUID, amount float64) (*model.TransactionResponse, error) {
//	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.Deposit", trace.WithAttributes(
//		attribute.String("user_id", userID.String()),
//		attribute.Float64("amount", amount),
//	))
//	defer span.End()
//
//	if amount <= 0 {
//		return nil, model.ErrInvalidAmount
//	}
//
//	user, err := s.userRepo.GetByID(ctx, userID)
//	if err != nil {
//		return nil, err
//	}
//
//	transaction := &model.Transaction{
//		UserID: userID,
//		Type:   model.TransactionTypeDeposit,
//		Amount: amount,
//		Status: model.TransactionStatusPending,
//	}
//
//	if err := s.txRepo.Create(ctx, transaction); err != nil {
//		return nil, err
//	}
//
//	reference, err := s.walletService.ProcessDeposit(ctx, userID, amount)
//	if err != nil {
//		err := s.txRepo.UpdateStatus(ctx, transaction.ID, model.TransactionStatusFailed)
//		if err != nil {
//			return nil, err
//		}
//		return nil, err
//	}
//
//	tx, err := s.db.BeginTx(ctx, nil)
//	if err != nil {
//		return nil, err
//	}
//	defer func(tx *sql.Tx) {
//		_ = tx.Rollback()
//	}(tx)
//
//	transaction.Reference = reference
//	transaction.Status = model.TransactionStatusCompleted
//	if err := s.txRepo.Update(ctx, transaction); err != nil {
//		return nil, err
//	}
//
//	newBalance := user.Balance + amount
//	if err := s.userRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
//		return nil, err
//	}
//
//	if err := tx.Commit(); err != nil {
//		return nil, err
//	}
//
//	return &model.TransactionResponse{
//		Transaction: *transaction,
//		Balance:     newBalance,
//	}, nil
//}
//
//func (s *TransactionService) Withdraw(ctx context.Context, userID uuid.UUID, amount float64) (*model.TransactionResponse, error) {
//	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.Withdraw", trace.WithAttributes(
//		attribute.String("user_id", userID.String()),
//		attribute.Float64("amount", amount),
//	))
//	defer span.End()
//
//	if amount <= 0 {
//		return nil, model.ErrInvalidAmount
//	}
//
//	user, err := s.userRepo.GetByID(ctx, userID)
//	if err != nil {
//		return nil, err
//	}
//
//	if user.Balance < amount {
//		return nil, model.ErrInsufficientBalance
//	}
//
//	transaction := &model.Transaction{
//		UserID: userID,
//		Type:   model.TransactionTypeWithdraw,
//		Amount: amount,
//		Status: model.TransactionStatusPending,
//	}
//
//	if err := s.txRepo.Create(ctx, transaction); err != nil {
//		return nil, err
//	}
//
//	reference, err := s.walletService.ProcessWithdraw(ctx, userID, amount)
//	if err != nil {
//		err := s.txRepo.UpdateStatus(ctx, transaction.ID, model.TransactionStatusFailed)
//		if err != nil {
//			return nil, err
//		}
//		return nil, err
//	}
//
//	tx, err := s.db.BeginTx(ctx, nil)
//	if err != nil {
//		return nil, err
//	}
//	defer func(tx *sql.Tx) {
//		_ = tx.Rollback()
//	}(tx)
//
//	transaction.Reference = reference
//	transaction.Status = model.TransactionStatusCompleted
//	if err := s.txRepo.Update(ctx, transaction); err != nil {
//		return nil, err
//	}
//
//	newBalance := user.Balance - amount
//	if err := s.userRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
//		return nil, err
//	}
//
//	if err := tx.Commit(); err != nil {
//		return nil, err
//	}
//
//	return &model.TransactionResponse{
//		Transaction: *transaction,
//		Balance:     newBalance,
//	}, nil
//}
//
//func (s *TransactionService) CancelTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
//	ctx, span := otel.Tracer("").Start(ctx, "TransactionService.CancelTransaction", trace.WithAttributes(
//		attribute.String("user_id", userID.String()),
//		attribute.String("transaction_id", transactionID.String()),
//	))
//	defer span.End()
//
//	transaction, err := s.txRepo.GetByID(ctx, transactionID)
//	if err != nil {
//		return err
//	}
//
//	if transaction.UserID != userID {
//		return model.ErrUnauthorized
//	}
//
//	if transaction.Status != model.TransactionStatusPending {
//		return model.ErrTransactionNotPending
//	}
//
//	if transaction.Reference != "" {
//		if err := s.walletService.CancelTransaction(ctx, transaction.Reference); err != nil {
//			return err
//		}
//	}
//
//	return s.txRepo.UpdateStatus(ctx, transactionID, model.TransactionStatusCanceled)
//}
