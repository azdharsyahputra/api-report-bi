package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Hardcoded secret key untuk signing JWT
const jwtSecretKey = "portal_report_bi_super_secret_key_2026"

// JWTClaims mendefinisikan payload yang ada di dalam token.
type JWTClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken membuat JWT token berdasarkan email.
func GenerateToken(email string) (string, error) {
	claims := JWTClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "portal-report-bi",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecretKey))
}

// AuthMiddleware memvalidasi JWT Bearer token di setiap request.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Authorization header missing or invalid format. Use: Bearer <token>",
				},
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Invalid or expired token",
				},
			})
			return
		}

		// Simpan email ke context agar bisa diakses handler jika perlu
		c.Set("email", claims.Email)
		c.Next()
	}
}
