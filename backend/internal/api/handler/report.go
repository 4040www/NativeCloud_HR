// report_handlers.go
package handlers

import (
	"net/http"
<<<<<<< HEAD
=======
	"strconv"
>>>>>>> architecture
	"time"

	"github.com/4040www/NativeCloud_HR/internal/service"
	"github.com/gin-gonic/gin"
<<<<<<< HEAD
	"gorm.io/gorm"
=======
>>>>>>> architecture
)

type AttendanceSummary struct {
	Date         string `json:"date"`
	Name         string `json:"name"`
	ClockInTime  string `json:"clock_in_time"`
	ClockOutTime string `json:"clock_out_time"`
	ClockInGate  string `json:"clock_in_gate"`
	ClockOutGate string `json:"clock_out_gate"`
	Status       string `json:"status"` // "On Time" / "Late" / "Abnormal"
}

<<<<<<< HEAD
func GetMyTodayRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")

		summary, err := service.GetTodayAttendanceSummary(db, userID)
=======
func GetMyTodayRecords(c *gin.Context) {
	userID := c.Param("userID")

	summary, err := service.GetTodayAttendanceSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if summary == nil {
		c.JSON(http.StatusOK, gin.H{"message": "no access records today"})
		return
	}
	c.JSON(http.StatusOK, summary)

}

func GetMyHistoryRecords(c *gin.Context) {
	userID := c.Param("userID")
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	summaries, err := service.GetAttendanceWithEmployee(userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summaries)

}

func GetMyPeriodRecords(c *gin.Context) {
	userID := c.Param("userID")
	startDate := c.Param("startDate")
	endDate := c.Param("endDate")

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startDate"})
		return
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endDate"})
		return
	}

	summaries, err := service.GetAttendanceWithEmployee(userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summaries)

}

func GetThisMonthTeam(c *gin.Context) {
	department := c.Param("department")
	monthsStr := c.DefaultQuery("months", "1")
	months, err := strconv.Atoi(monthsStr)
	if err != nil || months < 1 {
		months = 1
	}

	today := time.Now()
	var reports []interface{}

	for i := 0; i < months; i++ {
		targetMonth := today.AddDate(0, -i, 0).Format("2006-01")

		report, err := service.FetchMonthlyTeamReport(department, targetMonth)
>>>>>>> architecture
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
<<<<<<< HEAD
		if summary == nil {
			c.JSON(http.StatusOK, gin.H{"data": nil, "message": "No access records for today"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": summary})
	}
}

// func GetMyTodayRecords(c *gin.Context) {
// 	userID := c.Param("userID")

// 	logs, err := service.FetchTodayRecords(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if len(logs) == 0 {
// 		c.JSON(http.StatusOK, gin.H{"message": "no access records today"})
// 		return
// 	}

// 	emp, err := repository.GetEmployeeByID(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "employee not found"})
// 		return
// 	}

// 	var clockIn, clockOut *model.AccessLog
// 	isLate := false

// 	for _, log := range logs {
// 		if log.Direction == "IN" && (clockIn == nil || log.AccessTime.Before(clockIn.AccessTime)) {
// 			clockIn = &log
// 			if log.AccessTime.Hour() > 8 || (log.AccessTime.Hour() == 8 && log.AccessTime.Minute() > 30) {
// 				isLate = true
// 			}
// 		}
// 		if log.Direction == "OUT" && (clockOut == nil || log.AccessTime.After(clockOut.AccessTime)) {
// 			clockOut = &log
// 		}
// 	}

// 	resp := gin.H{
// 		"date":           clockIn.AccessTime.Format("2006-01-02"),
// 		"name":           emp.FirstName + " " + emp.LastName,
// 		"clock_in_time":  clockIn.AccessTime.Format("15:04"),
// 		"clock_out_time": clockOut.AccessTime.Format("15:04"),
// 		"clock_in_gate":  clockIn.GateName,
// 		"clock_out_gate": clockOut.GateName,
// 		"status":         "Normal",
// 	}
// 	if isLate {
// 		resp["status"] = "Late"
// 	}

// 	c.JSON(http.StatusOK, resp)
// }

func GetMyHistoryRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		end := time.Now()
		start := end.AddDate(0, 0, -30)

		summaries, err := service.GetAttendanceWithEmployee(db, userID, start, end)
=======

		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, reports)
}

func GetThisWeekTeam(c *gin.Context) {
	department := c.Param("department")
	weeksStr := c.DefaultQuery("weeks", "1")
	weeks, err := strconv.Atoi(weeksStr)
	if err != nil || weeks < 1 {
		weeks = 1
	}

	today := time.Now()
	var reports []interface{}

	for i := 0; i < weeks; i++ {
		end := today.AddDate(0, 0, -7*i)
		start := end.AddDate(0, 0, -int(end.Weekday())+1)

		report, err := service.FetchWeeklyTeamReport(department, start.Format("2006-01-02"), end.Format("2006-01-02"))
>>>>>>> architecture
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
<<<<<<< HEAD
		c.JSON(http.StatusOK, summaries)
	}
}

// func GetMyHistoryRecords(c *gin.Context) {
// 	userID := c.Param("userID")

// 	// 計算預設區間：今天往前 30 天
// 	end := time.Now()
// 	start := end.AddDate(0, 0, -30)
// 	startDate := start.Format("2006-01-02")
// 	endDate := end.Format("2006-01-02")

// 	records, err := service.FetchHistoryRecordsBetween(userID, startDate, endDate)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	emp, err := repository.GetEmployeeByID(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "employee not found"})
// 		return
// 	}

// 	dateMap := make(map[string][]model.AccessLog)
// 	for _, r := range records {
// 		day := r.AccessTime.Format("2006-01-02")
// 		dateMap[day] = append(dateMap[day], r)
// 	}

// 	var results []AttendanceSummary
// 	for date, logs := range dateMap {
// 		var clockIn, clockOut *model.AccessLog
// 		status := "On Time"

// 		for _, log := range logs {
// 			if log.Direction == "IN" && (clockIn == nil || log.AccessTime.Before(clockIn.AccessTime)) {
// 				clockIn = &log
// 				if log.AccessTime.Hour() > 8 || (log.AccessTime.Hour() == 8 && log.AccessTime.Minute() > 30) {
// 					status = "Late"
// 				}
// 			}
// 			if log.Direction == "OUT" && (clockOut == nil || log.AccessTime.After(clockOut.AccessTime)) {
// 				clockOut = &log
// 			}
// 		}

// 		if clockIn == nil || clockOut == nil {
// 			status = "Abnormal"
// 		}

// 		results = append(results, AttendanceSummary{
// 			Date:         date,
// 			Name:         emp.FirstName + " " + emp.LastName,
// 			ClockInTime:  formatTime(clockIn),
// 			ClockOutTime: formatTime(clockOut),
// 			ClockInGate:  getGate(clockIn),
// 			ClockOutGate: getGate(clockOut),
// 			Status:       status,
// 		})
// 	}

// 	c.JSON(http.StatusOK, results)
// }

func GetMyPeriodRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		startDate := c.Param("startDate")
		endDate := c.Param("endDate")

		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startDate"})
			return
		}
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endDate"})
			return
		}

		summaries, err := service.GetAttendanceWithEmployee(db, userID, start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, summaries)
	}
}

