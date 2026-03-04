package external

import (
	"context"
	"portal-report-bi/internal/domain"
)

type regencyAPIClient struct {
	baseURL string
}

func NewRegencyAPIClient(baseURL string) domain.RegencyAPIClient {
	return &regencyAPIClient{
		baseURL: baseURL,
	}
}

func (c *regencyAPIClient) GetAll(ctx context.Context) ([]domain.RegencyDTO, error) {

	return []domain.RegencyDTO{
		{ID: "3201", ProvinceID: "32", Name: "KABUPATEN BOGOR"},
		{ID: "3202", ProvinceID: "32", Name: "KABUPATEN SUKABUMI"},
		{ID: "3203", ProvinceID: "32", Name: "KABUPATEN CIANJUR"},
		{ID: "3204", ProvinceID: "32", Name: "KABUPATEN BANDUNG"},
		{ID: "3205", ProvinceID: "32", Name: "KABUPATEN GARUT"},
		{ID: "3206", ProvinceID: "32", Name: "KABUPATEN TASIKMALAYA"},
		{ID: "3207", ProvinceID: "32", Name: "KABUPATEN CIAMIS"},
		{ID: "3208", ProvinceID: "32", Name: "KABUPATEN KUNINGAN"},
		{ID: "3209", ProvinceID: "32", Name: "KABUPATEN CIREBON"},
		{ID: "3210", ProvinceID: "32", Name: "KABUPATEN MAJALENGKA"},
		{ID: "3211", ProvinceID: "32", Name: "KABUPATEN SUMEDANG"},
		{ID: "3212", ProvinceID: "32", Name: "KABUPATEN INDRAMAYU"},
		{ID: "3213", ProvinceID: "32", Name: "KABUPATEN SUBANG"},
		{ID: "3214", ProvinceID: "32", Name: "KABUPATEN PURWAKARTA"},
		{ID: "3215", ProvinceID: "32", Name: "KABUPATEN KARAWANG"},
		{ID: "3216", ProvinceID: "32", Name: "KABUPATEN BEKASI"},
		{ID: "3217", ProvinceID: "32", Name: "KABUPATEN BANDUNG BARAT"},
		{ID: "3218", ProvinceID: "32", Name: "KABUPATEN PANGANDARAN"},
		{ID: "3271", ProvinceID: "32", Name: "KOTA BOGOR"},
		{ID: "3272", ProvinceID: "32", Name: "KOTA SUKABUMI"},
		{ID: "3273", ProvinceID: "32", Name: "KOTA BANDUNG"},
		{ID: "3274", ProvinceID: "32", Name: "KOTA CIREBON"},
		{ID: "3275", ProvinceID: "32", Name: "KOTA BEKASI"},
		{ID: "3276", ProvinceID: "32", Name: "KOTA DEPOK"},
		{ID: "3277", ProvinceID: "32", Name: "KOTA CIMAHI"},
		{ID: "3278", ProvinceID: "32", Name: "KOTA TASIKMALAYA"},
		{ID: "3279", ProvinceID: "32", Name: "KOTA BANJAR"},
	}, nil
}
