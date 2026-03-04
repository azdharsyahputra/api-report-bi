package repository

import (
	"context"
	"portal-report-bi/internal/domain"
)

type authRepository struct{}

func NewAuthRepository() domain.AuthRepository {
	return &authRepository{}
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const (
		validEmail    = "admin@portal.com"
		validPassword = "rahasia_report_bi"
	)

	if email != validEmail {
		return nil, domain.ErrUserNotFound
	}

	return &domain.User{
		Email:    validEmail,
		Password: validPassword,
	}, nil
}
