package service

import (
	"context"
	"kentech-project/internal/adapters/auth"
	"kentech-project/internal/core/domain/model"
	"kentech-project/internal/core/port"
	"kentech-project/pkg/logger"
	"kentech-project/pkg/security"
	"sync"
)

var walletIDMutex sync.Mutex
var walletIDIndex int

// as the provided wallet does not support creating new users,
// we will use existing user IDs for testing purposes
var walletUserIDs = []int{
	34633089486, // USD
	34679664254, // EUR
	34616761765, // KES
	34673635133, // USD
}

type AuthService struct {
	userRepo   port.UserRepository
	jwtService *auth.JWTService
	logger     *logger.Logger
}

func NewAuthService(userRepo port.UserRepository, jwtService *auth.JWTService, log *logger.Logger) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     log,
	}
}

func (s *AuthService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	s.logger.Debugf("Register called: username=%s, email=%s", req.Username, req.Email)

	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		s.logger.Warnf("Register failed: username already exists: %s", req.Username)
		return nil, model.ErrUserAlreadyExists
	}

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		s.logger.Warnf("Register failed: email already exists: %s", req.Email)
		return nil, model.ErrUserAlreadyExists
	}

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Register failed: password hashing error: " + err.Error())
		return nil, err
	}

	walletIDMutex.Lock()
	assignedWalletID := walletUserIDs[walletIDIndex%len(walletUserIDs)]
	walletIDIndex++
	walletIDMutex.Unlock()

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		Password:     hashedPassword,
		WalletUserID: assignedWalletID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Register failed: user creation error: " + err.Error())
		return nil, err
	}

	s.logger.Infof("Register successful: user_id=%s, username=%s", user.ID.String(), user.Username)
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	s.logger.Debugf("Login called: username=%s", req.Username)

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Warnf("Login failed: invalid credentials for username=%s", req.Username)
		return nil, model.ErrInvalidCredentials
	}

	if !security.CheckPasswordHash(req.Password, user.Password) {
		s.logger.Warnf("Login failed: password mismatch for username=%s", req.Username)
		return nil, model.ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		s.logger.Error("Login failed: token generation error: " + err.Error())
		return nil, err
	}

	s.logger.Infof("Login successful: user_id=%s, username=%s", user.ID.String(), user.Username)
	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	s.logger.Debug("ValidateToken called")
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		s.logger.Warn("ValidateToken failed: " + err.Error())
		return nil, err
	}
	s.logger.Info("ValidateToken successful")
	return claims, nil
}
