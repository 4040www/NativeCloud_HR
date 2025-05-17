package routes

import (
	handlers "github.com/4040www/NativeCloud_HR/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(auth *gin.RouterGroup) {

	auth.POST("/login", handlers.LoginHandler)
	auth.POST("/logout", handlers.LogoutHandler)

}
