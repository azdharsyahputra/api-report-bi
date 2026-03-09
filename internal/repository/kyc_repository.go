package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

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

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

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

func (r *kycRepository) GetAllKyc(ctx context.Context, search string, limit, offset int) ([]domain.Kyc, int, error) {

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
		return nil, 0, err
	}

	rawJSON := string(respBody)

	// FIX 1: hapus newline dari query service
	rawJSON = strings.ReplaceAll(rawJSON, "\r", "")
	rawJSON = strings.ReplaceAll(rawJSON, "\n", "")
	rawJSON = strings.ReplaceAll(rawJSON, "\t", "")

	// FIX 2: escape invalid backslash (\6 \s dll)
	rawJSON = fixInvalidEscape(rawJSON)

	// FIX 3: sanitize windows path
	rawJSON = sanitizeWindowsPath(rawJSON)

	var apiResp struct {
		Data []struct {
			UserName   string `json:"user_name"`
			FullName   string `json:"full_name"`
			Address1   string `json:"address1"`
			Address2   string `json:"address2"`
			Zip        string `json:"zip"`
			KodeKota   string `json:"kode_kota"`
			KodeProv   string `json:"kode_prov"`
			Email      string `json:"email"`
			UploadType string `json:"upload_type"`
			FileName   string `json:"file_name"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(rawJSON), &apiResp); err != nil {
		return nil, 0, fmt.Errorf("json decode failed: %v, raw: %s", err, rawJSON)
	}

	// Group by user_name + full_name + address1 + email
	type groupKey struct {
		UserName string
		FullName string
		Address1 string
		Email    string
	}

	grouped := make(map[groupKey]*domain.Kyc)
	order := make([]groupKey, 0)

	for _, row := range apiResp.Data {
		key := groupKey{
			UserName: row.UserName,
			FullName: row.FullName,
			Address1: row.Address1,
			Email:    row.Email,
		}

		if existing, ok := grouped[key]; ok {
			existing.UploadType = append(existing.UploadType, row.UploadType)
			existing.FileName = append(existing.FileName, row.FileName)
		} else {
			grouped[key] = &domain.Kyc{
				UserName:   row.UserName,
				FullName:   row.FullName,
				Address1:   row.Address1,
				Address2:   row.Address2,
				Zip:        row.Zip,
				KodeKota:   row.KodeKota,
				KodeProv:   row.KodeProv,
				Email:      row.Email,
				UploadType: []string{row.UploadType},
				FileName:   []string{row.FileName},
			}
			order = append(order, key)
		}
	}

	// Build ordered result
	data := make([]domain.Kyc, 0, len(order))
	for _, key := range order {
		data = append(data, *grouped[key])
	}

	// Filter by search keyword (case-insensitive)
	if search != "" {
		s := strings.ToLower(search)
		filtered := make([]domain.Kyc, 0)
		for _, k := range data {
			if strings.Contains(strings.ToLower(k.UserName), s) ||
				strings.Contains(strings.ToLower(k.FullName), s) ||
				strings.Contains(strings.ToLower(k.Email), s) {
				filtered = append(filtered, k)
			}
		}
		data = filtered
	}

	total := len(data)

	if limit > 0 {
		if offset >= total {
			return []domain.Kyc{}, total, nil
		}

		end := offset + limit
		if end > total {
			end = total
		}

		return data[offset:end], total, nil
	}

	return data, total, nil
}

func sanitizeWindowsPath(s string) string {

	re := regexp.MustCompile(`c:\\+[^"]+`)

	return re.ReplaceAllStringFunc(s, func(path string) string {

		path = strings.ReplaceAll(path, "\\", "/")

		for strings.Contains(path, "//") {
			path = strings.ReplaceAll(path, "//", "/")
		}

		return path
	})
}

func fixInvalidEscape(s string) string {

	var b strings.Builder

	for i := 0; i < len(s); i++ {

		if s[i] == '\\' && i+1 < len(s) {

			c := s[i+1]

			if c != '\\' &&
				c != '"' &&
				c != '/' &&
				c != 'b' &&
				c != 'f' &&
				c != 'n' &&
				c != 'r' &&
				c != 't' &&
				c != 'u' {

				b.WriteByte('\\')
			}
		}

		b.WriteByte(s[i])
	}

	return b.String()
}
