package postgres

import (
	"context"
	"database/sql"
	"kentech-project/internal/core/domain/model"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, username, email, password, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Balance = 0.0

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.Password,
		user.Balance, user.CreatedAt, user.UpdatedAt)

	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, email, password, balance, created_at, updated_at
		FROM users WHERE id = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, model.ErrUserNotFound
	}

	return user, err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, balance, created_at, updated_at
		FROM users WHERE username = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, model.ErrUserNotFound
	}

	return user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, balance, created_at, updated_at
		FROM users WHERE email = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, model.ErrUserNotFound
	}

	return user, err
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET username = $2, email = $3, password = $4, 
		balance = $5, updated_at = $6 WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.Password,
		user.Balance, user.UpdatedAt)

	return err
}

func (r *UserRepository) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	query := `UPDATE users SET balance = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID, balance, time.Now())
	return err
}
