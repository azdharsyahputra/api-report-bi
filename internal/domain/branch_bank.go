package domain

import (
	"context"
	"time"
)

type BranchCodeBank struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	BranchCode    string    `json:"branch_code"`
	RegenciesCode string    `json:"regencies_code"`
	Regencies     string    `json:"regencies"`
	OfficeType    string    `json:"office_type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdateAt      time.Time `json:"update_at"`
}

type UpdateBranchCodeBankRequest struct {
	Name          string `json:"name" binding:"required"`
	BranchCode    string `json:"branch_code" binding:"required"`
	RegenciesCode string `json:"regencies_code" binding:"required"`
	Regencies     string `json:"regencies" binding:"required"`
	OfficeType    string `json:"office_type" binding:"required"`
}

type ImportSummary struct {
	TotalTables     int           `json:"total_tables_detected"`
	TotalRowsFound  int           `json:"total_rows_found"`
	TotalSuccess    int           `json:"total_success"`
	TotalFailed     int           `json:"total_failed"`
	Errors          []RowError    `json:"errors,omitempty"`
	ParsingDuration time.Duration `json:"parsing_duration"`
	InsertDuration  time.Duration `json:"insert_duration"`
}

type RowError struct {
	Row   int    `json:"row"`
	Table int    `json:"table_index"`
	Msg   string `json:"message"`
}

type BranchCodeBankRepository interface {
	Insert(ctx context.Context, code *BranchCodeBank) error
	GetAll(ctx context.Context, bankName, search string, limit, offset int) ([]BranchCodeBank, int, error)
	FindByID(ctx context.Context, id int) (*BranchCodeBank, error)
	Update(ctx context.Context, code *BranchCodeBank) (*BranchCodeBank, error)
	Delete(ctx context.Context, id int) error
	BulkInsert(ctx context.Context, codes []BranchCodeBank) error
}

type ExcelParser interface {
	Parse(ctx context.Context, fileContent []byte) ([]BranchCodeBank, *ImportSummary, error)
}
