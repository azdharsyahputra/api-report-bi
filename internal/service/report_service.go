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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	/* No changes needed to s.branchRepo.GetAll call because it was already updated to take "", "", 0, 0 in a previous step, but let me check if I should use a specific value here */
	reports, total, err := s.repo.GetPayBankReport(ctx, startDate, endDate, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	branches, _, err := s.branchRepo.GetAll(ctx, "", "", 0, 0)
	if err != nil {
		return nil, 0, err
	}

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
