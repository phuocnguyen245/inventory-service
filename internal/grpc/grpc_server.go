package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"inventory-service.com/m/internal/grpc/inventorypb" // Đảm bảo đường dẫn này đúng với go_package trong proto.
	"inventory-service.com/m/internal/repository"
)

// inventoryGRPCServer triển khai interface InventoryServiceServer được sinh ra từ proto.
type inventoryGRPCServer struct {
	inventorypb.UnimplementedInventoryServiceServer
	db   *sql.DB
	repo *repository.InventoryRepository
}

// CreateInventory thực hiện logic tạo mới tồn kho.
func (s *inventoryGRPCServer) CreateInventory(ctx context.Context, req *inventorypb.CreateInventoryRequest) (*inventorypb.CreateInventoryResponse, error) {
	log.Printf("Raw request received: %s - %s - %d", req.Id, req.Quantity)
	if req.Id == "" || req.Quantity < 0 {
		return &inventorypb.CreateInventoryResponse{
			Success: false,
			Message: "No item data received",
		}, nil
	}

	err := s.repo.CreateInventory(req.Id, req.Quantity)

	if err != nil {
		fmt.Println("Error creating inventory: ", err.Error())
		return &inventorypb.CreateInventoryResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	fmt.Println("Inventory created successfully")
	return &inventorypb.CreateInventoryResponse{
		Success: true,
		Message: "Inventory created successfully",
	}, nil
}

// UpdateInventory thực hiện logic cập nhật tồn kho.
func (s *inventoryGRPCServer) UpdateInventory(ctx context.Context, req *inventorypb.UpdateInventoryRequest) (*inventorypb.UpdateInventoryResponse, error) {
	log.Printf("gRPC UpdateInventory: id=%s, quantity_change=%d", req.GetId(), req.GetQuantityChange())
	// Ví dụ: Thực hiện cập nhật trong DB.
	return &inventorypb.UpdateInventoryResponse{
		Success: true,
		Message: "Inventory updated successfully",
	}, nil
}

// GetInventory thực hiện truy vấn thông tin tồn kho.
func (s *inventoryGRPCServer) GetInventory(ctx context.Context, req *inventorypb.GetInventoryRequest) (*inventorypb.GetInventoryResponse, error) {
	log.Printf("gRPC GetInventory: id=%s", req.GetId())
	// Ví dụ: Truy vấn DB để lấy số lượng tồn kho, dưới đây là giá trị mẫu.
	item := &inventorypb.InventoryItem{
		Id:       req.GetId(),
		Quantity: 100,
	}
	return &inventorypb.GetInventoryResponse{
		Item: item,
	}, nil
}

func (s *inventoryGRPCServer) GetInventories(ctx context.Context, req *inventorypb.GetInventoriesRequest) (*inventorypb.GetInventoriesResponse, error) {
	ids := req.Id

	if len(ids) == 0 {
		return &inventorypb.GetInventoriesResponse{Data: []*inventorypb.InventoryItem{}}, nil
	}

	items, err := s.repo.GetInventories(ids)

	if err != nil {
		return nil, err
	}

	return items, nil
}

// StartGRPCServer khởi chạy gRPC server trên cổng cấu hình.
// Hàm này chạy trong một goroutine và chờ tín hiệu dừng thông qua kênh grpcStop.
func StartGRPCServer(db *sql.DB, port string, grpcStop chan struct{}) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcServer, &inventoryGRPCServer{db: db, repo: repository.NewInventoryRepository(db)})
	log.Printf("gRPC Inventory Service is running on %s", port)

	// Chạy server trong một goroutine.
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Chờ tín hiệu dừng, sau đó tắt server một cách an toàn.
	<-grpcStop
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped gracefully")
}
