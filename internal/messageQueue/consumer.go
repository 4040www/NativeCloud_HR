package messagequeue

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

// 啟動 Kafka 消費者
func StartKafkaConsumer(brokers []string, topic string, groupID string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	log.Println("Kafka Consumer 啟動中...")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka 讀取失敗: %v", err)
			continue
		}

		log.Printf("Kafka 訊息接收: %s", string(msg.Value))

		// 這裡可以把訊息存進資料庫
	}
}

// NATS 版本

// package events

// import (
// 	"log"

// 	"github.com/nats-io/nats.go"
// )

// // 啟動 NATS 訂閱
// func StartNATSConsumer() {
// 	_, err := nc.Subscribe("access.events", func(msg *nats.Msg) {
// 		log.Printf("NATS 訊息接收: %s", string(msg.Data))

// 		// 這裡可以把訊息存進資料庫
// 	})

// 	if err != nil {
// 		log.Fatalf("NATS 訂閱失敗: %v", err)
// 	}

// 	log.Println("NATS Consumer 啟動完成")
// }
