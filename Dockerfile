FROM golang:1.24.1-bookworm

WORKDIR /app

# Copy file go.mod và go.sum, tải các dependency
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ mã nguồn
COPY . .

# Build ứng dụng
RUN go build -o inventory-service ./

# Expose các port cần thiết: HTTP (8080) và gRPC (50053)
EXPOSE 8080
EXPOSE 50053

CMD ["./inventory-service"]
