// report_routes.go
package routes

import (
	handler "github.com/4040www/NativeCloud_HR/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(router *gin.RouterGroup) {
	{
		router.GET("/myRecords/:userID", handler.GetMyTodayRecords)                                       // API #2
		router.GET("/historyRecords/:userID", handler.GetMyHistoryRecords)                                // API #3
		router.GET("/historyRecords/:userID/:startDate/:endDate", handler.GetMyPeriodRecords)             // API #4
		router.GET("/thisMonth/:department/:userID", handler.GetThisMonthTeam)                            // API #5
		router.GET("/thisWeek/:department/:userID", handler.GetThisWeekTeam)                              // API #6
		router.GET("/PeriodTime/:department/:startDate/:endDate/:userID", handler.GetCustomPeriodTeam)    // API #7
		router.GET("/AlertList/:startDate/:endDate/:userID", handler.GetAlertList)                        // API #8 新增
		router.GET("/inChargeDepartment/:userID", handler.GetInChargeDepartments)                         // API 取得自己能看到哪些部門
		router.GET("/summaryExportCSV/:department/:startDate/:endDate/:userID", handler.ExportSummaryCSV) // API #10
		router.GET("/summaryExportPDF/:department/:startDate/:endDate/:userID", handler.ExportSummaryPDF) // API #11
		router.GET("/myDepartments/:userID", handler.GetMyDepartments)                                    // API #11：取得使用者可查看的部門清單
		router.GET("/attendanceSummary", handler.FilterAttendanceSummary)                                 // API #12：依部門＋時間範圍查詢出勤摘要
		router.GET("/attendanceExportCSV", handler.ExportAttendanceSummaryCSV)                            // API #13：匯出 CSV 檔
		router.GET("/attendanceExportPDF", handler.ExportAttendanceSummaryPDF)                            // API #14：匯出 PDF 檔
	}
}
