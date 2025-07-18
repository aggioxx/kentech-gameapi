package service

import (
	"context"
	"kentech-project/internal/core/domain/model"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"kentech-project/internal/core/port"
)

type PlayerService struct {
	userRepo port.UserRepository
	txRepo   port.TransactionRepository
}

func NewPlayerService(userRepo port.UserRepository, txRepo port.TransactionRepository) *PlayerService {
	return &PlayerService{
		userRepo: userRepo,
		txRepo:   txRepo,
	}
}

func (s *PlayerService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	ctx, span := otel.Tracer("").Start(ctx, "PlayerService.GetProfile", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
	))
	defer span.End()
	return s.userRepo.GetByID(ctx, userID)
}

func (s *PlayerService) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]*model.Transaction, error) {
	ctx, span := otel.Tracer("").Start(ctx, "PlayerService.GetTransactionHistory", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
	))
	defer span.End()
	return s.txRepo.GetByUserID(ctx, userID)
}

func (s *PlayerService) GetBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "PlayerService.GetBalance", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
	))
	defer span.End()
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.Balance, nil
}
