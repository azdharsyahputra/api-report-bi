package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"portal-report-bi/internal/domain"
	"regexp"
	"strings"
)

type reportRepository struct {
	queryServiceURL string
}

func NewReportRepository(queryServiceURL string) domain.ReportRepository {
	return &reportRepository{queryServiceURL: queryServiceURL}
}

func normalizeQuery(q string) string {
	q = strings.ReplaceAll(q, "\n", " ")
	q = strings.ReplaceAll(q, "\t", " ")
	q = strings.Join(strings.Fields(q), " ")
	return strings.TrimSpace(q)
}

func (r *reportRepository) executeQuery(query string) ([]byte, error) {

	body := map[string]string{
		"qstr": query,
	}

	b, _ := json.Marshal(body)

	fmt.Println("=================================")
	fmt.Println("QUERY SERVICE URL:", r.queryServiceURL)
	fmt.Println("QUERY SENT:", query)
	fmt.Println("QUERY LENGTH:", len(query))
	fmt.Println("REQUEST BODY:", string(b))
	fmt.Println("=================================")

	resp, err := http.Post(
		r.queryServiceURL,
		"application/json",
		bytes.NewBuffer(b),
	)
	if err != nil {
		fmt.Println("HTTP ERROR:", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Println("RESPONSE STATUS:", resp.StatusCode)
	fmt.Println("RESPONSE BODY:", string(bodyBytes))
	fmt.Println("=================================")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query service error: %s", string(bodyBytes))
	}

	return bodyBytes, nil
}

func (r *reportRepository) GetReport(ctx context.Context, report []domain.Report) error {
	return nil
}

var (
	reNoRek  = regexp.MustCompile(`NO\. REK\s*:\s*([^|]+)`)
	reNama   = regexp.MustCompile(`NAMA\s*:\s*([^|]+)`)
	reBank   = regexp.MustCompile(`BANK\s*:\s*([^|]+)`)
	reJumlah = regexp.MustCompile(`JUMLAH\s*:\s*Rp\.\s*([^|]+)`)
)

func (r *reportRepository) GetPayBankReport(
	ctx context.Context,
	startDate, endDate string,
	limit, offset int,
) ([]domain.PayBankReport, int, error) {

	query := fmt.Sprintf(`
SELECT
    kode_produk,
    no_hp,
    pengirim,
    prefix_pengirim,
    kota_pengirim,
    no_rek,
    nama_penerima,
    bank_tujuan,
    jumlah,
    COUNT(*) AS volume
FROM
(
    SELECT
        tt.nom AS kode_produk,
        tt.user_name AS no_hp,
        su.full_name AS pengirim,
        TO_CHAR(r.id) AS prefix_pengirim,
        r.name AS kota_pengirim,
        REGEXP_SUBSTR(tt.te_transid, 'NO\. REK\s*:\s*([^|]+)', 1, 1, NULL, 1) AS no_rek,
        REGEXP_SUBSTR(tt.te_transid, 'NAMA\s*:\s*([^|]+)', 1, 1, NULL, 1) AS nama_penerima,
        REGEXP_SUBSTR(tt.te_transid, 'BANK\s*:\s*([^|]+)', 1, 1, NULL, 1) AS bank_tujuan,
        REGEXP_SUBSTR(tt.te_transid, 'JUMLAH\s*:\s*Rp\.\s*([^|]+)', 1, 1, NULL, 1) AS jumlah
    FROM vdapp_3.t_trans tt
    LEFT JOIN vdapp_3.t_store_user su ON tt.user_name = su.user_name
    LEFT JOIN vdapp_3.regencies r ON su.kode_kota = r.id
    WHERE tt.nom = 'PAYBANK'
    AND tt.trans_stat = 200
    AND tt.time_start >= TO_DATE('%s','YYYYMMDD')
    AND tt.time_start < TO_DATE('%s','YYYYMMDD')
)
GROUP BY
    kode_produk,
    no_hp,
    pengirim,
    prefix_pengirim,
    kota_pengirim,
    no_rek,
    nama_penerima,
    bank_tujuan,
    jumlah
ORDER BY volume DESC
	`, startDate, endDate)

	query = normalizeQuery(query)

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var apiResponse struct {
		Data []struct {
			KodeProduk     string      `json:"kode_produk"`
			NoHp           string      `json:"no_hp"`
			Pengirim       string      `json:"pengirim"`
			PrefixPengirim string      `json:"prefix_pengirim"`
			KotaPengirim   string      `json:"kota_pengirim"`
			NoRek          string      `json:"no_rek"`
			NamaPenerima   string      `json:"nama_penerima"`
			BankTujuan     string      `json:"bank_tujuan"`
			Jumlah         string      `json:"jumlah"`
			Volume         json.Number `json:"volume"`
		} `json:"data"`
	}

	fmt.Println("DEBUG: RAW JSON PAYBANK = ", string(respBody))

	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		fmt.Println("DEBUG: Unmarshal Error =", err)
		return nil, 0, err
	}

	fmt.Println("DEBUG: Unmarshaled API Data Length =", len(apiResponse.Data))

	var data []domain.PayBankReport

	for _, row := range apiResponse.Data {
		vol, _ := row.Volume.Int64()

		// Sanitize values
		jumlah := strings.ReplaceAll(strings.TrimSpace(row.Jumlah), ".", "")

		data = append(data, domain.PayBankReport{
			KodeProduk:     strings.TrimSpace(row.KodeProduk),
			Pengirim:       strings.TrimSpace(row.Pengirim),
			PrefixPengirim: strings.TrimSpace(row.PrefixPengirim),
			KotaPengirim:   strings.TrimSpace(row.KotaPengirim),
			NoRek:          strings.TrimSpace(row.NoRek),
			NamaPenerima:   strings.TrimSpace(row.NamaPenerima),
			BankTujuan:     strings.TrimSpace(row.BankTujuan),
			Jumlah:         jumlah,
			Volume:         vol,
		})
	}

	total := len(data)

	// In-memory Array Pagination exactly as it was originally used
	if limit > 0 {
		if offset >= total {
			return []domain.PayBankReport{}, total, nil
		}

		end := offset + limit
		if end > total {
			end = total
		}

		data = data[offset:end]
	}

	return data, total, nil
}
