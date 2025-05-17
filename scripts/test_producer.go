package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
)

type CheckInRequest struct {
	ID        string `json:"access_id"`
	UserID    string `json:"user_id"`
	Time      string `json:"access_time"`
	Direction string `json:"direction"`
	GateType  string `json:"gate_type"`
	GateName  string `json:"gate_name"`
}

func main() {
	// Kafka 設定
	brokers := []string{"kafka:9092"} // 你的 broker 地址
	topic := "checkin-topic"

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	// 隨機種子
	rand.Seed(time.Now().UnixNano())

	// 要送多少筆
	total := 4000

	for i := 0; i < total; i++ {
		checkIn := generateRandomCheckIn()

		// 轉 JSON
		jsonData, err := json.Marshal(checkIn)
		if err != nil {
			fmt.Printf("failed to marshal record %d: %v\n", i, err)
			continue
		}

		// 送出訊息
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(jsonData),
		}

		_, _, err = producer.SendMessage(msg)
		if err != nil {
			fmt.Printf("failed to send message %d: %v\n", i, err)
		} else {
			fmt.Printf("Sent record %d\n", i)
		}

		// 可以加一點 delay 避免瞬間打爆
		time.Sleep(1 * time.Millisecond)
	}

	fmt.Println("All messages sent!")
}

func generateRandomCheckIn() CheckInRequest {
	directions := []string{"IN", "OUT"}
	gateTypes := []string{"大門", "電梯", "側門"}
	gateNames := []string{"北門", "南門", "東門", "西門"}

	return CheckInRequest{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Time:      time.Now().Add(time.Duration(rand.Intn(3600)) * time.Second).Format(time.RFC3339),
		Direction: directions[rand.Intn(len(directions))],
		GateType:  gateTypes[rand.Intn(len(gateTypes))],
		GateName:  gateNames[rand.Intn(len(gateNames))],
	}
}
