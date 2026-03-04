package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"portal-report-bi/internal/domain"
)

type RegencyService struct {
	repo      domain.RegencyRepository
	apiClient domain.RegencyAPIClient
}

func NewRegencyService(repo domain.RegencyRepository, apiClient domain.RegencyAPIClient) *RegencyService {
	return &RegencyService{repo: repo, apiClient: apiClient}
}

func (s *RegencyService) Create(ctx context.Context, BIID string, regencyName string) error {

	regency := &domain.Regency{
		BIID:        BIID,
		RegencyName: regencyName,
	}

	return s.repo.Insert(ctx, regency)
}

func (s *RegencyService) Get(ctx context.Context) ([]domain.Regency, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	regencies, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	return regencies, nil
}

func (s *RegencyService) FindByID(ctx context.Context, id int) (*domain.Regency, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidId
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	regency, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return regency, nil
}

func (s *RegencyService) Delete(ctx context.Context, id int) error {
	if id == 0 {
		return domain.ErrRegencyNotFound
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.repo.Delete(ctx, id)
}

func (s *RegencyService) Update(ctx context.Context, regency *domain.Regency) (*domain.Regency, error) {
	if regency == nil || regency.BIID == "" {
		return nil, domain.ErrRegencyNotFound
	}

	if regency.RegencyName == "" {
		return nil, errors.New("Regency name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update, err := s.repo.Update(ctx, regency)
	if err != nil {
		return nil, err
	}

	return update, nil
}
func (s *RegencyService) SyncAndGetAll(
	ctx context.Context,
) ([]domain.Regency, error) {

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	apiData, err := s.apiClient.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	dbData, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	dbMap := make(map[string]domain.Regency)
	for _, r := range dbData {
		key := strings.ToUpper(strings.TrimSpace(r.RegencyName))
		dbMap[key] = r
	}

	var result []domain.Regency
	seenNames := make(map[string]bool)

	for _, r := range dbData {
		key := strings.ToUpper(strings.TrimSpace(r.RegencyName))
		result = append(result, r)
		seenNames[key] = true
	}

	for _, a := range apiData {
		key := strings.ToUpper(strings.TrimSpace(a.Name))
		if !seenNames[key] {
			regency := domain.Regency{
				BIID:        a.ID,
				RegencyName: a.Name,
				CreatedAt:   time.Now(),
			}
			result = append(result, regency)
		}
	}

	return result, nil
}
