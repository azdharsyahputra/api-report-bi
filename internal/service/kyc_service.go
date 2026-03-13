package service

import (
	"context"
	"portal-report-bi/internal/domain"
	"time"

	"go.uber.org/zap"
)

type KycService struct {
	repo   domain.KycRepository
	logger *zap.Logger
}

func NewKycService(repo domain.KycRepository, logger *zap.Logger) *KycService {
	return &KycService{
		repo:   repo,
		logger: logger,
	}
}

func (s *KycService) GetAllDataKyc(ctx context.Context, startDate, endDate, search string, limit, offset int) ([]domain.Kyc, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	results, total, err := s.repo.GetAllKyc(ctx, startDate, endDate, search, limit, offset)
	if err != nil {
		s.logger.Error("failed to get all kyc data",
			zap.String("startDate", startDate),
			zap.String("endDate", endDate),
			zap.Error(err),
		)
		return nil, 0, err
	}

	return results, total, nil
}
