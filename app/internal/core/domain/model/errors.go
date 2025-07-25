package model

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrWalletUserIDExhausted = errors.New("no more wallet user IDs available")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrTransactionNotPending = errors.New("transaction is not in pending status")
	ErrUnauthorized          = errors.New("unauthorized")
)
