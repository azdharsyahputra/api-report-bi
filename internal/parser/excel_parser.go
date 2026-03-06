package parser

import (
	"bytes"
	"context"
	"fmt"
	"portal-report-bi/internal/domain"
	"regexp"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type excelParser struct{}

func NewExcelParser() domain.ExcelParser {
	return &excelParser{}
}

func (p *excelParser) Parse(ctx context.Context, fileContent []byte) ([]domain.BranchCodeBank, *domain.ImportSummary, error) {
	start := time.Now()
	summary := &domain.ImportSummary{}

	f, err := excelize.OpenReader(bytes.NewReader(fileContent))
	if err != nil {
		return nil, summary, fmt.Errorf("failed to open excel: %w", err)
	}
	defer f.Close()

	var allBranches []domain.BranchCodeBank
	sheets := f.GetSheetList()

	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}

		if len(rows) == 0 {
			continue
		}

		for r, row := range rows {
			for c, cell := range row {
				if strings.TrimSpace(strings.ToUpper(cell)) == "BANK" {

					if p.isHeaderMatch(row, c) {
						summary.TotalTables++
						tableBranches, tableErrors := p.extractTable(rows, r, c, summary.TotalTables)

						allBranches = append(allBranches, tableBranches...)
						summary.Errors = append(summary.Errors, tableErrors...)
						summary.TotalRowsFound += len(tableBranches) + len(tableErrors)
						summary.TotalSuccess += len(tableBranches)
						summary.TotalFailed += len(tableErrors)
					}
				}
			}
		}
	}

	summary.ParsingDuration = time.Since(start)
	return allBranches, summary, nil
}

func (p *excelParser) isHeaderMatch(row []string, colIndex int) bool {
	expected := []string{"BRANCH KODE", "KOTA/KABUPATEN", "KCP/KCU"}
	if len(row) < colIndex+4 {
		return false
	}

	for i, head := range expected {
		actual := strings.TrimSpace(strings.ToUpper(row[colIndex+i+1]))
		if actual != head {
			return false
		}
	}
	return true
}

func (p *excelParser) extractTable(rows [][]string, startRow, startCol int, tableIdx int) ([]domain.BranchCodeBank, []domain.RowError) {
	var results []domain.BranchCodeBank
	var errors []domain.RowError

	for r := startRow + 1; r < len(rows); r++ {
		row := rows[r]
		if len(row) <= startCol || strings.TrimSpace(row[startCol]) == "" {
			break
		}

		bankName := p.cleanString(row[startCol])
		branchCodeRaw := ""
		if len(row) > startCol+1 {
			branchCodeRaw = p.cleanString(row[startCol+1])
		}

		city := ""
		if len(row) > startCol+2 {
			city = p.cleanString(row[startCol+2])
		}

		officeType := ""
		if len(row) > startCol+3 {
			officeType = p.cleanString(row[startCol+3])
		}

		if bankName == "" || branchCodeRaw == "" {
			errors = append(errors, domain.RowError{
				Row:   r + 1,
				Table: tableIdx,
				Msg:   "BANK or BRANCH KODE is missing or empty after normalization",
			})
			continue
		}

		branchCode := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, branchCodeRaw)

		if branchCode == "" {
			errors = append(errors, domain.RowError{
				Row:   r + 1,
				Table: tableIdx,
				Msg:   fmt.Sprintf("Invalid Branch Code (Empty After Cleaning): %s", branchCodeRaw),
			})
			continue
		}

		regencyCode := ""
		regencyName := city
		re := regexp.MustCompile(`^(\d+)\s*[-]\s*(.*)$`)
		matches := re.FindStringSubmatch(city)
		if len(matches) > 2 {
			regencyCode = matches[1]
			regencyName = strings.TrimSpace(matches[2])
		} else {
			reLeading := regexp.MustCompile(`^(\d+)`)
			leadingMatches := reLeading.FindStringSubmatch(city)
			if len(leadingMatches) > 1 {
				regencyCode = leadingMatches[1]
			}
		}

		regencyName = strings.ReplaceAll(regencyName, "WIL. KOTA ", "KOTA ")
		regencyName = strings.ReplaceAll(regencyName, "WIL. KAB. ", "KABUPATEN ")
		if strings.HasPrefix(regencyName, "KAB. ") {
			regencyName = "KABUPATEN " + strings.TrimPrefix(regencyName, "KAB. ")
		}
		regencyName = strings.TrimSpace(regencyName)

		results = append(results, domain.BranchCodeBank{
			Name:          bankName,
			BranchCode:    branchCode,
			RegenciesCode: regencyCode,
			Regencies:     regencyName,
			OfficeType:    officeType,
			CreatedAt:     time.Now().Format("2006-01-02 15:04:05"),
			UpdateAt:      time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	return results, errors
}

func (p *excelParser) cleanString(s string) string {

	s = strings.TrimSpace(s)

	return strings.ToUpper(s)
}
