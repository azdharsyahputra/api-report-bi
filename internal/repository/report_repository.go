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

func (r *reportRepository) executeQuery(query string) ([]byte, error) {
	body := map[string]string{
		"qstr": query,
	}

	b, _ := json.Marshal(body)

	resp, err := http.Post(
		r.queryServiceURL,
		"application/json",
		bytes.NewBuffer(b),
	)
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

func (r *reportRepository) GetReport(ctx context.Context, report []domain.Report) error {
	return nil
}

func (r *reportRepository) GetPayBankReport(
	ctx context.Context,
	startDate, endDate string,
	limit, offset int,
) ([]domain.PayBankReport, int, error) {

	query := fmt.Sprintf(`SELECT
		tt.nom AS kode_produk,
		su.full_name AS pengirim,
		TO_CHAR(r.id) AS prefix_pengirim,
		r.name AS kota_pengirim,
		tt.te_transid
	FROM vdapp_3.t_trans tt
	LEFT JOIN vdapp_3.t_store_user su ON tt.user_name = su.user_name
	LEFT JOIN vdapp_3.regencies r ON su.kode_kota = r.id
	WHERE tt.nom = 'PAYBANK' AND tt.trans_stat = 200
	AND tt.time_start BETWEEN TO_DATE('%s', 'yyyymmdd') AND TO_DATE('%s', 'yyyymmdd') + 1`, startDate, endDate)

	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var rawRows []struct {
		KodeProduk     string `json:"kode_produk"`
		Pengirim       string `json:"pengirim"`
		PrefixPengirim string `json:"prefix_pengirim"`
		KotaPengirim   string `json:"kota_pengirim"`
		TeTransid      string `json:"te_transid"`
	}

	if err := json.Unmarshal(respBody, &rawRows); err != nil {
		return nil, 0, err
	}

	// Go-based regex parsing instead of Oracle REGEXP_SUBSTR
	reNoRek := regexp.MustCompile(`NO\. REK\s*:\s*([^|]+)`)
	reNama := regexp.MustCompile(`NAMA\s*:\s*([^|]+)`)
	reBank := regexp.MustCompile(`BANK\s*:\s*([^|]+)`)
	reJumlah := regexp.MustCompile(`JUMLAH\s*:\s*Rp\.\s*([^|]+)`)

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

	groupedData := make(map[reportKey]int64)

	for _, row := range rawRows {
		transid := row.TeTransid
		var noRek, namaPenerima, bankTujuan, jumlah string

		if match := reNoRek.FindStringSubmatch(transid); len(match) > 1 {
			noRek = strings.TrimSpace(match[1])
		}
		if match := reNama.FindStringSubmatch(transid); len(match) > 1 {
			namaPenerima = strings.TrimSpace(match[1])
		}
		if match := reBank.FindStringSubmatch(transid); len(match) > 1 {
			bankTujuan = strings.TrimSpace(match[1])
		}
		if match := reJumlah.FindStringSubmatch(transid); len(match) > 1 {
			jumlah = strings.TrimSpace(match[1])
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
		groupedData[key]++
	}

	var data []domain.PayBankReport
	for key, vol := range groupedData {
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

	// Manual OrderBy Volume DESC
	sort.Slice(data, func(i, j int) bool {
		return data[i].Volume > data[j].Volume
	})

	total := len(data)

	// Manual Pagination
	if limit > 0 {
		if offset >= total {
			data = []domain.PayBankReport{}
		} else {
			end := offset + limit
			if end > total {
				end = total
			}
			data = data[offset:end]
		}
	}

	return data, total, nil
}
