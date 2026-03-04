package main

import (
	"log"
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

	db := config.InitDB()

	logger := config.InitLogger()
	defer logger.Sync()

	regencyRepo := repository.NewRegencyRepository(db)

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
	branchRepo := repository.NewBranchCodeBankRepository(db)
	reportRepo := repository.NewReportRepository(db)
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

	log.Println("server running on :8080")
	log.Fatal(r.Run(":8080"))
}
