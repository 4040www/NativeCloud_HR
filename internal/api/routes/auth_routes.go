package routes

import (
	handlers "github.com/4040www/NativeCloud_HR/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(auth *gin.RouterGroup) {
	auth.OPTIONS("/login", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.AbortWithStatus(204)
	})

	auth.POST("/login", handlers.LoginHandler)
	auth.POST("/logout", handlers.LogoutHandler)

}

// curl -X GET http://localhost:8080/api/auth/login \
// -H "Content-Type: application/json" \
// -d '{
// "email": "test@example.com",
// "password": "your_password"}'
