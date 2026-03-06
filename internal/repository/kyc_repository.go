package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"portal-report-bi/internal/domain"
)

type kycRepository struct {
	queryServiceURL string
}

func NewKycRepository(queryServiceURL string) domain.KycRepository {
	return &kycRepository{
		queryServiceURL: queryServiceURL,
	}
}

func (r *kycRepository) executeQuery(query string) ([]byte, error) {
	body := map[string]string{
		"qstr": query,
	}

	b, _ := json.Marshal(body)

	resp, err := http.Post(r.queryServiceURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query service error: %s", string(respBody))
	}

	return respBody, nil
}

func (r *kycRepository) GetAllKyc(ctx context.Context) ([]domain.Kyc, error) {
	query := `
		SELECT 
			NVL(user_name, ' ') as user_name, 
			NVL(full_name, ' ') as full_name, 
			NVL(upload_type, ' ') as upload_type, 
			NVL(file_name, ' ') as file_name
		FROM kyc
	`

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Data []domain.Kyc `json:"data"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
