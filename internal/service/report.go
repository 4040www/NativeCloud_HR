// report_service.go
package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/4040www/NativeCloud_HR/internal/repository"
	"github.com/jung-kurt/gofpdf"
)

func FetchTodayRecords(employeeID string) ([]model.AccessLog, error) {
	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start.Add(24 * time.Hour)
	return repository.GetAccessLogsByEmployeeBetween(employeeID, start, end)
}

func FetchHistoryRecords(employeeID string) ([]model.AccessLog, error) {
	start := time.Now().AddDate(0, -1, 0)
	end := time.Now()
	return repository.GetAccessLogsByEmployeeBetween(employeeID, start, end)
}

func FetchHistoryRecordsBetween(employeeID, startDate, endDate string) ([]model.AccessLog, error) {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	return repository.GetAccessLogsByEmployeeBetween(employeeID, start, end.Add(24*time.Hour))
}

func FetchMonthComparisonReport(departmentID, month string) (map[string]interface{}, map[string]interface{}, error) {
	current, err := FetchMonthlyTeamReport(departmentID, month)
	if err != nil {
		return nil, nil, err
	}
	timeObj, _ := time.Parse("2006-01", month)
	prevMonth := timeObj.AddDate(0, -1, 0).Format("2006-01")
	prev, err := FetchMonthlyTeamReport(departmentID, prevMonth)
	return current, prev, err
}

func FetchMonthlyTeamReport(departmentID, month string) (map[string]interface{}, error) {
	employees, err := repository.GetAllEmployees()
	if err != nil {
		return nil, err
	}

	loc := time.Now().Location()
	monthTime, _ := time.ParseInLocation("2006-01", month, loc)
	start := time.Date(monthTime.Year(), monthTime.Month(), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)

	fmt.Printf("🔍 Report period: %s ~ %s\n", start, end)

	totalHours, otHours := 0.0, 0.0
	overtimeCount := 0
	uniqueEmployees := make(map[string]bool)

	for _, e := range employees {
		if e.OrganizationID == departmentID {
			fmt.Printf("👤 %s %s (%s)\n", e.FirstName, e.LastName, e.EmployeeID)
			logs, _ := repository.GetAccessLogsByEmployeeBetween(e.EmployeeID, start, end)
			fmt.Printf("   ⏰ %d access logs\n", len(logs))
			workHours, _ := calculateDailyWorkHours(logs)
			fmt.Printf("   📊 Work hours: %.2f\n", workHours)

			totalHours += workHours
			if workHours > 8 {
				otHours += workHours - 8
				overtimeCount++
			}
			uniqueEmployees[e.EmployeeID] = true
		}
	}

	fmt.Println("✅ Done:", totalHours, otHours, overtimeCount, len(uniqueEmployees))

	return map[string]interface{}{
		"TotalWorkHours": totalHours,
		"TotalOTHours":   otHours,
		"OTHoursPerson":  overtimeCount,
		"OTHeadcounts":   len(uniqueEmployees),
	}, nil
}

func FetchWeeklyTeamReport(departmentID, startDate, endDate string) (map[string]interface{}, error) {
	return FetchCustomPeriodTeamReport(departmentID, startDate, endDate)
}

func FetchCustomPeriodTeamReport(departmentID, startDate, endDate string) (map[string]interface{}, error) {
	fmt.Println("⚙️ FetchCustomPeriodTeamReport called with:", departmentID, startDate, endDate)

	employees, err := repository.GetAllEmployees()
	if err != nil {
		return nil, err
	}
	// fmt.Println("Employee list: ", employees)
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	totalHours, otHours := 0.0, 0.0
	overtimeCount := 0
	uniqueEmployees := make(map[string]bool)

	for _, e := range employees {
		// fmt.Printf("👤 %s %s (%s)\n", e.FirstName, e.LastName, e.OrganizationID, departmentID)
		if e.OrganizationID == departmentID {
			fmt.Printf("⏰ %s %s (%s)\n", e.FirstName, e.LastName, e.EmployeeID)
			logs, _ := repository.GetAccessLogsByEmployeeBetween(e.EmployeeID, start, end.Add(24*time.Hour))
			workHours, _ := calculateDailyWorkHours(logs)

			fmt.Printf("👤 %s logs: %d, workHours: %.2f\n", e.EmployeeID, len(logs), workHours)

			totalHours += workHours
			if workHours > 8 {
				otHours += workHours - 8
				overtimeCount++
			}
			uniqueEmployees[e.EmployeeID] = true
		}
	}

	fmt.Println("✅ Done:", totalHours, otHours, overtimeCount, len(uniqueEmployees))

	return map[string]interface{}{
		"TotalWorkHours": totalHours,
		"TotalOTHours":   otHours,
		"OTHoursPerson":  overtimeCount,
		"OTHeadcounts":   len(uniqueEmployees),
	}, nil
}

