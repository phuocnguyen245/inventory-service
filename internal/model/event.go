package model

import "time"

type InventoryUpdateEvent struct {
	ItemID   int       `json:"item_id"`
	Change   int       `json:"change"`
	DateTime time.Time `json:"date_time"`
}
