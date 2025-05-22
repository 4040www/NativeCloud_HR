// report_routes.go
package routes

import (
	handler "github.com/4040www/NativeCloud_HR/internal/api/handler"
	"github.com/4040www/NativeCloud_HR/internal/db"
	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(router *gin.RouterGroup) {
	{
		router.GET("/myRecords/:userID", handler.GetMyTodayRecords(db.GetDB()))                                       // API #2
		router.GET("/historyRecords/:userID", handler.GetMyHistoryRecords(db.GetDB()))                                // API #3
		router.GET("/historyRecords/:userID/:startDate/:endDate", handler.GetMyPeriodRecords(db.GetDB()))             // API #4
		router.GET("/thisMonth/:department/:userID", handler.GetThisMonthTeam(db.GetDB()))                            // API #5 //變成可以填月和週
		router.GET("/thisWeek/:department/:userID", handler.GetThisWeekTeam(db.GetDB()))                              // API #6
		router.GET("/PeriodTime/:department/:startDate/:endDate/:userID", handler.GetCustomPeriodTeam(db.GetDB()))    // API #7
		router.GET("/AlertList/:startDate/:endDate/:userID", handler.GetAlertList(db.GetDB()))                        // API #8 新增
		router.GET("/inChargeDepartment/:userID", handler.GetInChargeDepartments)                                     // API 取得自己能看到哪些部門
		router.GET("/summaryExportCSV/:department/:startDate/:endDate/:userID", handler.ExportSummaryCSV(db.GetDB())) // API #10
		router.GET("/summaryExportPDF/:department/:startDate/:endDate/:userID", handler.ExportSummaryPDF(db.GetDB())) // API #11
		router.GET("/myDepartments/:userID", handler.GetMyDepartments)                                                // API #11：取得使用者可查看的部門清單
		router.GET("/attendanceSummary", handler.FilterAttendanceSummary(db.GetDB()))                                 // API #12：依部門＋時間範圍查詢出勤摘要
		router.GET("/attendanceExportCSV", handler.ExportAttendanceSummaryCSV(db.GetDB()))                            // API #13：匯出 CSV 檔
		router.GET("/attendanceExportPDF", handler.ExportAttendanceSummaryPDF(db.GetDB()))                            // API #14：匯出 PDF 檔
	}
}

// 需要傳部門的陣列進來
// router.GET("/summaryExportCSV/:department/:startDate/:endDate/:userID", handler.ExportSummaryCSV) // API #10
// router.GET("/summaryExportPDF/:department/:startDate/:endDate/:userID", handler.ExportSummaryPDF) // API #11
// router.GET("/attendanceSummary", handler.FilterAttendanceSummary)                                 // API #12：依部門＋時間範圍查詢出勤摘要
// router.GET("/attendanceExportCSV", handler.ExportAttendanceSummaryCSV)                            // API #13：匯出 CSV 檔
// router.GET("/attendanceExportPDF", handler.ExportAttendanceSummaryPDF)                            // API #14：匯出 PDF 檔
