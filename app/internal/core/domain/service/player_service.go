package service

import (
	"context"
	"kentech-project/internal/core/domain/model"
	"kentech-project/pkg/logger"

	"github.com/google/uuid"
	"kentech-project/internal/core/port"
)

type PlayerService struct {
	userRepo port.UserRepository
	txRepo   port.TransactionRepository
	logger   *logger.Logger
}

func NewPlayerService(userRepo port.UserRepository, txRepo port.TransactionRepository, log *logger.Logger) *PlayerService {
	return &PlayerService{
		userRepo: userRepo,
		txRepo:   txRepo,
		logger:   log,
	}
}

func (s *PlayerService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	s.logger.Debugf("GetProfile called: user_id=%s", userID.String())

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Warnf("GetProfile failed for user_id=%s: %s", userID.String(), err.Error())
		return nil, err
	}
	s.logger.Infof("GetProfile successful: user_id=%s", userID.String())
	return user, nil
}

func (s *PlayerService) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]*model.Transaction, error) {
	s.logger.Debugf("GetTransactionHistory called: user_id=%s", userID.String())

	transactions, err := s.txRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Warnf("GetTransactionHistory failed for user_id=%s: %s", userID.String(), err.Error())
		return nil, err
	}
	s.logger.Infof("GetTransactionHistory successful: user_id=%s, count=%d", userID.String(), len(transactions))
	return transactions, nil
}

func (s *PlayerService) GetBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	s.logger.Debugf("GetBalance called: user_id=%s", userID.String())

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Warnf("GetBalance failed for user_id=%s: %s", userID.String(), err.Error())
		return 0, err
	}
	s.logger.Infof("GetBalance successful: user_id=%s, balance=%f", userID.String(), user.Balance)
	return user.Balance, nil
}
