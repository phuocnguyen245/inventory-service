package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config chứa các cấu hình của ứng dụng.
type Config struct {
	PostgresDSN string
	RedisAddr   string
	KafkaBroker string
	KafkaTopic  string
	DLQTopic    string
	Port        string
	GRPCPort    string
}

func LoadConfig(path ...string) (*Config, error) {

	if len(path) == 0 {
		path = append(path, ".env")
	}

	fmt.Println("path: ", path, os.Getenv("POSTGRES_DSN"))

	err := godotenv.Load(path[0])

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
		KafkaBroker: os.Getenv("KAFKA_BROKER"),
		KafkaTopic:  os.Getenv("KAFKA_TOPIC"),
		DLQTopic:    os.Getenv("DLQ_TOPIC"),
		Port:        os.Getenv("PORT"),
		GRPCPort:    os.Getenv("GRPC_PORT"),
	}, nil
}
