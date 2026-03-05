package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"portal-report-bi/internal/config"
	"portal-report-bi/internal/external"
	"portal-report-bi/internal/handler"
	"portal-report-bi/internal/parser"
	"portal-report-bi/internal/repository"
	"portal-report-bi/internal/routes"
	"portal-report-bi/internal/service"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	queryServiceURL := os.Getenv("QUERY_SERVICE_URL")
	if queryServiceURL == "" {
		log.Fatal("QUERY_SERVICE_URL is required in .env file")
	}

	logger := config.InitLogger()
	defer logger.Sync()

	regencyRepo := repository.NewRegencyRepository(queryServiceURL)

	regencyAPIURL := os.Getenv("REGENCY_API_URL")
	if regencyAPIURL == "" {
		regencyAPIURL = "http://localhost:8081"
	}
	regencyAPIClient := external.NewRegencyAPIClient(regencyAPIURL)

	regencyService := service.NewRegencyService(regencyRepo, regencyAPIClient)

	regencyHandler := handler.NewRegencyHandler(
		regencyService,
		logger,
	)

	excelParser := parser.NewExcelParser()
	branchRepo := repository.NewBranchCodeBankRepository(queryServiceURL)
	reportRepo := repository.NewReportRepository(queryServiceURL)
	reportService := service.NewReportService(reportRepo, branchRepo)
	reportHandler := handler.NewReportHandler(reportService, logger)

	authRepo := repository.NewAuthRepository()
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService, logger)
	importService := service.NewImportService(excelParser, branchRepo, logger)
	importHandler := handler.NewImportHandler(importService, logger)

	branchBankService := service.NewBranchBankService(branchRepo)
	branchBankHandler := handler.NewBranchBankHandler(branchBankService, logger)

	r := gin.Default()

	routes.RegisterRoutes(r, authHandler, regencyHandler, reportHandler, importHandler, branchBankHandler)

	r.GET("/test-regencies", func(c *gin.Context) {
		query := "SELECT * FROM vdapp_3.regencies"
		body, _ := json.Marshal(map[string]string{"qstr": query})
		resp, err := http.Post(queryServiceURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"query_service_error": string(b)})
			return
		}

		var result interface{}
		if err := json.Unmarshal(b, &result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON Unmarshal failed", "raw_response": string(b)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	})

	r.POST("/debug-query", func(c *gin.Context) {
		var req struct {
			Query string `json:"query"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parse error: " + err.Error()})
			return
		}

		body, _ := json.Marshal(map[string]string{"qstr": req.Query})
		resp, err := http.Post(queryServiceURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"query_service_error": string(b)})
			return
		}

		var result interface{}
		if err := json.Unmarshal(b, &result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON Unmarshal failed", "raw_response": string(b)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	})

	log.Println("server running on :8080")
	log.Fatal(r.Run(":8080"))
}
