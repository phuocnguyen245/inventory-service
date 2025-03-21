package handler

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

// SetupRouter đăng ký các route cho ứng dụng
func SetupRouter(db *sql.DB, redisClient *redis.Client, kafkaProducer *kafka.Writer) *gin.Engine {
	router := gin.Default()

	handler := NewHandler(db, redisClient, kafkaProducer)
	// Đăng ký route cho việc cập nhật inventory với method của struct Handler
	router.PUT("/update-inventory", handler.UpdateInventoryHandler)

	// Các route khác có thể đăng ký thêm tại đây...

	return router
}
