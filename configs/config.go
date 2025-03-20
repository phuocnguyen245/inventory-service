package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	PostgresDSN string `yaml:"postgres_dsn"`
	RedisAddr   string `yaml:"redis_addr"`
	KafkaBroker string `yaml:"kafka_broker"`
	KafkaTopic  string `yaml:"kafka_topic"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
