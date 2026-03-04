package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"portal-report-bi/internal/domain"
	"portal-report-bi/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
	logger  *zap.Logger
}

func NewAuthHandler(svc *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{service: svc, logger: logger}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request payload",
				"details": err.Error(),
			},
		})
		return
	}

	token, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			h.logger.Warn("failed login attempt", zap.String("email", req.Email))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Invalid email or password",
				},
			})
			return
		}

		h.logger.Error("login error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
			},
		})
		return
	}

	h.logger.Info("user logged in", zap.String("email", req.Email))

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    domain.LoginResponse{Token: token},
	})
}
