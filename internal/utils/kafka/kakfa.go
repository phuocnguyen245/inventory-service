package kafkaUtils

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// ReadMessageWrapper bọc hàm ReadMessage của kafka.Reader để dễ mở rộng khi cần.
func ReadMessageWrapper(ctx context.Context, reader *kafka.Reader) (kafka.Message, error) {
	return reader.ReadMessage(ctx)
}

// WriteMessageWrapper bọc hàm WriteMessages của kafka.Writer.
func WriteMessageWrapper(ctx context.Context, writer *kafka.Writer, key []byte, value []byte) error {
	return writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}
