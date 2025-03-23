package service

import (
	"context"

	"inventory-service.com/m/internal/repository"
)

type InventoryService struct {
	repo *repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) CreateInventory(ctx context.Context, itemID string, quantity int32) error {
	return s.repo.CreateInventory(itemID, quantity)
}

func (s *InventoryService) GetInventory(ctx context.Context, itemID string) (int32, error) {
	return s.repo.GetInventory(itemID)
}
