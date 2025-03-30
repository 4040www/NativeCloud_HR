package events

import (
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// 初始化 Kafka Producer
func InitKafkaProducer(brokers []string, topic string) {
	writer = &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}

	log.Println("Kafka Producer 初始化完成")
}

// 發送刷卡事件
func SendCardEvent(employeeID string, action string, timestamp time.Time) error {
	msg := kafka.Message{
		Key:   []byte(employeeID),
		Value: []byte(action + "|" + timestamp.Format(time.RFC3339)),
	}

	err := writer.WriteMessages(nil, msg)
	if err != nil {
		log.Printf("發送 Kafka 訊息失敗: %v", err)
		return err
	}

	log.Printf("已發送 Kafka 訊息: 員工 %s %s @ %s", employeeID, action, timestamp)
	return nil
}


// NATS 版本

// package events

// import (
// 	"log"
// 	"time"

// 	"github.com/nats-io/nats.go"
// )

// var nc *nats.Conn

// // 初始化 NATS 連線
// func InitNATS(natsURL string) {
// 	var err error
// 	nc, err = nats.Connect(natsURL)
// 	if err != nil {
// 		log.Fatalf("NATS 連線失敗: %v", err)
// 	}
// 	log.Println("NATS 連線成功")
// }

// // 發送刷卡事件
// func SendCardEvent(employeeID string, action string, timestamp time.Time) {
// 	event := employeeID + "|" + action + "|" + timestamp.Format(time.RFC3339)
// 	if err := nc.Publish("access.events", []byte(event)); err != nil {
// 		log.Printf("NATS 訊息發送失敗: %v", err)
// 	} else {
// 		log.Printf("已發送 NATS 訊息: %s", event)
// 	}
// }
