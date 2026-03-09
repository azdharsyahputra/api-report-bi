package domain

import (
	"context"
)

type Kyc struct {
	UserName   string `json:"user_name"`
	FullName   string `json:"full_name"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Zip        string `json:"zip"`
	KodeKota   string `json:"kode_kota"`
	KodeProv   string `json:"kode_prov"`
	Email      string `json:"email"`
	UploadType string `json:"upload_type"`
	FileName   string `json:"file_name"`
}

type KycRepository interface {
	GetAllKyc(ctx context.Context, limit, offset int) ([]Kyc, int, error)
}
