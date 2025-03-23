package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"time"

	"inventory-service.com/m/internal/model"
	kafkaUtils "inventory-service.com/m/internal/utils/kafka"
	redisUtils "inventory-service.com/m/internal/utils/redis"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

// InventoryConsumer xử lý các sự kiện từ Kafka và cập nhật inventory.
type InventoryConsumer struct {
	db           *sql.DB
	redisClient  *redis.Client
	kafkaReader  *kafka.Reader
	dlqWriter    *kafka.Writer
	workerCount  int
	workerQueues []chan model.InventoryEvent // mảng channel cho mỗi worker
}

// getWorkerIndex tính chỉ số worker dựa trên giá trị string key (ví dụ: ItemID)
func getWorkerIndex(key string, workerCount int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % workerCount
}

// NewInventoryConsumer tạo mới một InventoryConsumer với số lượng worker mong muốn.
func NewInventoryConsumer(db *sql.DB, redisClient *redis.Client, kafkaReader *kafka.Reader, dlqWriter *kafka.Writer, workerCount int) *InventoryConsumer {
	queues := make([]chan model.InventoryEvent, workerCount)
	for i := 0; i < workerCount; i++ {
		queues[i] = make(chan model.InventoryEvent, 100) // mỗi channel có bộ đệm 100 event
	}
	return &InventoryConsumer{
		db:           db,
		redisClient:  redisClient,
		kafkaReader:  kafkaReader,
		dlqWriter:    dlqWriter,
		workerCount:  workerCount,
		workerQueues: queues,
	}
}

// pushToDLQ đưa event vào DLQ thông qua kafkaUtils.
func (c *InventoryConsumer) pushToDLQ(ctx context.Context, event model.InventoryEvent) {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Lỗi mã hóa event cho DLQ: %v", err)
		return
	}
	// Sử dụng event.ItemID làm key (có thể thay đổi nếu cần)
	err = kafkaUtils.WriteMessageWrapper(ctx, c.dlqWriter, []byte(event.Id), eventBytes)
	if err != nil {
		log.Printf("Lỗi gửi event vào DLQ: %v", err)
	} else {
		log.Printf("Event được đưa vào DLQ: itemID %s", event.Id)
	}
}

// Start bắt đầu vòng lặp đọc message từ Kafka và phân phối event vào worker pool.
func (c *InventoryConsumer) Start(ctx context.Context) {
	// Khởi chạy worker pool: mỗi worker lắng nghe một channel riêng.
	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx, c.workerQueues[i], i)
	}

	// Vòng lặp đọc message từ Kafka.
readLoop:
	for {
		msg, err := kafkaUtils.ReadMessageWrapper(ctx, c.kafkaReader)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("Context bị hủy, dừng nhận event")
				break readLoop
			default:
			}
			log.Printf("Lỗi đọc message Kafka: %v", err)
			continue
		}

		var event model.InventoryEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Lỗi giải mã message: %v", err)
			continue
		}

		// Sử dụng hàm getWorkerIndex để tính chỉ số worker từ event.ItemID (kiểu string)
		workerIndex := getWorkerIndex(event.Id, c.workerCount)

		select {
		case c.workerQueues[workerIndex] <- event:
			// event được gửi thành công vào channel của worker
		case <-ctx.Done():
			log.Println("Context bị hủy, dừng nhận event")
			break readLoop
		}
	}

	// Đóng tất cả các channel của worker để báo hiệu dừng.
	for _, queue := range c.workerQueues {
		close(queue)
	}
}

// worker xử lý các event nhận từ channel của nó theo thứ tự FIFO.
func (c *InventoryConsumer) worker(ctx context.Context, eventChan <-chan model.InventoryEvent, workerID int) {
	for event := range eventChan {
		if err := c.processEvent(ctx, event); err != nil {
			log.Printf("Worker %d: lỗi xử lý event cho item %s: %v", workerID, event.Id, err)
		} else {
			log.Printf("Worker %d: xử lý event %s cho item %s thành công", workerID, event.Type, event.Id)
		}
	}
}

// processEvent phân loại và xử lý event theo loại, có cơ chế retry 3 lần.
func (c *InventoryConsumer) processEvent(ctx context.Context, event model.InventoryEvent) error {
	var err error
	attempts := 3
	switch event.Type {
	case model.EventTypeCreate:
		for i := 0; i < attempts; i++ {
			err = c.attemptProcessCreate(ctx, event)
			if err == nil {
				return nil
			}
			log.Printf("Lỗi processCreate attempt %d cho item %s: %v", i+1, event.Id, err)
			time.Sleep(1 * time.Second)
		}
	case model.EventTypeUpdate:
		for i := 0; i < attempts; i++ {
			err = c.attemptProcessUpdate(ctx, event)
			if err == nil {
				return nil
			}
			log.Printf("Lỗi processUpdate attempt %d cho item %s: %v", i+1, event.Id, err)
			time.Sleep(1 * time.Second)
		}
	case model.EventTypeDelete:
		for i := 0; i < attempts; i++ {
			err = c.attemptProcessDelete(ctx, event)
			if err == nil {
				return nil
			}
			log.Printf("Lỗi processDelete attempt %d cho item %s: %v", i+1, event.Id, err)
			time.Sleep(1 * time.Second)
		}
	default:
		return fmt.Errorf("loại event không xác định: %s", event.Type)
	}
	// Nếu sau 3 lần vẫn thất bại, đưa event vào DLQ.
	c.pushToDLQ(ctx, event)
	return fmt.Errorf("xử lý event %s cho item %s thất bại sau %d lần: %v", event.Type, event.Id, attempts, err)
}

