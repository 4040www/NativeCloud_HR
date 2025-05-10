// report_handlers.go
package handlers

import (
	"net/http"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/4040www/NativeCloud_HR/internal/repository"
	"github.com/4040www/NativeCloud_HR/internal/service"
	"github.com/gin-gonic/gin"
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

func GetMyTodayRecords(c *gin.Context) {
	userID := c.Param("userID")

	logs, err := service.FetchTodayRecords(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(logs) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no access records today"})
		return
	}

	emp, err := repository.GetEmployeeByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "employee not found"})
		return
	}

	var clockIn, clockOut *model.AccessLog
	isLate := false

	for _, log := range logs {
		if log.Direction == "IN" && (clockIn == nil || log.AccessTime.Before(clockIn.AccessTime)) {
			clockIn = &log
			if log.AccessTime.Hour() > 8 || (log.AccessTime.Hour() == 8 && log.AccessTime.Minute() > 30) {
				isLate = true
			}
		}
		if log.Direction == "OUT" && (clockOut == nil || log.AccessTime.After(clockOut.AccessTime)) {
			clockOut = &log
		}
	}

	resp := gin.H{
		"date":           clockIn.AccessTime.Format("2006-01-02"),
		"name":           emp.FirstName + " " + emp.LastName,
		"clock_in_time":  clockIn.AccessTime.Format("15:04"),
		"clock_out_time": clockOut.AccessTime.Format("15:04"),
		"clock_in_gate":  clockIn.GateName,
		"clock_out_gate": clockOut.GateName,
		"status":         "Normal",
	}
	if isLate {
		resp["status"] = "Late"
	}

	c.JSON(http.StatusOK, resp)
}

func GetMyHistoryRecords(c *gin.Context) {
	userID := c.Param("userID")

	// 自動補上過去 30 天
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	startDate := start.Format("2006-01-02")
	endDate := end.Format("2006-01-02")

	logs, err := service.FetchHistoryRecordsBetween(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func GetMyPeriodRecords(c *gin.Context) {
	userID := c.Param("userID")
	startDate := c.Param("startDate")
	endDate := c.Param("endDate")

	records, err := service.FetchHistoryRecordsBetween(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	emp, err := repository.GetEmployeeByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "employee not found"})
		return
	}

	dateMap := make(map[string][]model.AccessLog)
	for _, r := range records {
		day := r.AccessTime.Format("2006-01-02")
		dateMap[day] = append(dateMap[day], r)
	}

	var results []AttendanceSummary
	for date, logs := range dateMap {
		var clockIn, clockOut *model.AccessLog
		status := "On Time"

		for _, log := range logs {
			if log.Direction == "IN" && (clockIn == nil || log.AccessTime.Before(clockIn.AccessTime)) {
				clockIn = &log
				if log.AccessTime.Hour() > 8 || (log.AccessTime.Hour() == 8 && log.AccessTime.Minute() > 30) {
					status = "Late"
				}
			}
			if log.Direction == "OUT" && (clockOut == nil || log.AccessTime.After(clockOut.AccessTime)) {
				clockOut = &log
			}
		}

		if clockIn == nil || clockOut == nil {
			status = "Abnormal"
		}

		results = append(results, AttendanceSummary{
			Date:         date,
			Name:         emp.FirstName + " " + emp.LastName,
			ClockInTime:  formatTime(clockIn),
			ClockOutTime: formatTime(clockOut),
			ClockInGate:  getGate(clockIn),
			ClockOutGate: getGate(clockOut),
			Status:       status,
		})
	}

	c.JSON(http.StatusOK, results)
}

func formatTime(log *model.AccessLog) string {
	if log == nil {
		return ""
	}
	return log.AccessTime.Format("15:04")
}

func getGate(log *model.AccessLog) string {
	if log == nil {
		return ""
	}
	return log.GateName
}

func GetThisMonthTeam(c *gin.Context) {
	department := c.Param("department")
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	current, prev, err := service.FetchMonthComparisonReport(department, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, []interface{}{current, prev})
}

func GetThisWeekTeam(c *gin.Context) {
	department := c.Param("department")
	now := time.Now()
	start := now.AddDate(0, 0, -int(now.Weekday())+1)
	end := start.AddDate(0, 0, 6)
	lastStart := start.AddDate(0, 0, -7)
	lastEnd := end.AddDate(0, 0, -7)

	current, err1 := service.FetchWeeklyTeamReport(department, start.Format("2006-01-02"), end.Format("2006-01-02"))
	prev, err2 := service.FetchWeeklyTeamReport(department, lastStart.Format("2006-01-02"), lastEnd.Format("2006-01-02"))
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load report"})
		return
	}
	c.JSON(http.StatusOK, []interface{}{current, prev})
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
}

func GetInChargeDepartments(c *gin.Context) {
	userID := c.Param("userID")
	depts := service.GetManagedDepartments(userID)
	c.JSON(http.StatusOK, depts)
}

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
}
