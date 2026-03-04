package service

import (
	"context"
	"errors"
	"portal-report-bi/internal/domain"
	"time"
)

type BranchBankService struct {
	repo domain.BranchCodeBankRepository
}

func NewBranchBankService(repo domain.BranchCodeBankRepository) *BranchBankService {
	return &BranchBankService{repo: repo}
}

func (s *BranchBankService) GetAll(ctx context.Context, bankName, search string, limit, offset int) ([]domain.BranchCodeBank, int, error) {
	return s.repo.GetAll(ctx, bankName, search, limit, offset)
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
		return nil, err
	}

	return update, nil
}
