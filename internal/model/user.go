package model

type Employee struct {
	EmployeeID     string `gorm:"primaryKey"`
	FirstName      string
	LastName       string
	IsManager      bool
	Password       string
	Email          string
	OrganizationID string
}

// 顯式指定 GORM 使用 "Employee" 這個資料表
func (Employee) TableName() string {
	return "employee"
}
