package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"portal-report-bi/internal/domain"
	"portal-report-bi/internal/service"
)

type RegencyHandler struct {
	service *service.RegencyService
	logger  *zap.Logger
}

func NewRegencyHandler(
	service *service.RegencyService,
	logger *zap.Logger,
) *RegencyHandler {

	return &RegencyHandler{
		service: service,
		logger:  logger,
	}
}

func (h *RegencyHandler) Create(c *gin.Context) {

	var req domain.CreateRegencyRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		h.logger.Warn("invalid regency create request",
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request payload",
			},
		})

		return
	}

	err := h.service.Create(
		c.Request.Context(),
		req.BIID,
		req.RegencyName,
	)

	if err != nil {

		h.logger.Error("failed to create regency",
			zap.String("bi_id", req.BIID),
			zap.String("regency_name", req.RegencyName),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "ORA-00001") {

			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"message": "BI ID already exists",
				},
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})

		return
	}

	h.logger.Info("regency created successfully",
		zap.String("bi_id", req.BIID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Regency created successfully",
	})
}

func (h *RegencyHandler) Get(c *gin.Context) {

	regencies, err := h.service.Get(c.Request.Context())
	if err != nil {

		h.logger.Error("failed to get regencies",
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
		"message": "Success fetch regencies",
		"data":    regencies,
	})
}

func (h *RegencyHandler) FindByID(c *gin.Context) {

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {

		h.logger.Warn("invalid regency id parameter",
			zap.String("id", idParam),
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid ID parameter",
			},
		})

		return
	}

	regency, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {

		h.logger.Error("failed to find regency",
			zap.Int("id", id),
			zap.Error(err),
		)

		if errors.Is(err, domain.ErrRegencyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"message": "Regency not found",
				},
			})
			return
		}

		if errors.Is(err, domain.ErrInvalidId) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "Invalid ID",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success fetch regency",
		"data":    regency,
	})
}

func (h *RegencyHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {

		h.logger.Warn("invalid regency id parameter",
			zap.String("id", idParam),
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid ID parameter",
			},
		})

		return
	}

	var req domain.UpdateRegencyRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		h.logger.Warn("invalid regency update request",
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request payload",
			},
		})

		return
	}

	regency := &domain.Regency{
		ID:          id,
		BIID:        req.BIID,
		RegencyName: req.RegencyName,
	}

	updated, err := h.service.Update(c.Request.Context(), regency)
	if err != nil {

		h.logger.Error("failed to update regency",
			zap.String("bi_id", req.BIID),
			zap.String("regency_name", req.RegencyName),
			zap.Error(err),
		)

		if errors.Is(err, domain.ErrRegencyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"message": "Regency not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Regency updated successfully",
		"data":    updated,
	})
}

func (h *RegencyHandler) Delete(c *gin.Context) {

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {

		h.logger.Warn("invalid regency id parameter",
			zap.String("id", idParam),
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid ID parameter",
			},
		})

		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {

		h.logger.Error("failed to delete regency",
			zap.Int("id", id),
			zap.Error(err),
		)

		if errors.Is(err, domain.ErrRegencyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"message": "Regency not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})

		return
	}

	h.logger.Info("regency deleted successfully",
		zap.Int("id", id),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Regency deleted successfully",
	})
}
func (h *RegencyHandler) SyncAndGet(c *gin.Context) {

	data, err := h.service.SyncAndGetAll(c.Request.Context())
	if err != nil {

		h.logger.Error("failed to sync and get regencies",
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
		"message": "Success sync and fetch regencies",
		"data":    data,
	})
}
