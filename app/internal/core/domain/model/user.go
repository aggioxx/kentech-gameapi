package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	WalletUserID int       `json:"wallet_user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Password     string    `json:"-"` // Don't include in JSON responses
	Balance      float64   `json:"balance"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
