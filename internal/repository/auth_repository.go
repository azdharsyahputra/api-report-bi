package repository

import (
	"context"
	"os"
	"portal-report-bi/internal/domain"
)

type authRepository struct{}

func NewAuthRepository() domain.AuthRepository {
	return &authRepository{}
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	validEmail := os.Getenv("ADMIN_EMAIL")
	validPassword := os.Getenv("ADMIN_PASSWORD")

	if email != validEmail {
		return nil, domain.ErrUserNotFound
	}

	return &domain.User{
		Email:    validEmail,
		Password: validPassword,
	}, nil
}
