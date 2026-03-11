package domain

import "context"

type Report struct {
	ID   int64  `json:"id"`
	Data string `json:"data"`
}

type PayBankReport struct {
	KodeProduk     string `json:"kode_produk"`
	Pengirim       string `json:"pengirim"`
	PrefixPengirim string `json:"prefix_pengirim"`
	KotaPengirim   string `json:"kota_pengirim"`
	NoRek          string `json:"no_rek"`
	PrefixPenerima string `json:"prefix_penerima"`
	NamaPenerima   string `json:"nama_penerima"`
	BankTujuan     string `json:"bank_tujuan"`
	Jumlah         string `json:"jumlah"`
	Volume         int64  `json:"volume"`
	TimeStart      string `json:"time_start"`
}

type MissingBranchReport struct {
	BankTujuan     string `json:"bank_tujuan"`
	PrefixPenerima string `json:"prefix_penerima"`
}

type PayBankReportRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type ReportRepository interface {
	GetReport(ctx context.Context, report []Report) error
	GetPayBankReport(ctx context.Context, startDate, endDate, search, bankTujuan string, limit, offset int) ([]PayBankReport, int, error)
}
