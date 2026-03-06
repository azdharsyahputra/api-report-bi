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
	"sort"
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
		SELECT tt.nom AS kode_produk,
			su.full_name AS pengirim,
			TO_CHAR(r.id) AS prefix_pengirim,
			r.name AS kota_pengirim,
			tt.te_transid
		FROM vdapp_3.t_trans tt
		LEFT JOIN vdapp_3.t_store_user su ON tt.user_name = su.user_name
		LEFT JOIN vdapp_3.regencies r ON su.kode_kota = r.id
		WHERE tt.nom = 'PAYBANK'
		AND tt.trans_stat = 200
		AND tt.time_start >= TO_DATE('%s','YYYYMMDD')
		AND tt.time_start < TO_DATE('%s','YYYYMMDD')
		`, startDate, endDate)

	query = normalizeQuery(query)

	fmt.Println("FINAL SQL:", query)

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var apiResponse struct {
		Data []struct {
			KodeProduk     string `json:"kode_produk"`
			Pengirim       string `json:"pengirim"`
			PrefixPengirim string `json:"prefix_pengirim"`
			KotaPengirim   string `json:"kota_pengirim"`
			TeTransid      string `json:"te_transid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		return nil, 0, err
	}

	type reportKey struct {
		KodeProduk     string
		Pengirim       string
		PrefixPengirim string
		KotaPengirim   string
		NoRek          string
		NamaPenerima   string
		BankTujuan     string
		Jumlah         string
	}

	grouped := make(map[reportKey]int64)

	for _, row := range apiResponse.Data {

		var noRek, namaPenerima, bankTujuan, jumlah string

		if m := reNoRek.FindStringSubmatch(row.TeTransid); len(m) > 1 {
			noRek = strings.TrimSpace(m[1])
		}
		if m := reNama.FindStringSubmatch(row.TeTransid); len(m) > 1 {
			namaPenerima = strings.TrimSpace(m[1])
		}
		if m := reBank.FindStringSubmatch(row.TeTransid); len(m) > 1 {
			bankTujuan = strings.TrimSpace(m[1])
		}
		if m := reJumlah.FindStringSubmatch(row.TeTransid); len(m) > 1 {
			jumlah = strings.ReplaceAll(strings.TrimSpace(m[1]), ".", "")
		}

		key := reportKey{
			KodeProduk:     row.KodeProduk,
			Pengirim:       row.Pengirim,
			PrefixPengirim: row.PrefixPengirim,
			KotaPengirim:   row.KotaPengirim,
			NoRek:          noRek,
			NamaPenerima:   namaPenerima,
			BankTujuan:     bankTujuan,
			Jumlah:         jumlah,
		}

		grouped[key]++
	}

	data := make([]domain.PayBankReport, 0, len(grouped))

	for key, vol := range grouped {
		data = append(data, domain.PayBankReport{
			KodeProduk:     key.KodeProduk,
			Pengirim:       key.Pengirim,
			PrefixPengirim: key.PrefixPengirim,
			KotaPengirim:   key.KotaPengirim,
			NoRek:          key.NoRek,
			NamaPenerima:   key.NamaPenerima,
			BankTujuan:     key.BankTujuan,
			Jumlah:         key.Jumlah,
			Volume:         vol,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Volume > data[j].Volume
	})

	total := len(data)

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