func GenerateAlertList(startDate, endDate string) ([]map[string]interface{}, error) {
	employees, err := repository.GetAllEmployees()
	if err != nil {
		return nil, err
	}
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var alerts []map[string]interface{}

	for _, e := range employees {
		logs, _ := repository.GetAccessLogsByEmployeeBetween(e.EmployeeID, start, end.Add(24*time.Hour))

		// 將紀錄按日期分類
		dayMap := make(map[string][]model.AccessLog)
		for _, log := range logs {
			dateStr := log.AccessTime.Format("2006-01-02")
			dayMap[dateStr] = append(dayMap[dateStr], log)
		}

		otCount := 0
		otHours := 0.0
		warningDays := 0
		alertDays := 0

		for _, dayLogs := range dayMap {
			workHours, _ := calculateDailyWorkHours(dayLogs)
			if workHours > 8 {
				otCount++
				otHours += workHours - 8
				if workHours >= 12 {
					alertDays++
				} else if workHours >= 10 {
					warningDays++
				}
			}
		}

		if otCount > 0 {
			status := "Normal"
			if alertDays >= 1 {
				status = "Alert"
			} else if warningDays >= 2 {
				status = "Warning"
			}

			alerts = append(alerts, map[string]interface{}{
				"EmployeeID": e.EmployeeID,
				"Name":       e.FirstName + " " + e.LastName,
				"OTCounts":   otCount,
				"OTHours":    otHours,
				"status":     status,
			})
		}
	}

	return alerts, nil
}

//	func GetManagedDepartments(userID string) []string {
//		// Mock 資料，實際應查角色或 DB 權限
//		if userID == "admin" {
//			return []string{"HR", "Sales", "Engineering"}
//		}
//		return []string{"Sales"}
//	}
func GetManagedDepartments(userID string) []string {
	depts, err := repository.GetManagedDepartmentsFromDB(userID)
	if err != nil || len(depts) == 0 {
		// fallback（或回傳空陣列）
		return []string{}
	}
	return depts
}

func calculateDailyWorkHours(logs []model.AccessLog) (float64, bool) {
	var clockIn, clockOut *time.Time
	isLate := false
	for _, log := range logs {
		if log.Direction == "IN" {
			if clockIn == nil || log.AccessTime.Before(*clockIn) {
				clockIn = &log.AccessTime
			}
			if log.AccessTime.Hour() > 8 || (log.AccessTime.Hour() == 8 && log.AccessTime.Minute() > 30) {
				isLate = true
			}
		}
		if log.Direction == "OUT" {
			if clockOut == nil || log.AccessTime.After(*clockOut) {
				clockOut = &log.AccessTime
			}
		}
	}
	if clockIn != nil && clockOut != nil {
		return clockOut.Sub(*clockIn).Hours(), isLate
	}
	return 0, isLate
}

func GetManagedDepartmentsFromDB(userID string) ([]string, error) {
	return repository.GetDepartmentsByManager(userID)
}

func GetAttendanceSummaryForDepartments(department, startDate, endDate string) ([]map[string]interface{}, error) {
	employees, err := repository.GetEmployeesByDepartment(department)
	if err != nil {
		return nil, err
	}
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var result []map[string]interface{}
	for _, emp := range employees {
		logs, _ := repository.GetAccessLogsByEmployeeBetween(emp.EmployeeID, start, end.Add(24*time.Hour))
		dateMap := make(map[string][]model.AccessLog)
		for _, r := range logs {
			day := r.AccessTime.Format("2006-01-02")
			dateMap[day] = append(dateMap[day], r)
		}
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
			result = append(result, map[string]interface{}{
				"date":         date,
				"employeeID":   emp.EmployeeID,
				"name":         emp.FirstName + " " + emp.LastName,
				"ClockInTime":  formatTime(clockIn),
				"ClockOutTime": formatTime(clockOut),
				"ClockInGate":  getGate(clockIn),
				"ClockOutGate": getGate(clockOut),
				"status":       status,
			})
		}
	}
	return result, nil
}

func GenerateAttendanceSummaryCSV(dept, start, end string) ([]byte, error) {
	summary, err := GetAttendanceSummaryForDepartments(dept, start, end)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	w.Write([]string{"date", "employee ID", "name", "clock-in time", "clock-in gate", "clock-out time", "clock-out gate", "status"})
	for _, row := range summary {
		w.Write([]string{
			row["date"].(string),
			row["employeeID"].(string),
			row["name"].(string),
			row["ClockInTime"].(string),
			row["ClockInGate"].(string),
			row["ClockOutTime"].(string),
			row["ClockOutGate"].(string),
			row["status"].(string),
		})
	}
	w.Flush()
	return b.Bytes(), nil
}

func GenerateAttendanceSummaryPDF(dept, start, end string) ([]byte, error) {
	summary, err := GetAttendanceSummaryForDepartments(dept, start, end)
	if err != nil {
		return nil, err
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Attendance Summary")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	for _, row := range summary {
		line := row["date"].(string) + " " + row["employeeID"].(string) + " " + row["name"].(string) + " " + row["ClockInTime"].(string) + " " + row["ClockOutTime"].(string) + " " + row["status"].(string)
		pdf.Cell(0, 10, line)
		pdf.Ln(6)
	}
	var b bytes.Buffer
	err = pdf.Output(&b)
	return b.Bytes(), err
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
