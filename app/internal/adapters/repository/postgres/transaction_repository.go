package postgres

import (
	"context"
	"database/sql"
	"kentech-project/internal/core/domain/model"
	"time"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
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

	return err
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, status, reference, created_at, updated_at
		FROM transactions WHERE id = $1
	`

	transaction := &model.Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Type, &transaction.Amount,
		&transaction.Status, &transaction.Reference, &transaction.CreatedAt, &transaction.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, model.ErrTransactionNotFound
	}

	return transaction, err
}

func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, status, reference, created_at, updated_at
		FROM transactions WHERE user_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
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
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, rows.Err()
}

func (r *TransactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	query := `
		UPDATE transactions SET type = $2, amount = $3, status = $4, 
		reference = $5, updated_at = $6 WHERE id = $1
	`

	transaction.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.Type, transaction.Amount, transaction.Status,
		transaction.Reference, transaction.UpdatedAt)

	return err
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TransactionStatus) error {
	query := `UPDATE transactions SET status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, status, time.Now())
	return err
}
