package repository

import (
	"time"

	"github.com/4040www/NativeCloud_HR/internal/db"
	"github.com/4040www/NativeCloud_HR/internal/model"
)

func GetAccessLogsByEmployeeAndMonth(employeeID string, month string) ([]model.AccessLog, error) {
	var logs []model.AccessLog
	startDate, _ := time.Parse("2006-01", month)
	endDate := startDate.AddDate(0, 1, 0)

	err := db.DB.Table("access_log").Where("employee_id = ? AND access_time BETWEEN ? AND ?", employeeID, startDate, endDate).Find(&logs).Error
	return logs, err
}

func GetAllEmployeeIDs() ([]string, error) {
	var ids []string
	err := db.DB.Model(&model.Employee{}).Pluck("employee_id", &ids).Error
	return ids, err
}
