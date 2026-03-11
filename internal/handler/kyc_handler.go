package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"portal-report-bi/internal/service"
)

type KycHandler struct {
	service *service.KycService
	log     *zap.Logger
}

func NewKycHandler(service *service.KycService, log *zap.Logger) *KycHandler {
	return &KycHandler{
		service: service,
		log:     log,
	}
}

func (h *KycHandler) GetAllKyc(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	search := c.Query("search")

	// set default date if empty
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -7).Format("20060102")
	}
	if endDate == "" {
		endDate = time.Now().Format("20060102")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	data, total, err := h.service.GetAllDataKyc(c.Request.Context(), startDate, endDate, search, limit, offset)
	if err != nil {
		h.log.Error("failed to get all kyc data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    data,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
