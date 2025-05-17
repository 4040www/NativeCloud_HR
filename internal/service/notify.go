package service

import (
	"fmt"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/repository"
)

type ProblemEmployee struct {
	EmployeeID string   `json:"employee_id"`
	Problems   []string `json:"problems"`
}

func FindProblematicEmployees() ([]ProblemEmployee, error) {
	employeeIDs, err := repository.GetAllEmployeeIDs()
	if err != nil {
		return nil, err
	}

	var result []ProblemEmployee
	now := time.Now()
	loc := now.Location()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	monthEnd := monthStart.AddDate(0, 1, 0)

	for _, eid := range employeeIDs {
		logs, err := repository.GetAccessLogsByEmployeeBetween(eid, monthStart, monthEnd)
		if err != nil {
			continue
		}

		// 按天分組
		dailyLogs := map[string][]time.Time{}
		for _, log := range logs {
			dateKey := log.AccessTime.Format("2006-01-02")
			dailyLogs[dateKey] = append(dailyLogs[dateKey], log.AccessTime)
		}

		lateCount := 0
		overtimeHours := 0.0

		for _, times := range dailyLogs {
			if len(times) < 2 {
				continue
			}
			// 找最早和最晚
			earliest, latest := times[0], times[0]
			for _, t := range times {
				if t.Before(earliest) {
					earliest = t
				}
				if t.After(latest) {
					latest = t
				}
			}

			// 判斷是否遲到
			threshold := time.Date(earliest.Year(), earliest.Month(), earliest.Day(), 8, 30, 0, 0, loc)
			if earliest.After(threshold) {
				lateCount++
			}

			// 計算加班時數
			workHours := latest.Sub(earliest).Hours()
			if workHours > 8 {
				overtimeHours += workHours - 8
			}
		}

		var problems []string
		if lateCount >= 4 {
			problems = append(problems, "TooManyLate")
		}
		if overtimeHours > 46 {
			problems = append(problems, "OvertimeExceeded")
		}

		if len(problems) > 0 {
			result = append(result, ProblemEmployee{
				EmployeeID: eid,
				Problems:   problems,
			})
		}
	}
	return result, nil
}
func NotifyManagerLate(employeeID string) string {
	loc := time.Now().Location()
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, loc)
	monthEnd := monthStart.AddDate(0, 1, 0)

	logs, err := repository.GetAccessLogsByEmployeeBetween(employeeID, monthStart, monthEnd)
	if err != nil {
		return "查詢失敗，請稍後再試。"
	}

	// 分組並找最早打卡
	dailyLogs := make(map[string][]time.Time)
	for _, log := range logs {
		if log.Direction != "IN" {
			continue
		}
		date := log.AccessTime.Format("2006-01-02")
		dailyLogs[date] = append(dailyLogs[date], log.AccessTime)
	}

	lateCount := 0
	for _, times := range dailyLogs {
		earliest := times[0]
		for _, t := range times {
			if t.Before(earliest) {
				earliest = t
			}
		}
		threshold := time.Date(earliest.Year(), earliest.Month(), earliest.Day(), 8, 30, 0, 0, loc)
		if earliest.After(threshold) {
			lateCount++
		}
	}

	if lateCount > 4 {
		return fmt.Sprintf("員工 %s 本月已遲到 %d 次，請主管關注。", employeeID, lateCount)
	}
	return fmt.Sprintf("員工 %s 本月遲到次數為 %d 次，尚無需警告。", employeeID, lateCount)
}

func NotifyHROvertime(employeeID string) string {
	loc := time.Now().Location()
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, loc)
	monthEnd := monthStart.AddDate(0, 1, 0)

	logs, err := repository.GetAccessLogsByEmployeeBetween(employeeID, monthStart, monthEnd)
	if err != nil {
		return "查詢失敗，請稍後再試。"
	}

	// 分組找最早與最晚
	dailyLogs := make(map[string][]time.Time)
	for _, log := range logs {
		date := log.AccessTime.Format("2006-01-02")
		dailyLogs[date] = append(dailyLogs[date], log.AccessTime)
	}

	overtime := 0.0
	for _, times := range dailyLogs {
		if len(times) < 2 {
			continue
		}
		earliest, latest := times[0], times[0]
		for _, t := range times {
			if t.Before(earliest) {
				earliest = t
			}
			if t.After(latest) {
				latest = t
			}
		}
		worked := latest.Sub(earliest).Hours()
		if worked > 8 {
			overtime += worked - 8
		}
	}

	if overtime > 46 {
		return fmt.Sprintf("員工 %s 本月加班時數為 %.1f 小時，請 HR 檢查。", employeeID, overtime)
	}
	return fmt.Sprintf("員工 %s 本月加班時數為 %.1f 小時，尚無需警告。", employeeID, overtime)
}

// func NotifyManagerLate(employeeID string) string {
// 	return "員工 " + employeeID + " 本月遲到超過 4 次，請主管關注。"
// }

// func NotifyHROvertime(employeeID string) string {
// 	return "員工 " + employeeID + " 本月加班超過 46 小時，請 HR 檢查。"
// }
