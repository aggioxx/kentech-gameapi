package service

import (
	"context"
	"kentech-project/internal/core/domain/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"kentech-project/internal/adapters/auth"
	"kentech-project/internal/core/port"
	"kentech-project/pkg/security"
)

type AuthService struct {
	userRepo   port.UserRepository
	jwtService *auth.JWTService
}

func NewAuthService(userRepo port.UserRepository, jwtService *auth.JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *AuthService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	ctx, span := otel.Tracer("").Start(ctx, "AuthService.Register", trace.WithAttributes(
		attribute.String("username", req.Username),
		attribute.String("email", req.Email),
	))
	defer span.End()

	// Check if user already exists
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, model.ErrUserAlreadyExists
	}

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, model.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	ctx, span := otel.Tracer("").Start(ctx, "AuthService.Login", trace.WithAttributes(
		attribute.String("username", req.Username),
	))
	defer span.End()

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, model.ErrInvalidCredentials
	}

	// Check password
	if !security.CheckPasswordHash(req.Password, user.Password) {
		return nil, model.ErrInvalidCredentials
	}

	// Generate token
	token, err := s.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.jwtService.ValidateToken(tokenString)
}
