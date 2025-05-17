package model

type CheckInRequest struct {
	ID        string `json:"access_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
	Time      string `json:"access_time" binding:"required"`
	Direction string `json:"direction" binding:"required"` // 進或出
	Gate_type string `json:"gate_type" binding:"required"`
	CheckinAt string `json:"gate_name" binding:"required"`
}
