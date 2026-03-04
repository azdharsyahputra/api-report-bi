package domain

import "context"

type ExportRow struct {
	PrefixPengirim   string
	PrefixPenerima   string
	NamaPenerima     string
	NamaPengirim     string
	VolumeTransaksi  string
	NominalTransaksi string
	TujuanTransaksi  string
}

type ExportBankService interface {
	Export(ctx context.Context, startDate, endDate string) ([]ExportRow, error)
}
