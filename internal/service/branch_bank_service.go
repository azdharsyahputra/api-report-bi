package service

import (
	"context"
	"errors"
	"portal-report-bi/internal/domain"
	"time"

	"go.uber.org/zap"
)

type BranchBankService struct {
	repo   domain.BranchCodeBankRepository
	logger *zap.Logger
}

func NewBranchBankService(repo domain.BranchCodeBankRepository, logger *zap.Logger) *BranchBankService {
	return &BranchBankService{
		repo:   repo,
		logger: logger,
	}
}

func (s *BranchBankService) GetAll(ctx context.Context, bankName, search string, limit, offset int) ([]domain.BranchCodeBank, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	results, total, err := s.repo.GetAll(ctx, bankName, search, limit, offset)
	if err != nil {
		s.logger.Error("failed to get all branch bank",
			zap.String("bankName", bankName),
			zap.String("search", search),
			zap.Error(err),
		)
		return nil, 0, err
	}

	return results, total, nil
}

func (s *BranchBankService) Update(ctx context.Context, code *domain.BranchCodeBank) (*domain.BranchCodeBank, error) {
	if code.ID == 0 {
		return nil, domain.ErrInvalidId
	}

	if code.Name == "" || code.BranchCode == "" || code.RegenciesCode == "" || code.Regencies == "" || code.OfficeType == "" {
		return nil, errors.New("all field must be filled")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	update, err := s.repo.Update(ctx, code)
	if err != nil {
		s.logger.Error("failed to update branch bank",
			zap.Int64("id", code.ID),
			zap.Error(err),
		)
		return nil, err
	}

	s.logger.Info("successfully updated branch bank",
		zap.Int64("id", code.ID),
		zap.String("name", code.Name),
	)

	return update, nil
}
