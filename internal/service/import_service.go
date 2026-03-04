package service

import (
	"context"
	"fmt"
	"portal-report-bi/internal/domain"
	"time"

	"go.uber.org/zap"
)

type ImportService struct {
	parser domain.ExcelParser
	repo   domain.BranchCodeBankRepository
	logger *zap.Logger
}

func NewImportService(parser domain.ExcelParser, repo domain.BranchCodeBankRepository, logger *zap.Logger) *ImportService {
	return &ImportService{
		parser: parser,
		repo:   repo,
		logger: logger,
	}
}

func (s *ImportService) ImportBranchBank(ctx context.Context, fileName string, fileContent []byte) (*domain.ImportSummary, error) {
	s.logger.Info("Starting excel import",
		zap.String("fileName", fileName),
		zap.Int("fileSize", len(fileContent)),
	)

	branches, summary, err := s.parser.Parse(ctx, fileContent)
	if err != nil {
		s.logger.Error("Excel parsing failed", zap.String("fileName", fileName), zap.Error(err))
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	s.logger.Info("Excel parsing completed",
		zap.String("fileName", fileName),
		zap.Int("tables_detected", summary.TotalTables),
		zap.Int("rows_found", summary.TotalRowsFound),
		zap.Duration("parsing_duration", summary.ParsingDuration),
	)

	if len(branches) == 0 {
		return summary, nil
	}

	startInsert := time.Now()
	err = s.repo.BulkInsert(ctx, branches)
	summary.InsertDuration = time.Since(startInsert)

	if err != nil {
		s.logger.Error("Bulk insert failed",
			zap.String("fileName", fileName),
			zap.Error(err),
			zap.Int("attempted_rows", len(branches)),
		)
		return summary, fmt.Errorf("database insert failed: %w", err)
	}

	s.logger.Info("Import successfully finished",
		zap.String("fileName", fileName),
		zap.Int("rows_inserted", len(branches)),
		zap.Duration("insert_duration", summary.InsertDuration),
	)

	return summary, nil
}
