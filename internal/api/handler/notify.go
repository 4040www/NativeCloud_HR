package handlers

import (
	"net/http"

	"github.com/4040www/NativeCloud_HR/internal/service"
	"github.com/gin-gonic/gin"
)

func OvertimeOrLateCheck(c *gin.Context) {
	problems, err := service.FindProblematicEmployees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢異常員工失敗"})
		return
	}
	if problems == nil {
		c.JSON(http.StatusOK, gin.H{"message": "本月無異常員工"})
		return
	}
	// 這裡可以根據需要進行進一步的處理，例如發送通知或記錄
	c.JSON(http.StatusOK, problems)
}

func NotifyManagerTooManyLate(c *gin.Context) {
	var req struct {
		EmployeeID string `json:"employee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	message := service.NotifyManagerLate(req.EmployeeID)
	c.JSON(http.StatusOK, gin.H{"message": message})
}

func NotifyHRExceedOvertime(c *gin.Context) {
	var req struct {
		EmployeeID string `json:"employee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	message := service.NotifyHROvertime(req.EmployeeID)
	c.JSON(http.StatusOK, gin.H{"message": message})
}
