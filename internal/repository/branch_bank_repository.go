package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"portal-report-bi/internal/domain"
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
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query service error: %s", string(bodyBytes))
	}
	return io.ReadAll(resp.Body)
}

func (r *branchCodeBankRepository) Insert(ctx context.Context, code *domain.BranchCodeBank) error {
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	query := fmt.Sprintf(`
		INSERT INTO vdapp_3.m_branch_kode_bank(name, branch_code, regencies_code, regencies, office_type, created_at, update_at)
		VALUES('%s', '%s', '%s', '%s', '%s', TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS'), TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS'))
	`, escapeString(code.Name), escapeString(code.BranchCode), escapeString(code.RegenciesCode), escapeString(code.Regencies), escapeString(code.OfficeType), nowStr, nowStr)

	_, err := r.executeQuery(query)
	return err
}

func (r *branchCodeBankRepository) GetAll(ctx context.Context, bankName, search string, limit, offset int) ([]domain.BranchCodeBank, int, error) {
	query := `
		SELECT id, 
		       NVL(name, ' ') as name, 
		       NVL(branch_code, ' ') as branch_code, 
		       NVL(regencies_code, ' ') as regencies_code, 
		       NVL(regencies, ' ') as regencies, 
		       NVL(office_type, ' ') as office_type, 
		       NVL(created_at, SYSDATE) as created_at, 
		       NVL(update_at, SYSDATE) as update_at,
		       COUNT(*) OVER() as total_count
		FROM vdapp_3.m_branch_kode_bank
		WHERE 1=1
	`

	if bankName != "" {
		query += fmt.Sprintf(" AND UPPER(name) = UPPER('%s')", escapeString(bankName))
	}

	if search != "" {
		pattern := escapeString("%" + search + "%")
		query += fmt.Sprintf(` 
			AND (
				UPPER(regencies) LIKE UPPER('%s') OR 
				UPPER(office_type) LIKE UPPER('%s') OR
				UPPER(name) LIKE UPPER('%s')
			)`, pattern, pattern, pattern)
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
			domain.BranchCodeBank
			TotalCount int `json:"total_count"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, 0, err
	}

	var results []domain.BranchCodeBank
	var totalCount int
	for _, b := range apiResp.Data {
		results = append(results, b.BranchCodeBank)
		totalCount = b.TotalCount
	}

	return results, totalCount, nil
}

func (r *branchCodeBankRepository) FindByID(ctx context.Context, id int) (*domain.BranchCodeBank, error) {
	query := fmt.Sprintf(`
		SELECT id, name, branch_code, regencies_code, regencies, office_type, created_at, update_at
		FROM vdapp_3.m_branch_kode_bank
		WHERE id = %d
	`, id)

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Data []domain.BranchCodeBank `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, nil
	}
	return &apiResp.Data[0], nil
}

func (r *branchCodeBankRepository) Update(ctx context.Context, code *domain.BranchCodeBank) (*domain.BranchCodeBank, error) {
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	query := fmt.Sprintf(`
		UPDATE vdapp_3.m_branch_kode_bank
		SET name = '%s', branch_code = '%s', regencies_code = '%s', regencies = '%s', office_type = '%s', update_at = TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS')
		WHERE id = %d
	`, escapeString(code.Name), escapeString(code.BranchCode), escapeString(code.RegenciesCode), escapeString(code.Regencies), escapeString(code.OfficeType), nowStr, code.ID)

	_, err := r.executeQuery(query)
	if err != nil {
		return nil, err
	}

	return code, nil
}

func (r *branchCodeBankRepository) Delete(ctx context.Context, id int) error {
	query := fmt.Sprintf(`DELETE FROM vdapp_3.m_branch_kode_bank WHERE id = %d`, id)
	_, err := r.executeQuery(query)
	return err
}

func (r *branchCodeBankRepository) BulkInsert(ctx context.Context, codes []domain.BranchCodeBank) error {
	if len(codes) == 0 {
		return nil
	}

	batchSize := 500
	for i := 0; i < len(codes); i += batchSize {
		end := i + batchSize
		if end > len(codes) {
			end = len(codes)
		}
		batch := codes[i:end]

		if err := r.executeBatchInsert(batch); err != nil {
			return err
		}
	}

	return nil
}

func (r *branchCodeBankRepository) executeBatchInsert(batch []domain.BranchCodeBank) error {
	for _, code := range batch {
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		query := fmt.Sprintf(`
			MERGE INTO vdapp_3.m_branch_kode_bank t
			USING (SELECT '%s' as name, '%s' as branch_code FROM dual) s
			ON (t.name = s.name AND t.branch_code = s.branch_code)
			WHEN MATCHED THEN
				UPDATE SET t.regencies_code = '%s', t.regencies = '%s', t.office_type = '%s', t.update_at = TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS')
			WHEN NOT MATCHED THEN
				INSERT (name, branch_code, regencies_code, regencies, office_type, created_at, update_at)
				VALUES ('%s', '%s', '%s', '%s', '%s', TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS'), TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS'))
		`,
			escapeString(code.Name), escapeString(code.BranchCode),
			escapeString(code.RegenciesCode), escapeString(code.Regencies), escapeString(code.OfficeType), nowStr,
			escapeString(code.Name), escapeString(code.BranchCode), escapeString(code.RegenciesCode), escapeString(code.Regencies), escapeString(code.OfficeType), nowStr, nowStr,
		)

		_, err := r.executeQuery(query)
		if err != nil {
			continue
		}
	}
	return nil
}
