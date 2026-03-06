package domain

import (
	"context"
)

type Kyc struct {
	UserName   string `json:"user_name"`
	FullName   string `json:"full_name"`
	UploadType string `json:"upload_type"`
	FileName   string `json:"file_name"`
}

type KycRepository interface {
	GetAllKyc(ctx context.Context) ([]Kyc, error)
}
