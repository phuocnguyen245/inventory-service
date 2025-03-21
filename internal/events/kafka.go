package events

import (
	"time"

	"github.com/segmentio/kafka-go"
)

// InitKafkaProducer khởi tạo một Kafka Writer để gửi message vào topic chỉ định.
func InitKafkaProducer(broker, topic string) (*kafka.Writer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker},
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireOne),
		// Có thể cấu hình thêm timeout, retry nếu cần
		WriteTimeout: 10 * time.Second,
	})
	// Nếu cần kiểm tra kết nối, bạn có thể gửi một message thử (tùy chọn).
	return writer, nil
}

// InitKafkaReader khởi tạo một Kafka Reader để nhận message từ topic chỉ định.
func InitKafkaReader(broker, topic string) *kafka.Reader {
	// Sử dụng một GroupID để đảm bảo tính đồng bộ của consumer group.
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  "inventory-consumer-group", // Cấu hình GroupID phù hợp với nhu cầu
		MinBytes: 10e3,                       // 10KB
		MaxBytes: 10e6,                       // 10MB
		MaxWait:  1 * time.Second,
	})
	return reader
}
