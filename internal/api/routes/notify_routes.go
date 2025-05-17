package routes

import (
	handlers "github.com/4040www/NativeCloud_HR/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func RegisterNotifyRoutes(router *gin.RouterGroup) {
	// notify := router.Group("/")
	{
		router.GET("/warning", handlers.OvertimeOrLateCheck)
		router.POST("/late", handlers.NotifyManagerTooManyLate)
		router.POST("/overtime", handlers.NotifyHRExceedOvertime)
	}
}

// curl -X POST http://localhost:8080/api/notify/late \
//   -H "Content-Type: application/json" \
//   -d '{"employee_id": "d3549701-c2a2-4857-b0d1-c3c7b71aed3d"}'

// curl -X POST http://localhost:8080/api/notify/overtime \
//   -H "Content-Type: application/json" \
//   -d '{"employee_id": "d3549701-c2a2-4857-b0d1-c3c7b71aed3d"}'
