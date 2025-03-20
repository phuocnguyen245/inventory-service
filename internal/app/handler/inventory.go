package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"inventory-service.com/m/internal/model"
)

var ctx = context.Background()

// UpdateInventoryHandler sử dụng Gin Context
func UpdateInventoryHandler(db *sql.DB, redisClient *redis.Client, kafkaProducer *kafka.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy tham số từ query
		itemIDStr := c.Query("item_id")
		changeStr := c.Query("change")

		itemID, err := strconv.Atoi(itemIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item_id không hợp lệ"})
			return
		}
		change, err := strconv.Atoi(changeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "change không hợp lệ"})
			return
		}

		// Cập nhật PostgreSQL (nên sử dụng transaction trong trường hợp thực tế)
		_, err = db.Exec("UPDATE inventory SET quantity = quantity + $1 WHERE id = $2", change, itemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật database"})
			return
		}

		// Invalidate cache Redis
		redisKey := "inventory:" + itemIDStr
		redisClient.Del(ctx, redisKey)

		// Gửi sự kiện cập nhật qua Kafka
		event := model.InventoryUpdateEvent{
			ItemID:   itemID,
			Change:   change,
			DateTime: time.Now(),
		}
		eventBytes, _ := json.Marshal(event)
		err = kafkaProducer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(itemIDStr),
			Value: eventBytes,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi gửi sự kiện Kafka"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Inventory updated"})
	}
}
