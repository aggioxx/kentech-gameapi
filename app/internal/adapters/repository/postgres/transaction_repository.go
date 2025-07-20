package postgres

import (
	"context"
	"database/sql"
	"kentech-project/internal/core/domain/model"
	"kentech-project/pkg/logger"
	"time"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewTransactionRepository(db *sql.DB, log *logger.Logger) *TransactionRepository {
	return &TransactionRepository{
		db:     db,
		logger: log,
	}
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	r.logger.Debug("Creating new transaction")
	query := `
		INSERT INTO transactions (id, user_id, type, amount, status, reference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	transaction.ID = uuid.New()
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.UserID, transaction.Type, transaction.Amount,
		transaction.Status, transaction.Reference, transaction.CreatedAt, transaction.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create transaction: " + err.Error())
		return err
	}
	r.logger.Infof("Transaction created: id=%s, user_id=%s, amount=%f", transaction.ID.String(), transaction.UserID.String(), transaction.Amount)
	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	r.logger.Debugf("Fetching transaction by ID: %s", id.String())
	query := `
		SELECT id, user_id, type, amount, status, reference, created_at, updated_at
		FROM transactions WHERE id = $1
	`

	transaction := &model.Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Type, &transaction.Amount,
		&transaction.Status, &transaction.Reference, &transaction.CreatedAt, &transaction.UpdatedAt)

	if err == sql.ErrNoRows {
		r.logger.Warnf("Transaction not found: id=%s", id.String())
		return nil, model.ErrTransactionNotFound
	}
	if err != nil {
		r.logger.Error("Failed to fetch transaction: " + err.Error())
		return nil, err
	}
	r.logger.Infof("Transaction fetched: id=%s", transaction.ID.String())
	return transaction, nil
}

func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Transaction, error) {
	r.logger.Debugf("Fetching transactions for user_id: %s", userID.String())
	query := `
		SELECT id, user_id, type, amount, status, reference, created_at, updated_at
		FROM transactions WHERE user_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to query transactions: " + err.Error())
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		transaction := &model.Transaction{}
		err := rows.Scan(
			&transaction.ID, &transaction.UserID, &transaction.Type, &transaction.Amount,
			&transaction.Status, &transaction.Reference, &transaction.CreatedAt, &transaction.UpdatedAt)
		if err != nil {
			r.logger.Error("Failed to scan transaction row: " + err.Error())
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if rows.Err() != nil {
		r.logger.Error("Row iteration error: " + rows.Err().Error())
		return nil, rows.Err()
	}
	r.logger.Infof("Fetched %d transactions for user_id=%s", len(transactions), userID.String())
	return transactions, nil
}

func (r *TransactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	r.logger.Debugf("Updating transaction: id=%s", transaction.ID.String())
	query := `
		UPDATE transactions SET type = $2, amount = $3, status = $4,
		reference = $5, updated_at = $6 WHERE id = $1
	`

	transaction.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.Type, transaction.Amount, transaction.Status,
		transaction.Reference, transaction.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to update transaction: " + err.Error())
		return err
	}
	r.logger.Infof("Transaction updated: id=%s", transaction.ID.String())
	return nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TransactionStatus) error {
	r.logger.Debugf("Updating transaction status: id=%s, status=%s", id.String(), status)
	query := `UPDATE transactions SET status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, status, time.Now())
	if err != nil {
		r.logger.Error("Failed to update transaction status: " + err.Error())
		return err
	}
	r.logger.Infof("Transaction status updated: id=%s, status=%s", id.String(), status)
	return nil
}
