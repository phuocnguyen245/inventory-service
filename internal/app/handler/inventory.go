package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"inventory-service.com/m/internal/model"
)

type Handler struct {
	db            *sql.DB
	redisClient   *redis.Client
	kafkaProducer *kafka.Writer
}

func NewHandler(db *sql.DB, redisClient *redis.Client, kafkaProducer *kafka.Writer) *Handler {
	return &Handler{
		db:            db,
		redisClient:   redisClient,
		kafkaProducer: kafkaProducer,
	}
}

func (h *Handler) UpdateInventoryHandler(c *gin.Context) {
	ctx := c.Request.Context() // dùng context từ request
	idStr := c.Query("id")
	changeStr := c.Query("change")

	change, err := strconv.Atoi(changeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "change không hợp lệ"})
		return
	}

	// Có thể cân nhắc sử dụng transaction nếu có nhiều thao tác liên quan
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khởi tạo transaction"})
		return
	}
	// Cập nhật PostgreSQL trong transaction
	_, err = tx.ExecContext(ctx, "UPDATE inventory SET quantity = quantity + $1 WHERE id = $2", change, idStr)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật database"})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi commit transaction"})
		return
	}

	// Invalidate cache Redis
	redisKey := "inventory:" + idStr
	if err := h.redisClient.Del(ctx, redisKey).Err(); err != nil {
		// Log lỗi, không nhất thiết trả về cho client
		// log.Printf("Lỗi xóa key Redis %s: %v", redisKey, err)
	}

	// Gửi sự kiện cập nhật qua Kafka
	event := model.InventoryUpdateEvent{
		Id:       idStr,
		Change:   change,
		DateTime: time.Now(),
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi mã hóa sự kiện"})
		return
	}
	err = h.kafkaProducer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(idStr),
		Value: eventBytes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi gửi sự kiện Kafka"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory updated"})
}
