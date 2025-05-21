package repository

import (
	"time"

	"github.com/4040www/NativeCloud_HR/internal/db"
	"github.com/4040www/NativeCloud_HR/internal/model"
)

func GetAccessLogsByEmployeeBetween(employeeID string, start, end time.Time) ([]model.AccessLog, error) {
	var logs []model.AccessLog
	err := db.DB.Table("access_log").Where("employee_id = ? AND access_time BETWEEN ? AND ?", employeeID, start, end).Order("access_time asc").Find(&logs).Error
	return logs, err
}

func GetAllEmployees() ([]model.Employee, error) {
	var employees []model.Employee
	err := db.DB.Find(&employees).Error

	return employees, err
}

// 原本的 GetEmployeeByID 函數
func GetEmployeeByID(id string) (*model.Employee, error) {
	var emp model.Employee
	err := db.DB.Where("employee_id = ?", id).First(&emp).Error
	return &emp, err
}

// For unit test
// func GetEmployeeByID(db *gorm.DB, id string) (*model.Employee, error) {
// 	var emp model.Employee
// 	err := db.Where("employee_id = ?", id).First(&emp).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &emp, nil
// }

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
