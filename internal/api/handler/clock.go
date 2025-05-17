package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"

	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/4040www/NativeCloud_HR/internal/service"
)

func CheckIn(c *gin.Context) {
	var req model.CheckInRequest

	// 解析使用者傳的 JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	// 從 JWT context 中取出 user_id
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(401, gin.H{"error": "User ID not found in token"})
		return
	}

	// 產生 UUID 當作 access_id
	req.ID = uuid.New().String()
	req.UserID = userID // 強制用 token 裡的，避免前端亂傳

	// 呼叫 service
	if err := service.Clock(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Check-in successful",
	})
}
