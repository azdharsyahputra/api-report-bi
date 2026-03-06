package routes

import (
	"github.com/gin-gonic/gin"

	"portal-report-bi/internal/handler"
	"portal-report-bi/internal/middleware"
)

func RegisterRoutes(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	regencyHandler *handler.RegencyHandler,
	reportHandler *handler.ReportHandler,
	importHandler *handler.ImportHandler,
	branchBankHandler *handler.BranchBankHandler,
	kycHandler *handler.KycHandler,
) {
	// Enable CORS
	r.Use(middleware.CORSMiddleware())

	// Public: login endpoint (tidak perlu token)
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// Protected: semua route di bawah ini butuh JWT Bearer token
	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		regencies := protected.Group("/regencies")
		{
			regencies.POST("", regencyHandler.Create)
			regencies.GET("", regencyHandler.Get)
			regencies.GET("/:id", regencyHandler.FindByID)
			regencies.PUT("/:id", regencyHandler.Update)
			regencies.DELETE("/:id", regencyHandler.Delete)
			regencies.GET("/sync", regencyHandler.SyncAndGet)
		}

		reports := protected.Group("/reports")
		{
			reports.GET("/paybank", reportHandler.GetPayBankReport)
			reports.GET("/paybank/export-csv", reportHandler.ExportPayBankReport)
		}

		imports := protected.Group("/import")
		{
			imports.POST("/branch-bank-excel", importHandler.ImportBranchBankFile)
		}

		branchBank := protected.Group("/branch-bank")
		{
			branchBank.GET("", branchBankHandler.GetAll)
			branchBank.PUT("/:id", branchBankHandler.Update)
		}

		kyc := protected.Group("/kyc")
		{
			kyc.GET("", kycHandler.GetAllKyc)
		}
	}
}
