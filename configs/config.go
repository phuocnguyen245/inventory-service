package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config chứa các cấu hình của ứng dụng.
type Config struct {
	PostgresDSN string `yaml:"postgres_dsn"`
	RedisAddr   string `yaml:"redis_addr"`
	KafkaBroker string `yaml:"kafka_broker"`
	KafkaTopic  string `yaml:"kafka_topic"`
	DLQTopic    string `yaml:"dlq_topic"` // Thêm trường DLQTopic để xử lý DLQ.
	Port        string `yaml:"port"`      // Ví dụ: ":8080"
}

// LoadConfig đọc file cấu hình YAML và trả về cấu hình.
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
