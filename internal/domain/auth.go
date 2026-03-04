package domain

import "context"

type User struct {
	Email    string
	Password string
}

type AuthRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
