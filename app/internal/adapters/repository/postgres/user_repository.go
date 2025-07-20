package postgres

import (
	"context"
	"database/sql"
	"kentech-project/internal/core/domain/model"
	"kentech-project/pkg/logger"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewUserRepository(db *sql.DB, log *logger.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: log,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	r.logger.Debug("Creating new user")

	query := `
		INSERT INTO users (id, wallet_user_id, username, email, password, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Balance = 0.0

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.WalletUserID, user.Username, user.Email, user.Password,
		user.Balance, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create user: " + err.Error())
		return err
	}
	r.logger.Infof("User created: id=%s, username=%s, email=%s", user.ID.String(), user.Username, user.Email)
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	r.logger.Debugf("Fetching user by ID: %s", id.String())

	query := `
		SELECT id, wallet_user_id, username, email, password, balance, created_at, updated_at
		FROM users WHERE id = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.WalletUserID, &user.Username, &user.Email, &user.Password,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		r.logger.Warnf("User not found: id=%s", id.String())
		return nil, model.ErrUserNotFound
	}
	if err != nil {
		r.logger.Error("Failed to fetch user: " + err.Error())
		return nil, err
	}
	r.logger.Infof("User fetched: id=%s, username=%s", user.ID.String(), user.Username)
	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	r.logger.Debugf("Fetching user by username: %s", username)

	query := `
		SELECT id, wallet_user_id, username, email, password, balance, created_at, updated_at
		FROM users WHERE username = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.WalletUserID, &user.Username, &user.Email, &user.Password,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		r.logger.Warnf("User not found: username=%s", username)
		return nil, model.ErrUserNotFound
	}
	if err != nil {
		r.logger.Error("Failed to fetch user by username: " + err.Error())
		return nil, err
	}
	r.logger.Infof("User fetched by username: id=%s, username=%s", user.ID.String(), user.Username)
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	r.logger.Debugf("Fetching user by email: %s", email)

	query := `
		SELECT id, wallet_user_id, username, email, password, balance, created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.WalletUserID, &user.Username, &user.Email, &user.Password,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		r.logger.Warnf("User not found: email=%s", email)
		return nil, model.ErrUserNotFound
	}
	if err != nil {
		r.logger.Error("Failed to fetch user by email: " + err.Error())
		return nil, err
	}
	r.logger.Infof("User fetched by email: id=%s, email=%s", user.ID.String(), user.Email)
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	r.logger.Debugf("Updating user: id=%s", user.ID.String())
	query := `
		UPDATE users SET wallet_user_id = $2, username = $3, email = $4, password = $5,
		balance = $6, updated_at = $7 WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.WalletUserID, user.Username, user.Email, user.Password,
		user.Balance, user.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to update user: " + err.Error())
		return err
	}
	r.logger.Infof("User updated: id=%s, username=%s", user.ID.String(), user.Username)
	return nil
}

func (r *UserRepository) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	r.logger.Debugf("Updating user balance: id=%s, balance=%f", userID.String(), balance)
	query := `UPDATE users SET balance = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID, balance, time.Now())
	if err != nil {
		r.logger.Error("Failed to update user balance: " + err.Error())
		return err
	}
	r.logger.Infof("User balance updated: id=%s, balance=%f", userID.String(), balance)
	return nil
}
