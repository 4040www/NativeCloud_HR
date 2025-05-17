package api

import (
	"time"

	"github.com/4040www/NativeCloud_HR/internal/api/routes"
	"github.com/gin-gonic/gin"
)

// modularized api router
func SetupRoutes(r *gin.Engine) {
	apiGroup := r.Group("/api")

	apiGroup.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "online",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	routes.RegisterAuthRoutes(apiGroup.Group("/auth"))
	routes.RegisterClockRoutes(apiGroup.Group("/clock"))

	routes.RegisterNotifyRoutes(apiGroup.Group("/notify"))
	routes.RegisterReportRoutes(apiGroup.Group("/report"))

	protected := apiGroup.Group("/")
	protected.Use(JWTMiddleware())
	// routes.RegisterNotifyRoutes(protected.Group("/notify"))
	// routes.RegisterReportRoutes(protected.Group("/report"))
}
