package handler

import (
	"io"
	"net/http"
	"portal-report-bi/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ImportHandler struct {
	service *service.ImportService
	logger  *zap.Logger
}

func NewImportHandler(service *service.ImportService, logger *zap.Logger) *ImportHandler {
	return &ImportHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ImportHandler) ImportBranchBankFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "File is required",
			},
		})
		return
	}
	defer file.Close()

	if !h.isExcel(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid file format, only .xlsm is allowed",
			},
		})
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Failed to read file",
			},
		})
		return
	}

	summary, err := h.service.ImportBranchBank(c.Request.Context(), header.Filename, content)
	if err != nil {
		h.logger.Error("Excel import failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Failed to process import: " + err.Error(),
			},
			"data": summary,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Import finished",
		"data":    summary,
	})
}

func (h *ImportHandler) isExcel(fileName string) bool {
	return len(fileName) > 5 && fileName[len(fileName)-5:] == ".xlsm"
}
