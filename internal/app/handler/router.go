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

	// Đăng ký route cho việc cập nhật inventory
	router.PUT("/update-inventory", UpdateInventoryHandler(db, redisClient, kafkaProducer))

	// Các route khác có thể đăng ký thêm tại đây...

	return router
}
