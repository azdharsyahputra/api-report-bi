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

	resp, err := http.Post(
		r.queryServiceURL,
		"application/json",
		bytes.NewBuffer(b),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query service error: %s", string(bodyBytes))
	}

	return bodyBytes, nil
}


var (
	reNoRek  = regexp.MustCompile(`NO\. REK\s*:\s*([^|]+)`)
	reNama   = regexp.MustCompile(`NAMA\s*:\s*([^|]+)`)
	reBank   = regexp.MustCompile(`BANK\s*:\s*([^|]+)`)
	reJumlah = regexp.MustCompile(`JUMLAH\s*:\s*Rp\.\s*([^|]+)`)
)

func (r *reportRepository) GetPayBankReport(
	ctx context.Context,
	startDate, endDate, search, bankTujuan string,
	limit, offset int,
) ([]domain.PayBankReport, int, error) {

	query := fmt.Sprintf(`
SELECT tt.nom, su.full_name, r.id, r.name, tt.te_transid, TO_CHAR(tt.time_start, 'YYYY-MM-DD HH24:MI:SS') as time_start
FROM t_trans tt
LEFT JOIN t_store_user su ON tt.user_name = su.user_name
LEFT JOIN regencies r ON su.kode_kota = r.id
WHERE tt.nom IN ('PAYBANK','PAYBANKPROMO')
AND tt.trans_stat = 200
AND tt.time_start >= TO_DATE('%s','YYYYMMDD')
AND tt.time_start <  TO_DATE('%s','YYYYMMDD') + 1
	`, startDate, endDate)

	query = normalizeQuery(query)

	respBody, err := r.executeQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var apiResponse struct {
		Data []struct {
			Nom       string `json:"nom"`
			FullName  string `json:"full_name"`
			ID        string `json:"id"`
			Name      string `json:"name"`
			TeTransid string `json:"te_transid"`
			TimeStart string `json:"time_start"`
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
		TimeStart      string
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
			if strings.ToUpper(bankTujuan) == "CIMB" {
				bankTujuan = "CIMB NIAGA"
			}
		}
		if m := reJumlah.FindStringSubmatch(row.TeTransid); len(m) > 1 {
			jumlah = strings.ReplaceAll(strings.TrimSpace(m[1]), ".", "")
		}

		key := reportKey{
			KodeProduk:     row.Nom,
			Pengirim:       row.FullName,
			PrefixPengirim: fmt.Sprintf("%v", row.ID),
			KotaPengirim:   row.Name,
			NoRek:          noRek,
			NamaPenerima:   namaPenerima,
			BankTujuan:     bankTujuan,
			Jumlah:         jumlah,
			TimeStart:      row.TimeStart,
		}

		grouped[key]++
	}

	data := make([]domain.PayBankReport, 0, len(grouped))
	searchTerm := strings.ToLower(search)
	if strings.ToUpper(bankTujuan) == "CIMB" {
		bankTujuan = "CIMB NIAGA"
	}
	bankTujuanFilter := strings.ToLower(bankTujuan)

	for key, vol := range grouped {
		if searchTerm != "" {
			match := strings.Contains(strings.ToLower(key.Pengirim), searchTerm) ||
				strings.Contains(strings.ToLower(key.NamaPenerima), searchTerm) ||
				strings.Contains(strings.ToLower(key.NoRek), searchTerm)
			if !match {
				continue
			}
		}

		if bankTujuanFilter != "" {
			if strings.ToLower(key.BankTujuan) != bankTujuanFilter {
				continue
			}
		}

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
			TimeStart:      key.TimeStart,
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
