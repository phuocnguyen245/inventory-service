package events

import (
	"github.com/segmentio/kafka-go"
)

func InitKafkaProducer(broker, topic string) (*kafka.Writer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   topic,
	})
	return writer, nil
}
