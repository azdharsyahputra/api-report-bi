package handler

import (
	"net/http"
	"strconv"

	"portal-report-bi/internal/domain"
	"portal-report-bi/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BranchBankHandler struct {
	service *service.BranchBankService
	logger  *zap.Logger
}

func NewBranchBankHandler(
	service *service.BranchBankService,
	logger *zap.Logger,
) *BranchBankHandler {

	return &BranchBankHandler{
		service: service,
		logger:  logger,
	}
}

func (h *BranchBankHandler) GetAll(c *gin.Context) {
	bankName := c.Query("bank_name")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	regencies, total, err := h.service.GetAll(c.Request.Context(), bankName, search, limit, offset)
	if err != nil {
		h.logger.Error("failed to get branch bank",
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
		"message": "Success fetch branch bank",
		"data":    regencies,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *BranchBankHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid ID parameter",
			},
		})
		return
	}

	var req domain.UpdateBranchCodeBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request payload: " + err.Error(),
			},
		})
		return
	}

	branchBank := &domain.BranchCodeBank{
		ID:            id,
		Name:          req.Name,
		BranchCode:    req.BranchCode,
		RegenciesCode: req.RegenciesCode,
		Regencies:     req.Regencies,
		OfficeType:    req.OfficeType,
	}

	updated, err := h.service.Update(c.Request.Context(), branchBank)
	if err != nil {
		h.logger.Error("failed to update branch bank",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
			},
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success update branch bank",
		"data":    updated,
	})
}
