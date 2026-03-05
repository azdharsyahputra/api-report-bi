package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"portal-report-bi/internal/domain"
)

type regencyRepository struct {
	queryServiceURL string
}

func NewRegencyRepository(queryServiceURL string) domain.RegencyRepository {
	return &regencyRepository{queryServiceURL: queryServiceURL}
}

func (r *regencyRepository) executeQuery(query string) ([]byte, error) {
	body := map[string]string{
		"qstr": query,
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(r.queryServiceURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query service error: %s", string(bodyBytes))
	}
	return io.ReadAll(resp.Body)
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func (r *regencyRepository) Insert(ctx context.Context, regency *domain.Regency) error {
	query := fmt.Sprintf(`
		INSERT INTO regencies (bi_id, regency_name)
		VALUES ('%s', '%s')
	`, escapeString(regency.BIID), escapeString(regency.RegencyName))

	_, err := r.executeQuery(query)
	return err
}

func (r *regencyRepository) Get(ctx context.Context) ([]domain.Regency, error) {
	query := `
		SELECT bi_id, regency_name
		FROM regencies
	`

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, err
	}

	var result []domain.Regency
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}
func (r *regencyRepository) FindByID(ctx context.Context, id int) (*domain.Regency, error) {
	query := fmt.Sprintf(`
		SELECT bi_id, regency_name
		FROM regencies
		WHERE id = %d
	`, id)

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, err
	}

	var result []domain.Regency
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, domain.ErrRegencyNotFound
	}

	return &result[0], nil
}

func (r *regencyRepository) Update(ctx context.Context, regency *domain.Regency) (*domain.Regency, error) {
	query := fmt.Sprintf(`
		UPDATE regencies
		SET regency_name = '%s'
		WHERE id = %d
	`, escapeString(regency.RegencyName), regency.ID)

	_, err := r.executeQuery(query)
	if err != nil {
		return nil, err
	}

	return regency, nil
}

func (r *regencyRepository) Delete(ctx context.Context, id int) error {
	query := fmt.Sprintf(`
		DELETE FROM regencies
		WHERE id = %d
	`, id)

	_, err := r.executeQuery(query)
	return err
}

func (r *regencyRepository) UpdateBIIDByName(ctx context.Context, name string, biid string) error {
	query := fmt.Sprintf(`
		UPDATE regencies
		SET bi_id = '%s'
		WHERE UPPER(TRIM(regency_name)) = UPPER(TRIM('%s'))
	`, escapeString(biid), escapeString(name))

	_, err := r.executeQuery(query)
	return err
}
