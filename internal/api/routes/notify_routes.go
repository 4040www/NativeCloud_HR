package routes

import (
	handlers "github.com/4040www/NativeCloud_HR/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func RegisterNotifyRoutes(notify *gin.RouterGroup) {

	notify.GET("/warning", handlers.OvertimeOrLateCheck)      // GET /notify/warning
	notify.POST("/late", handlers.NotifyManagerTooManyLate)   // POST /notify/late
	notify.POST("/overtime", handlers.NotifyHRExceedOvertime) // POST /notify/overtime

}
