package service

import (
	"bytes"
	"context"
	"fmt"
	"portal-report-bi/internal/domain"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ReportService struct {
	repo       domain.ReportRepository
	branchRepo domain.BranchCodeBankRepository
	logger     *zap.Logger
}

func NewReportService(repo domain.ReportRepository, branchRepo domain.BranchCodeBankRepository, logger *zap.Logger) *ReportService {
	return &ReportService{
		repo:       repo,
		branchRepo: branchRepo,
		logger:     logger,
	}
}

func (s *ReportService) GetPayBankReport(ctx context.Context, startDate, endDate, search, bankTujuan string, limit, offset int) ([]domain.PayBankReport, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	reports, total, err := s.repo.GetPayBankReport(ctx, startDate, endDate, search, bankTujuan, limit, offset)
	if err != nil {
		s.logger.Error("failed to get pay bank report",
			zap.String("startDate", startDate),
			zap.String("endDate", endDate),
			zap.Error(err),
		)
		return nil, 0, err
	}

	branches, _, err := s.branchRepo.GetAll(ctx, "", "", 0, 0)
	if err != nil {
		s.logger.Error("failed to get branch bank for mapping", zap.Error(err))
		return nil, 0, err
	}

	kotaToRegencyCode := make(map[string]string)
	branchToRegencyCode := make(map[string]string)

	for _, b := range branches {
		if b.RegenciesCode != "" {
			exact := strings.ToUpper(strings.TrimSpace(b.Regencies))
			exact = strings.ReplaceAll(exact, "WILL KOTA ", "KOTA ")
			exact = strings.ReplaceAll(exact, "WIL. KOTA ", "KOTA ")
			exact = strings.ReplaceAll(exact, "WIL KOTA ", "KOTA ")
			kotaToRegencyCode[exact] = b.RegenciesCode

			clean := strings.ReplaceAll(exact, "KOTA ", "")
			clean = strings.ReplaceAll(clean, "KABUPATEN ", "")
			clean = strings.ReplaceAll(clean, "KAB. ", "")
			clean = strings.TrimSpace(clean)
			if clean != "" && kotaToRegencyCode[clean] == "" {
				kotaToRegencyCode[clean] = b.RegenciesCode
			}

			branchToRegencyCode[b.BranchCode] = b.RegenciesCode
		}
	}

	for i := range reports {
		branchCodePenerima := s.extractPrefix(reports[i].BankTujuan, reports[i].NoRek)
		if rc, ok := branchToRegencyCode[branchCodePenerima]; ok {
			reports[i].PrefixPenerima = rc
		} else {
			reports[i].PrefixPenerima = branchCodePenerima
		}
		kota := strings.ToUpper(strings.TrimSpace(reports[i].KotaPengirim))
		if rc, ok := kotaToRegencyCode[kota]; ok {
			reports[i].PrefixPengirim = rc
		} else {
			clean := strings.ReplaceAll(kota, "KOTA ", "")
			clean = strings.ReplaceAll(clean, "KABUPATEN ", "")
			clean = strings.ReplaceAll(clean, "KAB. ", "")
			clean = strings.TrimSpace(clean)
			if rc, ok := kotaToRegencyCode[clean]; ok {
				reports[i].PrefixPengirim = rc
			}
		}
	}

	return reports, total, nil
}

func (s *ReportService) GetMissingBranchReport(ctx context.Context, startDate, endDate, search, bankTujuan string, limit, offset int) ([]domain.MissingBranchReport, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	reports, _, err := s.repo.GetPayBankReport(ctx, startDate, endDate, search, bankTujuan, 0, 0)
	if err != nil {
		return nil, 0, err
	}

	branches, _, err := s.branchRepo.GetAll(ctx, bankTujuan, "", 0, 0)
	if err != nil {
		return nil, 0, err
	}

	branchToRegencyCode := make(map[string]string)
	for _, b := range branches {
		if b.RegenciesCode != "" {
			branchToRegencyCode[b.BranchCode] = b.RegenciesCode
		}
	}

	uniqueResults := make(map[string]domain.MissingBranchReport)
	for i := range reports {
		branchCodePenerima := s.extractPrefix(reports[i].BankTujuan, reports[i].NoRek)
		if branchCodePenerima == "" {
			continue
		}

		mappedPrefix := ""
		if rc, ok := branchToRegencyCode[branchCodePenerima]; ok {
			mappedPrefix = rc
		} else {
			mappedPrefix = branchCodePenerima
		}

		if mappedPrefix == branchCodePenerima {
			key := fmt.Sprintf("%s|%s", reports[i].BankTujuan, mappedPrefix)
			if _, exists := uniqueResults[key]; !exists {
				uniqueResults[key] = domain.MissingBranchReport{
					BankTujuan:     reports[i].BankTujuan,
					PrefixPenerima: mappedPrefix,
				}
			}
		}
	}

	var finalResults []domain.MissingBranchReport
	for _, res := range uniqueResults {
		finalResults = append(finalResults, res)
	}

	total := len(finalResults)
	if limit > 0 {
		end := offset + limit
		if end > total {
			end = total
		}
		if offset >= total {
			return []domain.MissingBranchReport{}, total, nil
		}
		finalResults = finalResults[offset:end]
	}

	return finalResults, total, nil
}

func (s *ReportService) ExportPayBankReport(ctx context.Context, startDate, endDate, bankTujuan string) ([]byte, error) {
	reports, _, err := s.GetPayBankReport(ctx, startDate, endDate, "", bankTujuan, 0, 0)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for _, r := range reports {
		line := s.buildExportLine(r)
		buf.WriteString(line + "\r\n")
	}

	return buf.Bytes(), nil
}

func (s *ReportService) ExportPayBankExcel(ctx context.Context, startDate, endDate, bankTujuan string) ([]byte, error) {
	reports, _, err := s.GetPayBankReport(ctx, startDate, endDate, "", bankTujuan, 0, 0)
	if err != nil {
		return nil, err
	}

	f, err := excelize.OpenFile("internal/template/AplikasiExcelLTDBBnewkosong.xlsm")
	if err != nil {
		return nil, fmt.Errorf("failed to open excel template: %v", err)
	}
	defer f.Close()

	sheet := "G0003"

	// Update period in LTDBB Header & G0003
	yearInt := 0
	monthInt := 0
	if len(startDate) >= 6 {
		yearInt, _ = strconv.Atoi(startDate[:4])
		monthInt, _ = strconv.Atoi(startDate[4:6])
	}

	// LTDBB Header
	f.SetCellValue("LTDBB Header", "E5", yearInt)
	f.SetCellValue("LTDBB Header", "F5", monthInt)

	// G0003 header
	f.SetCellValue(sheet, "C5", yearInt)
	f.SetCellValue(sheet, "D5", monthInt)

	// Update record count in Header
	f.SetCellValue("LTDBB Header", "E18", len(reports))

	for r := 7; r <= 9; r++ {
		for _, col := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}

	kabMap := make(map[string]string)
	tujuanMap := make(map[string]string)

	if sheet3Rows, err := f.GetRows("Sheet3"); err == nil {
		for i, row := range sheet3Rows {
			if i == 0 {
				continue
			}
			if len(row) > 26 {
				val := row[26]
				if len(val) >= 4 {
					kabMap[val[:4]] = val
				}
			}
			if len(row) > 81 {
				val := row[81]
				if len(val) > 0 {
					tujuanMap[val[:1]] = val
				}
			}
		}
	}

	lastRow := 7 + len(reports) - 1
	if lastRow < 9 {
		lastRow = 9
	}

	dvKab := excelize.NewDataValidation(true)
	dvKab.Sqref = fmt.Sprintf("B7:C%d", lastRow)
	dvKab.SetSqrefDropList("Kabupaten")
	f.AddDataValidation(sheet, dvKab)

	dvTujuan := excelize.NewDataValidation(true)
	dvTujuan.Sqref = fmt.Sprintf("H7:H%d", lastRow)
	dvTujuan.SetSqrefDropList("TujuanTransaksi")
	f.AddDataValidation(sheet, dvTujuan)

	// Cache styles from row 7
	colStyles := make(map[string]int)
	for _, col := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q"} {
		styleID, _ := f.GetCellStyle(sheet, col+"7")
		colStyles[col] = styleID
	}

	// Create custom style for "[Delete]" (Bold + Red) based on J7 style
	deleteStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FF0000",
		},
	})
	// Merge it with existing J7 style if possible (or just use it)
	// For simplicity, we create a new one, but we could try to inherit borders etc.

	for i, r := range reports {
		row := 7 + i

		pengirimPrefix := s.padLeftZero(r.PrefixPengirim, 4)
		pengirimDropdown := kabMap[pengirimPrefix]
		if pengirimDropdown == "" && pengirimPrefix != "" && pengirimPrefix != "0000" {
			pengirimDropdown = pengirimPrefix
		}

		penerimaPrefix := s.padLeftZero(r.PrefixPenerima, 4)
		penerimaDropdown := kabMap[penerimaPrefix]
		if penerimaDropdown == "" && penerimaPrefix != "" && penerimaPrefix != "0000" {
			penerimaDropdown = penerimaPrefix
		}

		if pengirimPrefix == "0000" || pengirimDropdown == "" {
			pengirimDropdown = ""
		}
		if penerimaPrefix == "0000" || penerimaDropdown == "" {
			penerimaDropdown = ""
		}

		tujuanVal := "3"
		tujuanDropdown := tujuanMap[tujuanVal]
		if tujuanDropdown == "" {
			tujuanDropdown = tujuanVal
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), cleanCellValue(pengirimDropdown))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), cleanCellValue(penerimaDropdown))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), cleanCellValue(strings.ToUpper(r.NamaPenerima)))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), cleanCellValue(strings.ToUpper(r.Pengirim)))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), r.Volume)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), cleanCellValue(r.Jumlah))
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), cleanCellValue(tujuanDropdown))

		// Populate dynamic formulas
		f.SetCellFormula(sheet, fmt.Sprintf("A%d", row), "ROW()-6")
		f.SetCellFormula(sheet, fmt.Sprintf("I%d", row), fmt.Sprintf(`LEFT(TRIM(B%d),4)&LEFT(TRIM(C%d),4)&LEFT(TRIM(D%d)&REPT(" ",50),50)&LEFT(TRIM(E%d)&REPT(" ",50),50)&RIGHT(REPT("0",12)&F%d,12)&RIGHT(REPT("0",15)&G%d,15)&LEFT(H%d,1)`, row, row, row, row, row, row, row))
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), " [Delete] ")
		f.SetCellFormula(sheet, fmt.Sprintf("M%d", row), fmt.Sprintf(`IF(OR(LEFT(C%d,1)="3",LEFT(C%d,1)="4"),$D$5-1,$D$5)`, row, row))
		f.SetCellFormula(sheet, fmt.Sprintf("O%d", row), fmt.Sprintf(`DATE($C$5,M%d+1,0)`, row))
		f.SetCellFormula(sheet, fmt.Sprintf("P%d", row), fmt.Sprintf(`DATE(Q%d,M%d,"01")`, row, row))
		f.SetCellFormula(sheet, fmt.Sprintf("Q%d", row), fmt.Sprintf(`IF(OR(LEFT(C%d,1)="3",LEFT(C%d,1)="4"),$C$5-2,$C$5)`, row, row))

		// Apply styles to all cells in the row
		for col, styleID := range colStyles {
			if col == "J" {
				f.SetCellStyle(sheet, col+strconv.Itoa(row), col+strconv.Itoa(row), deleteStyle)
			} else {
				f.SetCellStyle(sheet, col+strconv.Itoa(row), col+strconv.Itoa(row), styleID)
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel: %v", err)
	}

	return buf.Bytes(), nil
}

