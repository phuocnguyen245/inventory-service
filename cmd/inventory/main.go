package main

import (
	"log"

	"inventory-service.com/m/internal/app/handler"
	"inventory-service.com/m/internal/cache"
	"inventory-service.com/m/internal/db"
	"inventory-service.com/m/internal/events"
	"inventory-service.com/m/pkg/config"
)

func main() {
	// 1. Load cấu hình
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Lỗi đọc cấu hình: %v", err)
	}

	// 2. Kết nối PostgreSQL
	dbConn, err := db.InitPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Lỗi kết nối PostgreSQL: %v", err)
	}
	defer dbConn.Close()

	// 3. Kết nối Redis
	redisClient, err := cache.InitRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Lỗi kết nối Redis: %v", err)
	}
	defer redisClient.Close()

	// 4. Khởi tạo Kafka Producer
	kafkaProducer, err := events.InitKafkaProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("Lỗi kết nối Kafka: %v", err)
	}
	defer kafkaProducer.Close()

	// 5. Thiết lập Gin router
	router := handler.SetupRouter(dbConn, redisClient, kafkaProducer)

	// 6. Chạy server trên port 8080
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Lỗi khởi chạy server: %v", err)
	}
}
