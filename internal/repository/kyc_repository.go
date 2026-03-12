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

func (r *kycRepository) GetAllKyc(ctx context.Context, startDate, endDate, search string, limit, offset int) ([]domain.Kyc, int, error) {

	query := fmt.Sprintf(`
		select 
			user_name,
			full_name,
			is_kyc_approved,
			nvl(balance,0) saldo,
			no_id nik,
			kode_prov,
			b.name province,
			kode_kota,
			trim(address1||' '||address2) alamat,
			c.name kab_kota,
			kode_kec,
			d.name kec,
			kode_kel_des,
			e.name kel_des,
			zip kode_pos,
			to_char(joining_date,'dd/mm/yyyy') tanggal_gabung,
			replace(replace(get_all_store_user_ekyc(store_id,user_name),'c:\superx\img\','https://api.cashplus.id:3601/get_img?file='),'\','/') kyc_files
		from t_store_user a 
		left join provinces b on a.kode_prov=b.id 
		left join regencies c on a.kode_kota=c.id 
		left join districts d on a.kode_kec=d.id 
		left join villages e on a.kode_kel_des=e.id 
		where nvl(is_kyc_approved,0)=1 
		and nvl(lvl,0)>1 
		and joining_date between to_date('%s','yyyymmdd') and to_date('%s','yyyymmdd')+1
	`, startDate, endDate)

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
		Data []domain.Kyc `json:"data"`
	}

	if err := json.Unmarshal([]byte(rawJSON), &apiResp); err != nil {
		return nil, 0, fmt.Errorf("json decode failed: %v, raw: %s", err, rawJSON)
	}

	data := apiResp.Data

	// Clean up kyc_files: trim spaces and fix backslashes
	for i := range data {
		if data[i].KycFiles != "" {
			files := strings.Split(data[i].KycFiles, ",")
			for j := range files {
				files[j] = strings.TrimSpace(files[j])
				// Ensure forward slash for URL consistency
				files[j] = strings.ReplaceAll(files[j], "\\", "/")
			}
			data[i].KycFiles = strings.Join(files, ",")
		}
	}

	// Filter by search keyword (case-insensitive)
	if search != "" {
		s := strings.ToLower(search)
		filtered := make([]domain.Kyc, 0)
		for _, k := range data {
			if strings.Contains(strings.ToLower(k.UserName), s) ||
				strings.Contains(strings.ToLower(k.FullName), s) {
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