// func GetMyPeriodRecords(c *gin.Context) {
// 	userID := c.Param("userID")
// 	startDate := c.Param("startDate")
// 	endDate := c.Param("endDate")

// 	records, err := service.FetchHistoryRecordsBetween(userID, startDate, endDate)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	emp, err := repository.GetEmployeeByID(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "employee not found"})
// 		return
// 	}

// 	dateMap := make(map[string][]model.AccessLog)
// 	for _, r := range records {
// 		day := r.AccessTime.Format("2006-01-02")
// 		dateMap[day] = append(dateMap[day], r)
// 	}

// 	var results []AttendanceSummary
// 	for date, logs := range dateMap {
// 		var clockIn, clockOut *model.AccessLog
// 		status := "On Time"

// 		for _, log := range logs {
// 			if log.Direction == "IN" && (clockIn == nil || log.AccessTime.Before(clockIn.AccessTime)) {
// 				clockIn = &log
// 				if log.AccessTime.Hour() > 8 || (log.AccessTime.Hour() == 8 && log.AccessTime.Minute() > 30) {
// 					status = "Late"
// 				}
// 			}
// 			if log.Direction == "OUT" && (clockOut == nil || log.AccessTime.After(clockOut.AccessTime)) {
// 				clockOut = &log
// 			}
// 		}

// 		if clockIn == nil || clockOut == nil {
// 			status = "Abnormal"
// 		}

// 		results = append(results, AttendanceSummary{
// 			Date:         date,
// 			Name:         emp.FirstName + " " + emp.LastName,
// 			ClockInTime:  formatTime(clockIn),
// 			ClockOutTime: formatTime(clockOut),
// 			ClockInGate:  getGate(clockIn),
// 			ClockOutGate: getGate(clockOut),
// 			Status:       status,
// 		})
// 	}

// 	c.JSON(http.StatusOK, results)
// }

// func formatTime(log *model.AccessLog) string {
// 	if log == nil {
// 		return ""
// 	}
// 	return log.AccessTime.Format("15:04")
// }

// func getGate(log *model.AccessLog) string {
// 	if log == nil {
// 		return ""
// 	}
// 	return log.GateName
// }

func GetThisMonthTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Param("department")
		month := c.DefaultQuery("month", time.Now().Format("2006-01"))

		current, prev, err := service.FetchMonthComparisonReport(db, department, month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, []interface{}{current, prev})
	}
}

func GetThisWeekTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Param("department")
		now := time.Now()

		start := now.AddDate(0, 0, -int(now.Weekday())+1) // 本週一
		end := start.AddDate(0, 0, 6)                     // 本週日
		lastStart := start.AddDate(0, 0, -7)              // 上週一
		lastEnd := end.AddDate(0, 0, -7)                  // 上週日

		current, err1 := service.FetchWeeklyTeamReport(db, department, start.Format("2006-01-02"), end.Format("2006-01-02"))
		prev, err2 := service.FetchWeeklyTeamReport(db, department, lastStart.Format("2006-01-02"), lastEnd.Format("2006-01-02"))
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load report"})
			return
		}
		c.JSON(http.StatusOK, []interface{}{current, prev})
	}
}

func GetCustomPeriodTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Param("department")
		startDate := c.Param("startDate")
		endDate := c.Param("endDate")

		result, err := service.FetchCustomPeriodTeamReport(db, department, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func GetAlertList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := c.Param("startDate")
		end := c.Param("endDate")

		list, err := service.GenerateAlertList(db, start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, list)
	}
=======

		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, reports)
}

func GetCustomPeriodTeam(c *gin.Context) {
	department := c.Param("department")
	startDate := c.Param("startDate")
	endDate := c.Param("endDate")
	result, err := service.FetchCustomPeriodTeamReport(department, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)

}

func GetAlertList(c *gin.Context) {
	start := c.Param("startDate")
	end := c.Param("endDate")
	list, err := service.GenerateAlertList(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
>>>>>>> architecture
}

func GetInChargeDepartments(c *gin.Context) {
	userID := c.Param("userID")
	depts := service.GetManagedDepartments(userID)
	c.JSON(http.StatusOK, depts)
}

<<<<<<< HEAD
func ExportSummaryCSV(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Param("department")
		start := c.Param("startDate")
		end := c.Param("endDate")

		data, err := service.GenerateAttendanceSummaryCSV(db, department, start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Disposition", "attachment; filename=summary.csv")
		c.Data(http.StatusOK, "text/csv", data)
	}
}

func ExportSummaryPDF(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Param("department")
		start := c.Param("startDate")
		end := c.Param("endDate")

		data, err := service.GenerateAttendanceSummaryPDF(db, department, start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Disposition", "attachment; filename=summary.pdf")
		c.Data(http.StatusOK, "application/pdf", data)
	}
=======
func ExportSummaryCSV(c *gin.Context) {
	department := c.Param("department")
	start := c.Param("startDate")
	end := c.Param("endDate")
	data, err := service.GenerateAttendanceSummaryCSV(department, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=summary.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

func ExportSummaryPDF(c *gin.Context) {
	department := c.Param("department")
	start := c.Param("startDate")
	end := c.Param("endDate")
	data, err := service.GenerateAttendanceSummaryPDF(department, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=summary.pdf")
	c.Data(http.StatusOK, "application/pdf", data)
>>>>>>> architecture
}

// Page: Attendence Log
func GetMyDepartments(c *gin.Context) {
	userID := c.Param("userID")
	departments, err := service.GetManagedDepartmentsFromDB(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, departments)
}

<<<<<<< HEAD
func FilterAttendanceSummary(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Query("department")
		fromDate := c.Query("fromDate")
		toDate := c.Query("toDate")

		result, err := service.GetAttendanceSummaryForDepartments(db, department, fromDate, toDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func ExportAttendanceSummaryCSV(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Query("department")
		fromDate := c.Query("fromDate")
		toDate := c.Query("toDate")

		csvData, err := service.GenerateAttendanceSummaryCSV(db, department, fromDate, toDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Disposition", "attachment; filename=summary.csv")
		c.Data(http.StatusOK, "text/csv", csvData)
	}
}

func ExportAttendanceSummaryPDF(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		department := c.Query("department")
		fromDate := c.Query("fromDate")
		toDate := c.Query("toDate")

		pdfData, err := service.GenerateAttendanceSummaryPDF(db, department, fromDate, toDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Disposition", "attachment; filename=summary.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	}
=======
func FilterAttendanceSummary(c *gin.Context) {
	department := c.Query("department")
	fromDate := c.Query("fromDate")
	toDate := c.Query("toDate")
	result, err := service.GetAttendanceSummaryForDepartments(department, fromDate, toDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func ExportAttendanceSummaryCSV(c *gin.Context) {
	department := c.Query("department")
	fromDate := c.Query("fromDate")
	toDate := c.Query("toDate")
	csvData, err := service.GenerateAttendanceSummaryCSV(department, fromDate, toDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=summary.csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}

func ExportAttendanceSummaryPDF(c *gin.Context) {
	department := c.Query("department")
	fromDate := c.Query("fromDate")
	toDate := c.Query("toDate")
	pdfData, err := service.GenerateAttendanceSummaryPDF(department, fromDate, toDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=summary.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfData)
>>>>>>> architecture
}
