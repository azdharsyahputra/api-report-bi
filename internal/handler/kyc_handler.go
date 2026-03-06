package handler

import (
	"net/http"

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
	data, err := h.service.GetAllDataKyc(c.Request.Context())
	if err != nil {
		h.log.Error("failed to get all kyc data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    data,
	})
}
