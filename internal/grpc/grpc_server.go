package grpc

import (
	"context"
	"database/sql"
	"log"
	"net"

	"google.golang.org/grpc"
	"inventory-service.com/m/internal/grpc/inventorypb" // Đảm bảo đường dẫn này đúng với go_package trong proto.
)

// inventoryGRPCServer triển khai interface InventoryServiceServer được sinh ra từ proto.
type inventoryGRPCServer struct {
	inventorypb.UnimplementedInventoryServiceServer
	db *sql.DB
}

// CreateInventory thực hiện logic tạo mới tồn kho.
func (s *inventoryGRPCServer) CreateInventory(ctx context.Context, req *inventorypb.CreateInventoryRequest) (*inventorypb.CreateInventoryResponse, error) {
	// Ở đây bạn có thể gọi logic nghiệp vụ để insert dữ liệu vào DB.
	log.Printf("gRPC CreateInventory: item_id=%s, quantity=%d", req.GetItem().ItemId, req.GetItem().Quantity)
	// Ví dụ: Nếu insert thành công:
	return &inventorypb.CreateInventoryResponse{
		Success: true,
		Message: "Inventory created successfully",
	}, nil
}

// UpdateInventory thực hiện logic cập nhật tồn kho.
func (s *inventoryGRPCServer) UpdateInventory(ctx context.Context, req *inventorypb.UpdateInventoryRequest) (*inventorypb.UpdateInventoryResponse, error) {
	log.Printf("gRPC UpdateInventory: item_id=%s, quantity_change=%d", req.GetItemId(), req.GetQuantityChange())
	// Ví dụ: Thực hiện cập nhật trong DB.
	return &inventorypb.UpdateInventoryResponse{
		Success: true,
		Message: "Inventory updated successfully",
	}, nil
}

// GetInventory thực hiện truy vấn thông tin tồn kho.
func (s *inventoryGRPCServer) GetInventory(ctx context.Context, req *inventorypb.GetInventoryRequest) (*inventorypb.GetInventoryResponse, error) {
	log.Printf("gRPC GetInventory: item_id=%s", req.GetItemId())
	// Ví dụ: Truy vấn DB để lấy số lượng tồn kho, dưới đây là giá trị mẫu.
	item := &inventorypb.InventoryItem{
		ItemId:   req.GetItemId(),
		Quantity: 100,
	}
	return &inventorypb.GetInventoryResponse{
		Item: item,
	}, nil
}

// StartGRPCServer khởi chạy gRPC server trên cổng cấu hình.
// Hàm này chạy trong một goroutine và chờ tín hiệu dừng thông qua kênh grpcStop.
func StartGRPCServer(db *sql.DB, port string, grpcStop chan struct{}) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcServer, &inventoryGRPCServer{db: db})
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
