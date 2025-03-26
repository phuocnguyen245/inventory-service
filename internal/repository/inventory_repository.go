package repository

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"inventory-service.com/m/internal/grpc/inventorypb"
)

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) CreateInventory(productId string, quantity int32) error {
	fmt.Println("Creating inventory for item: ", productId)
	query := `
		INSERT INTO inventory (id, quantity) 
		VALUES ($1, $2)
	`
	fmt.Println("Querying")
	_, err := r.db.Exec(query, productId, quantity)
	return err
}

func (r *InventoryRepository) GetInventory(itemID string) (int32, error) {
	var quantity int32
	err := r.db.QueryRow("SELECT quantity FROM inventories WHERE id = $1", itemID).Scan(&quantity)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return quantity, err
}

func (r *InventoryRepository) GetInventories(itemIDs []string) (*inventorypb.GetInventoriesResponse, error) {

	rows, err := r.db.Query("SELECT id, quantity FROM inventories WHERE id = ANY($1)", pq.Array(itemIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*inventorypb.InventoryItem

	for rows.Next() {
		r := &inventorypb.InventoryItem{}
		if err := rows.Scan(&r.Id, &r.Quantity); err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	response := &inventorypb.GetInventoriesResponse{
		Data: result,
	}

	return response, rows.Err()
}
