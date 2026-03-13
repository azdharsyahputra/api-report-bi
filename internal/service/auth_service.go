package service

import (
	"context"
	"time"

	"portal-report-bi/internal/domain"
	"portal-report-bi/internal/middleware"

	"go.uber.org/zap"
)

type AuthService struct {
	repo   domain.AuthRepository
	logger *zap.Logger
}

func NewAuthService(repo domain.AuthRepository, logger *zap.Logger) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("login attempt failed: user not found", zap.String("email", email))
		return "", domain.ErrInvalidCredentials
	}

	if user.Password != password {
		s.logger.Warn("login attempt failed: invalid password", zap.String("email", email))
		return "", domain.ErrInvalidCredentials
	}

	s.logger.Info("user logged in successfully", zap.String("email", email))
	return middleware.GenerateToken(email)
}
