package service

import (
	"context"
	"time"

	"portal-report-bi/internal/domain"
	"portal-report-bi/internal/middleware"
)

type AuthService struct {
	repo domain.AuthRepository
}

func NewAuthService(repo domain.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	if user.Password != password {
		return "", domain.ErrInvalidCredentials
	}

	return middleware.GenerateToken(email)
}
