package main

import (
	"fmt"
	"log"
	"net/http"

	"NativeCloud_HR/config"
	"NativeCloud_HR/internal/api"
	"NativeCloud_HR/internal/db"
	"NativeCloud_HR/internal/events"
	"NativeCloud_HR/internal/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化設定
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("設定加載失敗: %v", err)
	}

	// 初始化日誌
	utils.InitLogger()

	// 連接資料庫
	db.InitPostgres(cfg.DatabaseURL)
	db.InitRedis(cfg.RedisURL)

	// 初始化 Kafka
	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "access_logs"
	events.InitKafkaProducer(kafkaBrokers, kafkaTopic)
	go events.StartKafkaConsumer(kafkaBrokers, kafkaTopic, "access_group")

	// // 初始化 NATS
	// natsURL := "nats://localhost:4222"
	// events.InitNATS(natsURL)
	// go events.StartNATSConsumer()

	// 設置 API 路由
	router := gin.Default()
	api.SetupRoutes(router)

	// 啟動 HTTP 伺服器
	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("伺服器啟動於 %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}
