package routes

import (
	handlers "github.com/4040www/NativeCloud_HR/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(report *gin.RouterGroup) {

	report.GET("/me", handlers.GenerateMyReport)           // GET /report/me
	report.GET("/department", handlers.GenerateDeptReport) // GET /report/department
	report.GET("/today", handlers.GetTodayRecords)         // GET /report/today

}