func (s *ReportService) buildExportLine(r domain.PayBankReport) string {
	prefixPengirim := s.padLeftZero(r.PrefixPengirim, 4)
	prefixPenerima := s.padLeftZero(r.PrefixPenerima, 4)
	namaPenerima := s.padRight(cleanCellValue(strings.ToUpper(r.NamaPenerima)), 50)
	namaPengirim := s.padRight(cleanCellValue(strings.ToUpper(r.Pengirim)), 50)
	volume := s.padLeftZero(strconv.FormatInt(r.Volume, 10), 15)
	nominal := s.padLeftZero(cleanCellValue(r.Jumlah), 12)
	tujuan := "3"

	return prefixPengirim + prefixPenerima + namaPenerima + namaPengirim + volume + nominal + tujuan
}

func (s *ReportService) padRight(sStr string, length int) string {
	if len(sStr) > length {
		return sStr[:length]
	}
	return fmt.Sprintf("%-*s", length, sStr)
}

func (s *ReportService) padLeftZero(sStr string, length int) string {
	if len(sStr) > length {
		return sStr[len(sStr)-length:]
	}
	return fmt.Sprintf("%0*s", length, sStr)
}

// cleanCellValue removes hidden newline/carriage-return characters from cell values
func cleanCellValue(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.TrimSpace(s)
}

func (s *ReportService) extractPrefix(bankName, noRek string) string {
	if noRek == "" {
		return ""
	}

	reNumeric := regexp.MustCompile(`[^0-9]`)
	cleanNoRek := reNumeric.ReplaceAllString(noRek, "")

	var length int
	switch bankName {
	case "BCA", "ARTHA", "BRI", "BSM", "DANAMON":
		length = 4
	case "BNI":
		length = 3
	case "MANDIRI", "CIMB NIAGA", "CIMB":
		length = 5
	default:
		return ""
	}

	if len(cleanNoRek) >= length {
		return cleanNoRek[:length]
	}

	return cleanNoRek
}
