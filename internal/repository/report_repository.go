package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"portal-report-bi/internal/domain"
)

type reportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) domain.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetReport(ctx context.Context, report []domain.Report) error {
	return nil
}

func (r *reportRepository) GetPayBankReport(
	ctx context.Context,
	startDate, endDate string,
	limit, offset int,
) ([]domain.PayBankReport, int, error) {

	query := fmt.Sprintf(`
	SELECT
		kode_produk,
		pengirim,
		prefix_pengirim,
		kota_pengirim,
		no_rek,
		nama_penerima,
		bank_tujuan,
		jumlah,
		volume,
		COUNT(*) OVER() AS total_count
	FROM (
		SELECT
			kode_produk,
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
				su.full_name AS pengirim,
				TO_CHAR(r.id) AS prefix_pengirim,
				r.name AS kota_pengirim,
				REGEXP_SUBSTR(tt.te_transid, 'NO\\. REK\\s*:\\s*([^|]+)', 1, 1, NULL, 1) AS no_rek,
				REGEXP_SUBSTR(tt.te_transid, 'NAMA\\s*:\\s*([^|]+)', 1, 1, NULL, 1) AS nama_penerima,
				REGEXP_SUBSTR(tt.te_transid, 'BANK\\s*:\\s*([^|]+)', 1, 1, NULL, 1) AS bank_tujuan,
				REGEXP_SUBSTR(tt.te_transid, 'JUMLAH\\s*:\\s*Rp\\.\\s*([^|]+)', 1, 1, NULL, 1) AS jumlah
			FROM t_trans tt
			LEFT JOIN t_store_user su
				ON tt.user_name = su.user_name
			LEFT JOIN regencies r
				ON su.kode_kota = r.id
			WHERE tt.nom = 'PAYBANK'
			AND tt.trans_stat = 200
			AND tt.time_start BETWEEN TO_DATE('%s', 'yyyymmdd')
								  AND TO_DATE('%s', 'yyyymmdd') + 1
		)
		GROUP BY
			kode_produk,
			pengirim,
			prefix_pengirim,
			kota_pengirim,
			no_rek,
			nama_penerima,
			bank_tujuan,
			jumlah
	)
	ORDER BY volume DESC
	`, startDate, endDate)

	if limit > 0 {
		query += fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	}

	resp, err := ExecuteQuery(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			domain.PayBankReport
			VolumeRaw json.Number `json:"volume"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	var data []domain.PayBankReport
	for _, r := range result.Data {
		item := r.PayBankReport
		if vol, err := r.VolumeRaw.Int64(); err == nil {
			item.Volume = vol
		}
		data = append(data, item)
	}

	total := len(data)

	return data, total, nil
}
