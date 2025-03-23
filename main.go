package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-service.com/m/configs"
	"inventory-service.com/m/internal/app/consumer"
	"inventory-service.com/m/internal/app/handler"
	"inventory-service.com/m/internal/cache"
	"inventory-service.com/m/internal/db"
	"inventory-service.com/m/internal/events"
	grpcServer "inventory-service.com/m/internal/grpc" // Giả sử file grpc_server.go nằm trong package main của cmd/inventory
)

func main() {
	// 1. Load cấu hình từ file YAML.
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Kết nối PostgreSQL.
	dbConn, err := db.InitPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Error connecting to Postgres: %v", err)
	}
	defer dbConn.Close()

	// 3. Kết nối Redis thông qua package cache.
	redisClient, err := cache.InitRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	// Giả sử package cache cung cấp hàm Close nếu cần, hoặc để garbage collection quản lý.

	// 4. Khởi tạo Kafka Producer cho topic chính.
	kafkaProducer, err := events.InitKafkaProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("Error initializing Kafka Producer: %v", err)
	}
	defer kafkaProducer.Close()

	// 5. Khởi tạo Kafka Reader cho topic chính và DLQ, đồng thời Kafka Producer cho DLQ.
	kafkaReader := events.InitKafkaReader(cfg.KafkaBroker, cfg.KafkaTopic)
	dlqReader := events.InitKafkaReader(cfg.KafkaBroker, cfg.DLQTopic)
	dlqWriter, err := events.InitKafkaProducer(cfg.KafkaBroker, cfg.DLQTopic)
	if err != nil {
		log.Fatalf("Error initializing Kafka Producer for DLQ: %v", err)
	}
	defer dlqWriter.Close()

	// 6. Thiết lập Gin router.
	router := handler.SetupRouter(dbConn, redisClient, kafkaProducer)

	// 7. Tạo HTTP server với graceful shutdown.
	httpSrv := &http.Server{
		Addr:    cfg.Port, // Ví dụ: ":8080"
		Handler: router,
	}

	// 8. Tạo context gốc để quản lý vòng đời của HTTP server và consumer.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 9. Khởi chạy consumer chính và DLQ consumer trong các goroutine riêng.
	workerCount := 5 // Số lượng worker cho consumer.
	invConsumer := consumer.NewInventoryConsumer(dbConn, redisClient, kafkaReader, dlqWriter, workerCount)
	go invConsumer.Start(ctx)
	go invConsumer.StartDLQConsumer(ctx, dlqReader)

	// 10. Khởi chạy gRPC server trên cổng cấu hình (ví dụ: ":50051").
	grpcStop := make(chan struct{})
	go grpcServer.StartGRPCServer(dbConn, cfg.GRPCPort, grpcStop)

	// 11. Khởi chạy HTTP server trong goroutine riêng.
	go func() {
		log.Printf("HTTP server running at %s", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 12. Lắng nghe tín hiệu dừng từ hệ điều hành để graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, shutting down...")

	// Hủy context để báo hiệu dừng cho consumer và các goroutine khác.
	cancel()
	close(grpcStop)

	// Đóng HTTP server với timeout cho graceful shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
