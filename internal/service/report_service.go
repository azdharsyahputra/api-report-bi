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
)

type ReportService struct {
	repo       domain.ReportRepository
	branchRepo domain.BranchCodeBankRepository
}

func NewReportService(repo domain.ReportRepository, branchRepo domain.BranchCodeBankRepository) *ReportService {
	return &ReportService{
		repo:       repo,
		branchRepo: branchRepo,
	}
}

func (s *ReportService) GetPayBankReport(ctx context.Context, startDate, endDate string, limit, offset int) ([]domain.PayBankReport, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	fmt.Println("DEBUG SVC: Fetching PayBank report...")
	reports, total, err := s.repo.GetPayBankReport(ctx, startDate, endDate, limit, offset)
	if err != nil {
		fmt.Println("DEBUG SVC: GetPayBankReport ERROR:", err)
		return nil, 0, err
	}
	fmt.Println("DEBUG SVC: Got", len(reports), "report rows, total:", total)

	fmt.Println("DEBUG SVC: Fetching branch bank data...")
	branches, _, err := s.branchRepo.GetAll(ctx, "", "", 0, 0)
	if err != nil {
		fmt.Println("DEBUG SVC: branchRepo.GetAll ERROR:", err)
		return nil, 0, err
	}
	fmt.Println("DEBUG SVC: Got", len(branches), "branches")

	kotaToRegencyCode := make(map[string]string)
	branchToRegencyCode := make(map[string]string)

	for _, b := range branches {
		if b.RegenciesCode != "" {
			exact := strings.ToUpper(strings.TrimSpace(b.Regencies))
			// Hapus kotoran seperti "WILL KOTA ", "WIL. KOTA ", "WIL KOTA "
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
		// prefix_penerima mapping
		branchCodePenerima := s.extractPrefix(reports[i].BankTujuan, reports[i].NoRek)
		if rc, ok := branchToRegencyCode[branchCodePenerima]; ok {
			reports[i].PrefixPenerima = rc
		} else {
			reports[i].PrefixPenerima = branchCodePenerima
		}

		// prefix_pengirim mapping
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

func (s *ReportService) ExportPayBankReport(ctx context.Context, startDate, endDate string) ([]byte, error) {
	reports, _, err := s.GetPayBankReport(ctx, startDate, endDate, 0, 0)
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

func (s *ReportService) ExportPayBankExcel(ctx context.Context, startDate, endDate string) ([]byte, error) {
	reports, _, err := s.GetPayBankReport(ctx, startDate, endDate, 0, 0)
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
	year := ""
	month := ""
	if len(startDate) >= 6 {
		year = startDate[:4]
		month = startDate[4:6]
	}

	// LTDBB Header
	f.SetCellValue("LTDBB Header", "E5", year)
	f.SetCellValue("LTDBB Header", "F5", month)

	// G0003 header
	f.SetCellValue(sheet, "C5", year)
	f.SetCellValue(sheet, "D5", month)

	// Update record count
	f.SetCellValue(sheet, "K4", len(reports))
	f.SetCellValue("LTDBB Header", "E18", len(reports))

	// Clear existing template data rows (rows 7-9)
	for r := 7; r <= 9; r++ {
		for _, col := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}

	// Build prefix-to-dropdown map from Sheet3 (Column AA)
	kabMap := make(map[string]string)
	if sheet3Rows, err := f.GetRows("Sheet3"); err == nil {
		for i, row := range sheet3Rows {
			if i == 0 {
				continue
			}
			if len(row) > 26 {
				val := row[26] // AA is index 26
				if len(val) >= 4 {
					kabMap[val[:4]] = val
				}
			}
		}
	}

	// Write report data starting at row 7
	for i, r := range reports {
		row := 7 + i

		pengirimPrefix := s.padLeftZero(r.PrefixPengirim, 4)
		pengirimDropdown := kabMap[pengirimPrefix]
		if pengirimDropdown == "" && pengirimPrefix != "" && pengirimPrefix != "0000" {
			pengirimDropdown = pengirimPrefix // Fallback if not found in template
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

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), pengirimDropdown)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), penerimaDropdown)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), strings.ToUpper(r.NamaPenerima))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), strings.ToUpper(r.Pengirim))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), r.Volume)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), r.Jumlah)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "3")
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel: %v", err)
	}

	return buf.Bytes(), nil
}

func (s *ReportService) buildExportLine(r domain.PayBankReport) string {
	// Logic from user request
	prefixPengirim := s.padLeftZero(r.PrefixPengirim, 4)
	prefixPenerima := s.padLeftZero(r.PrefixPenerima, 4)
	namaPenerima := s.padRight(strings.ToUpper(r.NamaPenerima), 50)
	namaPengirim := s.padRight(strings.ToUpper(r.Pengirim), 50)
	volume := s.padLeftZero(strconv.FormatInt(r.Volume, 10), 15)
	nominal := s.padLeftZero(r.Jumlah, 12)
	tujuan := "3" // Default based on user example

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
	case "MANDIRI", "CIMB NIAGA":
		length = 5
	default:
		return ""
	}

	if len(cleanNoRek) >= length {
		return cleanNoRek[:length]
	}

	return cleanNoRek
}
