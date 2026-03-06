package service

import (
	"context"
	"portal-report-bi/internal/domain"
)

type KycService struct {
	repo domain.KycRepository
}

func NewKycService(repo domain.KycRepository) *KycService {
	return &KycService{repo: repo}
}

func (s *KycService) GetAllDataKyc(ctx context.Context) ([]domain.Kyc, error) {
	return s.repo.GetAllKyc(ctx)
}
