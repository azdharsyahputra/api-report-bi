package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"portal-report-bi/internal/config"
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

	excelParser := parser.NewExcelParser()
	branchRepo := repository.NewBranchCodeBankRepository(queryServiceURL)
	reportRepo := repository.NewReportRepository(queryServiceURL)
	reportService := service.NewReportService(reportRepo, branchRepo, logger)
	reportHandler := handler.NewReportHandler(reportService, logger)

	authRepo := repository.NewAuthRepository()
	authService := service.NewAuthService(authRepo, logger)
	authHandler := handler.NewAuthHandler(authService, logger)
	importService := service.NewImportService(excelParser, branchRepo, logger)
	importHandler := handler.NewImportHandler(importService, logger)

	branchBankService := service.NewBranchBankService(branchRepo, logger)
	branchBankHandler := handler.NewBranchBankHandler(branchBankService, logger)

	kycRepo := repository.NewKycRepository(queryServiceURL)
	kycService := service.NewKycService(kycRepo, logger)
	kycHandler := handler.NewKycHandler(kycService, logger)

	r := gin.Default()

	routes.RegisterRoutes(r, authHandler, reportHandler, importHandler, branchBankHandler, kycHandler)
	log.Println("server running on :8080")
	log.Fatal(r.Run(":8080"))
}
