package main

import (
	"fmt"
	"log"
	"time"

	"github.com/4040www/NativeCloud_HR/config"
	"github.com/4040www/NativeCloud_HR/internal/api"
	"github.com/4040www/NativeCloud_HR/internal/db"
	messagequeue "github.com/4040www/NativeCloud_HR/internal/messageQueue"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化設定
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("設定加載失敗: %v", err)
	}

	// 初始化日誌
	// utils.InitLogger()

	// 連接資料庫
	db.InitPostgres()

	// message queue 相關
	// 初始化 Kafka
	brokers := "kafka:9092" // 你 Kafka 的 broker 位址
	if err := messagequeue.InitKafka(brokers); err != nil {
		log.Fatalf("failed to init kafka: %v", err)
	}
	if err := messagequeue.StartConsumer(brokers, "checkin-consumer-group"); err != nil {
		log.Fatalf("failed to start consumer: %v", err)
	}

	// 設置 API 路由
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000", "http://localhost:8080"}, // 修正這裡
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api.SetupRoutes(router)

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("伺服器啟動於 %s", serverAddr)

	// 改用 Gin 提供的啟動方式
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}
