package model

import "time"

// NotifyRecord 用來記錄系統發出的通知
type NotifyRecord struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Type       string    `json:"type"`        // 通知類型 (LateWarning, OvertimeAlert, etc.)
	Title      string    `json:"title"`       // 通知標題
	Content    string    `json:"content"`     // 通知內容
	ReceiverID uint      `json:"receiver_id"` // 收到通知的人（主管 / HR）
	EmployeeID uint      `json:"employee_id"` // 被通知相關的員工 ID
	IsRead     bool      `json:"is_read"`     // 是否已讀
	CreatedAt  time.Time `json:"created_at"`  // 發送時間
}
