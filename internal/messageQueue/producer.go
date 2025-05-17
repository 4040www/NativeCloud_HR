package messagequeue

import (
	"encoding/json"

	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/Shopify/sarama"
)

const CheckInTopic = "checkin-topic" // 你的 Kafka topic 名字

func SendCheckIn(record *model.CheckInRequest) error {
	producer := GetProducer()

	bytes, err := json.Marshal(record)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: CheckInTopic,
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = producer.SendMessage(msg)
	return err
}
