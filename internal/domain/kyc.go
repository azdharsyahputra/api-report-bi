package domain

import (
	"context"
)

type Kyc struct {
	UserName      string `json:"user_name"`
	FullName      string `json:"full_name"`
	IsKycApproved int    `json:"is_kyc_approved"`
	Saldo         string `json:"saldo"`
	Nik           string `json:"nik"`
	KodeProv      string `json:"kode_prov"`
	Province      string `json:"province"`
	KodeKota      string `json:"kode_kota"`
	Alamat        string `json:"alamat"`
	KabKota       string `json:"kab_kota"`
	KodeKec       string `json:"kode_kec"`
	Kec           string `json:"kec"`
	KodeKelDes    string `json:"kode_kel_des"`
	KelDes        string `json:"kel_des"`
	KodePos       string `json:"kode_pos"`
	TanggalGabung string `json:"tanggal_gabung"`
	KycFiles      string `json:"kyc_files"`
}

type KycRepository interface {
	GetAllKyc(ctx context.Context, search string, limit, offset int) ([]Kyc, int, error)
}
