// report_service.go
package service

import (
	"bytes"
	"encoding/csv"
	"sort"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/4040www/NativeCloud_HR/internal/repository"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// Get simple employee's attendance summary
<<<<<<< HEAD
func GetTodayAttendanceSummary(db *gorm.DB, userID string) (*model.AttendanceSummary, error) {
	logs, err := FetchTodayRecords(db, userID)
=======
func GetTodayAttendanceSummary(userID string) (*model.AttendanceSummary, error) {
	logs, err := FetchTodayRecords(userID)
>>>>>>> architecture
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, nil
	}

<<<<<<< HEAD
	emp, err := repository.GetEmployeeByID(db, userID) // Modified to use the DB instance for unit tests
=======
	emp, err := repository.GetEmployeeByID(userID)
>>>>>>> architecture
	if err != nil {
		return nil, err
	}

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

	if clockIn == nil && clockOut == nil {
		status = "Day Off"
	}

	var date string
	if clockIn != nil {
		date = clockIn.AccessTime.Format("2006-01-02")
	} else if clockOut != nil {
		date = clockOut.AccessTime.Format("2006-01-02")
	} else {
		date = time.Now().Format("2006-01-02")
	}

	summary := &model.AttendanceSummary{
		Date:         date,
		Name:         emp.FirstName + " " + emp.LastName,
		ClockInTime:  formatTime(clockIn),
		ClockOutTime: formatTime(clockOut),
		ClockInGate:  getGate(clockIn),
		ClockOutGate: getGate(clockOut),
		Status:       status,
	}

	return summary, nil

}

<<<<<<< HEAD
func GetAttendanceWithEmployee(db *gorm.DB, userID string, start, end time.Time) ([]model.AttendanceSummary, error) {
	records, err := repository.GetAccessLogsByEmployeeBetween(db, userID, start, end.Add(24*time.Hour))
=======
func GetAttendanceWithEmployee(userID string, start, end time.Time) ([]model.AttendanceSummary, error) {
	records, err := repository.GetAccessLogsByEmployeeBetween(userID, start, end.Add(24*time.Hour))
>>>>>>> architecture
	if err != nil {
		return nil, err
	}

<<<<<<< HEAD
	emp, err := repository.GetEmployeeByID(db, userID) // Modified to use the DB instance for unit tests
=======
	emp, err := repository.GetEmployeeByID(userID)
>>>>>>> architecture
	if err != nil {
		return nil, err
	}

	dateMap := make(map[string][]model.AccessLog)
	for _, r := range records {
		day := r.AccessTime.Format("2006-01-02")
		dateMap[day] = append(dateMap[day], r)
	}

	var results []model.AttendanceSummary
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

		if clockIn == nil && clockOut == nil {
			status = "Day Off"
		}

		results = append(results, model.AttendanceSummary{
			Date:         date,
			Name:         emp.FirstName + " " + emp.LastName,
			ClockInTime:  formatTime(clockIn),
			ClockOutTime: formatTime(clockOut),
			ClockInGate:  getGate(clockIn),
			ClockOutGate: getGate(clockOut),
			Status:       status,
		})
	}

	return results, nil

}

<<<<<<< HEAD
func FetchTodayRecords(db *gorm.DB, employeeID string) ([]model.AccessLog, error) {
=======
func FetchTodayRecords(employeeID string) ([]model.AccessLog, error) {
>>>>>>> architecture
	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start.Add(24 * time.Hour)
	return repository.GetAccessLogsByEmployeeBetween(db, employeeID, start, end)
}

func FetchHistoryRecords(db *gorm.DB, employeeID string) ([]model.AccessLog, error) {
	start := time.Now().AddDate(0, -1, 0)
	end := time.Now()
	return repository.GetAccessLogsByEmployeeBetween(db, employeeID, start, end)
}

func FetchHistoryRecordsBetween(db *gorm.DB, employeeID, startDate, endDate string) ([]model.AccessLog, error) {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	return repository.GetAccessLogsByEmployeeBetween(db, employeeID, start, end.Add(24*time.Hour))
}

func FetchMonthComparisonReport(db *gorm.DB, departmentID, month string) (map[string]interface{}, map[string]interface{}, error) {
	current, err := FetchMonthlyTeamReport(db, departmentID, month)
	if err != nil {
		return nil, nil, err
	}
	timeObj, _ := time.Parse("2006-01", month)
	prevMonth := timeObj.AddDate(0, -1, 0).Format("2006-01")
	prev, err := FetchMonthlyTeamReport(db, departmentID, prevMonth)
	return current, prev, err
}

func FetchMonthlyTeamReport(db *gorm.DB, departmentID, month string) (map[string]interface{}, error) {
	employees, err := repository.GetAllEmployees(db)
	if err != nil {
		return nil, err
	}

	loc := time.Now().Location()
	monthTime, _ := time.ParseInLocation("2006-01", month, loc)
	start := time.Date(monthTime.Year(), monthTime.Month(), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)

	// fmt.Printf("🔍 Report period: %s ~ %s\n", start, end)

	totalHours, otHours := 0.0, 0.0
	overtimeCount := 0
	uniqueEmployees := make(map[string]bool)

	for _, e := range employees {
		if e.OrganizationID == departmentID {
			// fmt.Printf("👤 %s %s (%s)\n", e.FirstName, e.LastName, e.EmployeeID)
			logs, _ := repository.GetAccessLogsByEmployeeBetween(db, e.EmployeeID, start, end)
			// fmt.Printf("   ⏰ %d access logs\n", len(logs))
			workHours, _ := calculateDailyWorkHours(logs)
			// fmt.Printf("   📊 Work hours: %.2f\n", workHours)

			totalHours += workHours
			if workHours > 8 {
				otHours += workHours - 8
				overtimeCount++
			}
			uniqueEmployees[e.EmployeeID] = true
		}
	}

	// fmt.Println("✅ Done:", totalHours, otHours, overtimeCount, len(uniqueEmployees))

	return map[string]interface{}{
		"TotalWorkHours": totalHours,
		"TotalOTHours":   otHours,
		"OTHoursPerson":  overtimeCount,
		"OTHeadcounts":   len(uniqueEmployees),
	}, nil
}

func FetchWeeklyTeamReport(db *gorm.DB, departmentID, startDate, endDate string) (map[string]interface{}, error) {
	return FetchCustomPeriodTeamReport(db, departmentID, startDate, endDate)
}

func FetchCustomPeriodTeamReport(db *gorm.DB, departmentID, startDate, endDate string) (map[string]interface{}, error) {
	// fmt.Println("⚙️ FetchCustomPeriodTeamReport called with:", departmentID, startDate, endDate)

	employees, err := repository.GetAllEmployees(db)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Employee list: ", employees)
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	// fmt.Println("time: ", start, end)
	totalHours, otHours := 0.0, 0.0
	overtimeCount := 0
	uniqueEmployees := make(map[string]bool)

	for _, e := range employees {
		// fmt.Printf("👤 %s %s (%s)\n", e.FirstName, e.LastName, e.OrganizationID, departmentID)
		if e.OrganizationID == departmentID {
			// fmt.Printf("⏰ %s %s (%s)\n", e.FirstName, e.LastName, e.EmployeeID)
			logs, _ := repository.GetAccessLogsByEmployeeBetween(db, e.EmployeeID, start, end.Add(24*time.Hour))
			// fmt.Println("logs", logs)
			workHours, _ := calculateDailyWorkHours(logs)

			// fmt.Printf("👤 %s logs: %d, workHours: %.2f\n", e.EmployeeID, len(logs), workHours)

			totalHours += workHours
			if workHours > 8 {
				otHours += workHours - 8
				overtimeCount++
			}
			uniqueEmployees[e.EmployeeID] = true
		}
	}

	// fmt.Println("✅ Done:", totalHours, otHours, overtimeCount, len(uniqueEmployees))

	return map[string]interface{}{
		"TotalWorkHours": totalHours,
		"TotalOTHours":   otHours,
		"OTHoursPerson":  overtimeCount,
		"OTHeadcounts":   len(uniqueEmployees),
	}, nil
}

func GenerateAlertList(db *gorm.DB, startDate, endDate string) ([]map[string]interface{}, error) {
	employees, err := repository.GetAllEmployees(db)
	if err != nil {
		return nil, err
	}
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var alerts []map[string]interface{}

	for _, e := range employees {
		logs, _ := repository.GetAccessLogsByEmployeeBetween(db, e.EmployeeID, start, end.Add(24*time.Hour))
		// fmt.Println("Employee:", e.EmployeeID)
		// fmt.Println("Logs:", logs)
		// 將紀錄按日期分類
		dayMap := make(map[string][]model.AccessLog)
		for _, log := range logs {
			dateStr := log.AccessTime.Format("2006-01-02")
			dayMap[dateStr] = append(dayMap[dateStr], log)
		}
		// fmt.Println("daymap:", dayMap)
		otCount := 0
		otHours := 0.0
		warningDays := 0
		alertDays := 0

		for _, dayLogs := range dayMap {
			// fmt.Println("daylog:", dayLogs)
			workHours, _ := calculateDailyWorkHours(dayLogs)
			// fmt.Println("Work hours:", workHours)
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

			// 只加入 Warning 或 Alert
			if status == "Warning" || status == "Alert" {
				alerts = append(alerts, map[string]interface{}{
					"EmployeeID": e.EmployeeID,
					"Name":       e.FirstName + " " + e.LastName,
					"OTCounts":   otCount,
					"OTHours":    otHours,
					"status":     status,
				})
			}
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
	// fmt.Println("GetManagedDepartments called with userID:", userID)
	depts, err := GetManagedDepartmentsFromDB(userID)
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

func GetAttendanceSummaryForDepartments(db *gorm.DB, department, startDate, endDate string) ([]map[string]interface{}, error) {
	employees, err := repository.GetEmployeesByDepartment(department)
	if err != nil {
		return nil, err
	}

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	// 建立從 start 到 end 的每一天清單
	dates := []string{}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("2006-01-02"))
	}

	var result []map[string]interface{}
	for _, emp := range employees {
<<<<<<< HEAD
		logs, _ := repository.GetAccessLogsByEmployeeBetween(db, emp.EmployeeID, start, end.Add(24*time.Hour))
=======
		logs, _ := repository.GetAccessLogsByEmployeeBetween(emp.EmployeeID, start, end.Add(24*time.Hour))

		// 將 logs 根據日期分組
>>>>>>> architecture
		dateMap := make(map[string][]model.AccessLog)
		for _, r := range logs {
			day := r.AccessTime.Format("2006-01-02")
			dateMap[day] = append(dateMap[day], r)
		}

		for _, date := range dates {
			logs := dateMap[date]
			var clockIn, clockOut *model.AccessLog
			status := "On Time"

			if len(logs) == 0 {
				status = "Day Off"
			} else {
				for _, log := range logs {
					if log.Direction == "IN" && (clockIn == nil || log.AccessTime.Before(clockIn.AccessTime)) {
						clockIn = &log
					}
					if log.Direction == "OUT" && (clockOut == nil || log.AccessTime.After(clockOut.AccessTime)) {
						clockOut = &log
					}
				}

				if clockIn == nil || clockOut == nil {
					status = "Abnormal"
				} else if clockIn.AccessTime.Hour() > 8 || (clockIn.AccessTime.Hour() == 8 && clockIn.AccessTime.Minute() > 30) {
					status = "Late"
				}
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

	// 排序：日期從新到舊
	sort.Slice(result, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", result[i]["date"].(string))
		dateJ, _ := time.Parse("2006-01-02", result[j]["date"].(string))
		return dateI.After(dateJ)
	})

	return result, nil
}

/*
Status要有：
1. On Time --> 8:30 前到
2. Late -->
3. Leave Early --> 17:30 前走的就算是
4. Day Off（週末）-->
5. Abnormal（沒有打卡紀錄）--> 一天有打一次卡（只有一次進或是只有一次出）
*/

func GenerateAttendanceSummaryCSV(db *gorm.DB, dept, start, end string) ([]byte, error) {
	summary, err := GetAttendanceSummaryForDepartments(db, dept, start, end)
	if err != nil {
		return nil, err
	}

	sort.Slice(summary, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", summary[i]["date"].(string))
		dateJ, _ := time.Parse("2006-01-02", summary[j]["date"].(string))
		return dateI.After(dateJ)
	})

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

func GenerateAttendanceSummaryPDF(db *gorm.DB, dept, start, end string) ([]byte, error) {
	summary, err := GetAttendanceSummaryForDepartments(db, dept, start, end)
	if err != nil {
		return nil, err
	}

	// ✅ 排序：從新到舊
	sort.Slice(summary, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", summary[i]["date"].(string))
		dateJ, _ := time.Parse("2006-01-02", summary[j]["date"].(string))
		return dateI.After(dateJ)
	})

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// ✅ 標題
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Attendance Summary")
	pdf.Ln(12)

	// ✅ 表頭
	header := []string{"Date", "Employee ID", "Name", "Clock-in Time", "Clock-in Gate", "Clock-out Time", "Clock-out Gate", "Status"}
	pdf.SetFont("Arial", "B", 10)
	for _, str := range header {
		pdf.CellFormat(25, 8, str, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// ✅ 表格內容
	pdf.SetFont("Arial", "", 9)
	for _, row := range summary {
		pdf.CellFormat(25, 7, row["date"].(string), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, row["employeeID"].(string), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, row["name"].(string), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, row["ClockInTime"].(string), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, row["ClockInGate"].(string), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, row["ClockOutTime"].(string), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, row["ClockOutGate"].(string), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, row["status"].(string), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
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
