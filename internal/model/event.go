package model

import "time"

type InventoryUpdateEvent struct {
	Id       string    `json:"id"`
	Change   int       `json:"change"`
	DateTime time.Time `json:"date_time"`
}
