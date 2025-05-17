// model/report_model.go
package model

import "time"

type AccessLog struct {
	AccessID     string `gorm:"primaryKey"`
	EmployeeID   string
	AccessTime   time.Time
	Direction    string
	GateType     string
	GateName     string
	AccessResult string
}

type MyRecordResponse struct {
	Date         string `json:"date"`
	Name         string `json:"name"`
	ClockInTime  string `json:"clock_in_time"`
	ClockOutTime string `json:"clock_out_time"`
	ClockInGate  string `json:"clock_in_gate"`
	ClockOutGate string `json:"clock_out_gate"`
	Status       string `json:"status"`
}

// type Employees struct {
// 	EmployeeID     string `gorm:"primaryKey"`
// 	FirstName      string
// 	LastName       string
// 	IsManager      bool
// 	Password       string
// 	Email          string
// 	OrganizationID string
// }
