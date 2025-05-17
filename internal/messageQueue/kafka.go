package messagequeue

import (
	"strings"

	"github.com/Shopify/sarama"
)

var (
	producer sarama.SyncProducer
)

func InitKafka(brokers string) error {
	brokerList := strings.Split(brokers, ",")
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	var err error
	producer, err = sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return err
	}

	return nil
}

func GetProducer() sarama.SyncProducer {
	return producer
}
