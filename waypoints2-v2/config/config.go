package config

import (
	"blackbird-eu/waypoints2-v2/internal/rabbitmq"
	"blackbird-eu/waypoints2-v2/pkg/logger"
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	RabbitMQ   rabbitmq.Config `yaml:"rabbitmq"`
	ServerInfo struct {
		Port string `yaml:"port"`
	} `yaml:"http"`
	Scanner struct {
		IntervalMinutes uint `yaml:"interval_minutes"`
	} `yaml:"scanner"`
}

var AppConfig Config

func LoadConfig() {
	file, err := os.ReadFile("config/config.yaml")
	if err != nil {
		logger.Log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(file, &AppConfig)
	if err != nil {
		logger.Log.Fatalf("Error unmarshaling config: %v", err)
	}

	logger.Log.Info("✅ Config loaded successfully.")
}
