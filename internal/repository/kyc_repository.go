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
			tsu.user_name, 
			tsu.full_name, 
			tsu.address1, 
			tsu.address2, 
			tsu.zip, 
			tsu.kode_kota, 
			tsu.kode_prov, 
			tsu.email, 
			tku.upload_type, 
			tku.file_name 
		FROM t_store_user tsu 
		INNER JOIN t_kyc_upload tku ON tsu.user_name = tku.user_name 
		WHERE tsu.is_kyc_approved = 1
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
