package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"portal-report-bi/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReportHandler struct {
	service *service.ReportService
	logger  *zap.Logger
}

func NewReportHandler(service *service.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ReportHandler) GetPayBankReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	bankTujuan := c.Query("bank_tujuan")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "start_date and end_date query parameters are required in yyyymmdd format",
			},
		})
		return
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	report, total, err := h.service.GetPayBankReport(c.Request.Context(), startDate, endDate, search, bankTujuan, limit, offset)
	if err != nil {
		h.logger.Error("failed to get paybank report",
			zap.String("start_date", startDate),
			zap.String("end_date", endDate),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success fetch paybank report",
		"data":    report,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *ReportHandler) ExportPayBankReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "start_date and end_date query parameters are required in yyyymmdd format",
			},
		})
		return
	}

	bankTujuan := c.Query("bank_tujuan")
	data, err := h.service.ExportPayBankReport(c.Request.Context(), startDate, endDate, bankTujuan)
	if err != nil {
		h.logger.Error("failed to export paybank report",
			zap.String("start_date", startDate),
			zap.String("end_date", endDate),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})
		return
	}

	fileName := fmt.Sprintf("export_paybank_%s_%s.txt", startDate, endDate)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "text/plain")
	c.Data(http.StatusOK, "text/plain", data)
}

func (h *ReportHandler) ExportPayBankExcelReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "start_date and end_date query parameters are required in yyyymmdd format",
			},
		})
		return
	}

	bankTujuan := c.Query("bank_tujuan")
	data, err := h.service.ExportPayBankExcel(c.Request.Context(), startDate, endDate, bankTujuan)
	if err != nil {
		h.logger.Error("failed to export paybank excel report",
			zap.String("start_date", startDate),
			zap.String("end_date", endDate),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})
		return
	}

	fileName := fmt.Sprintf("LTDBB_paybank_%s_%s.xlsm", startDate, endDate)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/vnd.ms-excel.sheet.macroEnabled.12")
	c.Data(http.StatusOK, "application/vnd.ms-excel.sheet.macroEnabled.12", data)
}
