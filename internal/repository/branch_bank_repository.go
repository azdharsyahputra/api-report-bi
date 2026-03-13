package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"portal-report-bi/internal/domain"
	"strings"
	"time"
)

type branchCodeBankRepository struct {
	queryServiceURL string
}

func NewBranchCodeBankRepository(queryServiceURL string) domain.BranchCodeBankRepository {
	return &branchCodeBankRepository{queryServiceURL: queryServiceURL}
}

func (r *branchCodeBankRepository) executeQuery(query string) ([]byte, error) {

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

func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func (r *branchCodeBankRepository) Insert(ctx context.Context, code *domain.BranchCodeBank) error {

	nowStr := time.Now().Format("2006-01-02 15:04:05")

	query := fmt.Sprintf(`
	INSERT INTO m_branch_kode_bank
	(name, branch_code, regencies_code, regencies, office_type, created_at, update_at)
	VALUES('%s','%s','%s','%s','%s',
	TO_DATE('%s','YYYY-MM-DD HH24:MI:SS'),
	TO_DATE('%s','YYYY-MM-DD HH24:MI:SS'))
	`,
		escapeString(code.Name),
		escapeString(code.BranchCode),
		escapeString(code.RegenciesCode),
		escapeString(code.Regencies),
		escapeString(code.OfficeType),
		nowStr,
		nowStr,
	)

	_, err := r.executeQuery(query)
	return err
}

func (r *branchCodeBankRepository) BulkInsert(ctx context.Context, codes []domain.BranchCodeBank) error {

	workerCount := 10
	jobs := make(chan domain.BranchCodeBank)
	errChan := make(chan error, workerCount)

	for w := 0; w < workerCount; w++ {
		go func() {
			for code := range jobs {
				if err := r.Insert(ctx, &code); err != nil {
					errChan <- err
					return
				}
			}
			errChan <- nil
		}()
	}

	for _, code := range codes {
		jobs <- code
	}

	close(jobs)

	for i := 0; i < workerCount; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

func (r *branchCodeBankRepository) GetAll(ctx context.Context, bankName, search string, limit, offset int) ([]domain.BranchCodeBank, int, error) {

	query := `
	SELECT id,
	       NVL(name,' ') as name,
	       NVL(branch_code,' ') as branch_code,
	       NVL(regencies_code,' ') as regencies_code,
	       NVL(regencies,' ') as regencies,
	       NVL(office_type,' ') as office_type,
	       NVL(created_at,SYSDATE) as created_at,
	       NVL(update_at,SYSDATE) as update_at,
	       COUNT(*) OVER() as total_count
	FROM m_branch_kode_bank
	WHERE 1=1
	`

	if bankName != "" {
		query += fmt.Sprintf(" AND UPPER(name)=UPPER('%s')", escapeString(bankName))
	}

	if search != "" {

		pattern := escapeString("%" + search + "%")

		query += fmt.Sprintf(`
		AND (
			UPPER(regencies) LIKE UPPER('%s')
			OR UPPER(office_type) LIKE UPPER('%s')
			OR UPPER(name) LIKE UPPER('%s')
			OR UPPER(branch_code) LIKE UPPER('%s')
			OR UPPER(regencies_code) LIKE UPPER('%s')
		)`, pattern, pattern, pattern, pattern, pattern)
	}

	query += " ORDER BY name ASC, id ASC"

	if limit > 0 {
		query += fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	}

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var apiResp struct {
		Data []struct {
			ID            json.Number `json:"id"`
			Name          string      `json:"name"`
			BranchCode    string      `json:"branch_code"`
			RegenciesCode string      `json:"regencies_code"`
			Regencies     string      `json:"regencies"`
			OfficeType    string      `json:"office_type"`
			CreatedAt     string      `json:"created_at"`
			UpdateAt      string      `json:"update_at"`
			TotalCount    json.Number `json:"total_count"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, 0, err
	}

	var results []domain.BranchCodeBank
	var totalCount int

	for _, b := range apiResp.Data {
		id, _ := b.ID.Int64()
		tc, _ := b.TotalCount.Int64()
		totalCount = int(tc)

		results = append(results, domain.BranchCodeBank{
			ID:            id,
			Name:          b.Name,
			BranchCode:    b.BranchCode,
			RegenciesCode: b.RegenciesCode,
			Regencies:     b.Regencies,
			OfficeType:    b.OfficeType,
			CreatedAt:     b.CreatedAt,
			UpdateAt:      b.UpdateAt,
		})
	}

	return results, totalCount, nil
}

func (r *branchCodeBankRepository) Update(ctx context.Context, code *domain.BranchCodeBank) (*domain.BranchCodeBank, error) {

	nowStr := time.Now().Format("2006-01-02 15:04:05")

	query := fmt.Sprintf(`
	UPDATE m_branch_kode_bank
	SET name='%s', branch_code='%s', regencies_code='%s', regencies='%s', office_type='%s',
	update_at=TO_DATE('%s','YYYY-MM-DD HH24:MI:SS')
	WHERE id=%d
	`,
		escapeString(code.Name),
		escapeString(code.BranchCode),
		escapeString(code.RegenciesCode),
		escapeString(code.Regencies),
		escapeString(code.OfficeType),
		nowStr,
		code.ID,
	)

	_, err := r.executeQuery(query)
	if err != nil {
		return nil, err
	}

	return code, nil
}
