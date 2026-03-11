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

func (s *KycService) GetAllDataKyc(ctx context.Context, startDate, endDate, search string, limit, offset int) ([]domain.Kyc, int, error) {
	return s.repo.GetAllKyc(ctx, startDate, endDate, search, limit, offset)
}
