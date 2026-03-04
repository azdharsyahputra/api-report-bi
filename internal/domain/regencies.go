package domain

import (
	"context"
	"time"
)

type Regency struct {
	ID          int       `json:"id"`
	BIID        string    `json:"bi_id"`
	RegencyName string    `json:"regency_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type RegencyDTO struct {
	ID         string `json:"id"`
	ProvinceID string `json:"province_id"`
	Name       string `json:"name"`
}

type RegencyRepository interface {
	Insert(ctx context.Context, regency *Regency) error
	Get(ctx context.Context) ([]Regency, error)
	FindByID(ctx context.Context, id int) (*Regency, error)
	Update(ctx context.Context, regency *Regency) (*Regency, error)
	UpdateBIIDByName(ctx context.Context, name string, biid string) error
	Delete(ctx context.Context, id int) error
}

type RegencyAPIClient interface {
	GetAll(ctx context.Context) ([]RegencyDTO, error)
}

type CreateRegencyRequest struct {
	BIID        string `json:"bi_id" binding:"required"`
	RegencyName string `json:"regency_name" binding:"required"`
}

type UpdateRegencyRequest struct {
	BIID        string `json:"bi_id" binding:"required"`
	RegencyName string `json:"regency_name" binding:"required"`
}
