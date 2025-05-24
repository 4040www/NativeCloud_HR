package repository

import (
	"errors"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/db"
	"github.com/4040www/NativeCloud_HR/internal/model"
	"gorm.io/gorm"
)

func GetAccessLogsByEmployeeBetween(db *gorm.DB, employeeID string, start, end time.Time) ([]model.AccessLog, error) {
	var logs []model.AccessLog
	err := db.Table("access_log").Where("employee_id = ? AND access_time BETWEEN ? AND ?", employeeID, start, end).Order("access_time asc").Find(&logs).Error
	return logs, err
}

func GetAllEmployees(db *gorm.DB) ([]model.Employee, error) {
	var employees []model.Employee
	err := db.Find(&employees).Error

	return employees, err
}

// 原本的 GetEmployeeByID 函數
// func GetEmployeeByID(id string) (*model.Employee, error) {
// 	var emp model.Employee
// 	err := db.DB.Where("employee_id = ?", id).First(&emp).Error
// 	return &emp, err
// }

// For unit test
func GetEmployeeByID(db *gorm.DB, id string) (*model.Employee, error) {
	var emp model.Employee
	if err := db.Where("employee_id = ?", id).First(&emp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 查無資料但不是錯誤
		}
		return nil, err // 真正錯誤
	}
	return &emp, nil
}

// Page: Attendance Page
func GetDepartmentsByManager(userID string) ([]string, error) {
	// 假設使用者和部門間有關聯
	var departments []string
	err := db.DB.Table("employee").Select("distinct organization_id").Where("employee_id = ?", userID).Scan(&departments).Error
	return departments, err
}

func GetEmployeesByDepartment(department string) ([]model.Employee, error) {
	var emps []model.Employee
	err := db.DB.Where("organization_id = ?", department).Find(&emps).Error
	return emps, err
}
