package domain

import "errors"

var (
	ErrRegencyNotFound    = errors.New("regency not found")
	ErrInvalidId          = errors.New("invalid id")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