func (c *InventoryConsumer) attemptProcessCreate(ctx context.Context, event model.InventoryEvent) error {
	_, err := c.db.ExecContext(ctx, "INSERT INTO inventory (id, quantity) VALUES ($1, $2)", event.Id, event.Quantity)
	if err != nil {
		return fmt.Errorf("lỗi insert database: %v", err)
	}
	cacheKey := fmt.Sprintf("inventory:%s", event.Id)
	if err := redisUtils.InvalidateCache(ctx, c.redisClient, cacheKey); err != nil {
		log.Printf("Lỗi xoá cache cho item %s: %v", event.Id, err)
	}
	return nil
}

func (c *InventoryConsumer) attemptProcessUpdate(ctx context.Context, event model.InventoryEvent) error {
	lockKey := fmt.Sprintf("lock:inventory:%s", event.Id)
	lockExpiration := 10 * time.Second

	locked, err := redisUtils.AcquireLock(ctx, c.redisClient, lockKey, lockExpiration)
	if err != nil || !locked {
		if err != nil {
			log.Printf("Lỗi đặt khóa cho item %s: %v", event.Id, err)
		}
		return fmt.Errorf("không thể lấy lock cho item %s", event.Id)
	}
	defer func() {
		if err := redisUtils.ReleaseLock(ctx, c.redisClient, lockKey); err != nil {
			log.Printf("Lỗi giải phóng lock %s: %v", lockKey, err)
		}
	}()

	_, err = c.db.ExecContext(ctx, "UPDATE inventory SET quantity = quantity + $1 WHERE id = $2", event.Quantity, event.Id)
	if err != nil {
		return fmt.Errorf("lỗi cập nhật database: %v", err)
	}

	cacheKey := fmt.Sprintf("inventory:%s", event.Id)
	if err := redisUtils.InvalidateCache(ctx, c.redisClient, cacheKey); err != nil {
		log.Printf("Lỗi xoá cache cho item %s: %v", event.Id, err)
	}

	return nil
}

func (c *InventoryConsumer) attemptProcessDelete(ctx context.Context, event model.InventoryEvent) error {
	lockKey := fmt.Sprintf("lock:inventory:%s", event.Id)
	lockExpiration := 10 * time.Second

	locked, err := redisUtils.AcquireLock(ctx, c.redisClient, lockKey, lockExpiration)
	if err != nil || !locked {
		if err != nil {
			log.Printf("Lỗi đặt khóa cho item %s: %v", event.Id, err)
		}
		return fmt.Errorf("không thể lấy lock cho item %s", event.Id)
	}
	defer func() {
		if err := redisUtils.ReleaseLock(ctx, c.redisClient, lockKey); err != nil {
			log.Printf("Lỗi giải phóng lock %s: %v", lockKey, err)
		}
	}()

	_, err = c.db.ExecContext(ctx, "DELETE FROM inventory WHERE id = $1", event.Id)
	if err != nil {
		return fmt.Errorf("lỗi xóa database: %v", err)
	}

	cacheKey := fmt.Sprintf("inventory:%s", event.Id)
	if err := redisUtils.InvalidateCache(ctx, c.redisClient, cacheKey); err != nil {
		log.Printf("Lỗi xoá cache cho item %s: %v", event.Id, err)
	}

	return nil
}

// StartDLQConsumer đọc các event từ DLQ và cố gắng reprocess chúng.
// Nếu reprocess không thành công, bạn có thể lưu trữ hoặc gửi cảnh báo.
func (c *InventoryConsumer) StartDLQConsumer(ctx context.Context, dlqReader *kafka.Reader) {
	for {
		msg, err := kafkaUtils.ReadMessageWrapper(ctx, dlqReader)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("Context bị hủy, dừng DLQ consumer")
				return
			default:
				log.Printf("Lỗi đọc message DLQ: %v", err)
				continue
			}
		}

		var event model.InventoryEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Lỗi giải mã DLQ event: %v", err)
			// Ở đây có thể chuyển event sang một hệ thống lưu trữ lỗi khác để xử lý thủ công.
			continue
		}

		log.Printf("Đang cố gắng reprocess event từ DLQ cho item %s", event.Id)
		// Cố gắng reprocess event từ DLQ.
		if err := c.processEvent(ctx, event); err != nil {
			log.Printf("Reprocess DLQ event thất bại cho item %s: %v", event.Id, err)
			// Nếu reprocess không thành công, bạn có thể lưu trữ event này vào database hoặc hệ thống giám sát để xử lý sau.
		} else {
			log.Printf("Reprocess DLQ event thành công cho item %s", event.Id)
		}
	}
}
