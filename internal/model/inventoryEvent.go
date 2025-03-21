package model

import "time"

// InventoryEventType định nghĩa các loại event.
type InventoryEventType string

const (
	EventTypeCreate InventoryEventType = "create"
	EventTypeUpdate InventoryEventType = "update"
	EventTypeDelete InventoryEventType = "delete"
)

// InventoryEvent định nghĩa cấu trúc chung của các event liên quan đến inventory.
type InventoryEvent struct {
	Type     InventoryEventType `json:"type"`      // Loại event: create, update, delete, ...
	ItemID   string             `json:"item_id"`   // ID của sản phẩm
	Quantity int                `json:"quantity"`  // Số lượng, dùng cho create/update
	DateTime time.Time          `json:"date_time"` // Thời gian event xảy ra
}
