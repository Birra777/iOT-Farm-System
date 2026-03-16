package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all environment-driven configuration for AgriStream services.
type Config struct {
	// Kafka
	KafkaBrokers         string
	KafkaTopicSoilRaw    string
	KafkaTopicWeatherRaw string
	KafkaTopicAggregated string
	KafkaTopicAlerts     string
	KafkaTopicDLQ        string

	// PostgreSQL
	PostgresDSN string

	// API
	APIPort int

	// Simulator
	SimulatorPublishIntervalSeconds int

	// Email alerts (optional — leave SMTPHost empty to disable)
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	AlertEmailTo string // comma-separated list of recipient addresses

	// AI Advisor (optional — leave AnthropicAPIKey empty to disable)
	AnthropicAPIKey string
	AnthropicModel  string
}

// Load reads .env (if present) then environment variables and returns a Config.
func Load() (*Config, error) {
	// Best-effort load of .env; ignore error if file absent in production.
	_ = godotenv.Load()

	cfg := &Config{
		KafkaBrokers:         getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicSoilRaw:    getEnv("KAFKA_TOPIC_SOIL_RAW", "sensors.soil.raw"),
		KafkaTopicWeatherRaw: getEnv("KAFKA_TOPIC_WEATHER_RAW", "sensors.weather.raw"),
		KafkaTopicAggregated: getEnv("KAFKA_TOPIC_AGGREGATED", "sensors.aggregated"),
		KafkaTopicAlerts:     getEnv("KAFKA_TOPIC_ALERTS", "alerts.anomaly"),
		KafkaTopicDLQ:        getEnv("KAFKA_TOPIC_DLQ", "sensors.dlq"),
		PostgresDSN:          getEnv("POSTGRES_DSN", "postgres://agristream:agristream@localhost:5433/agristream?sslmode=disable"),
	}

	var err error
	cfg.APIPort, err = getEnvInt("API_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid API_PORT: %w", err)
	}

	cfg.SimulatorPublishIntervalSeconds, err = getEnvInt("SIMULATOR_PUBLISH_INTERVAL_SECONDS", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid SIMULATOR_PUBLISH_INTERVAL_SECONDS: %w", err)
	}

	cfg.SMTPHost = getEnv("SMTP_HOST", "")
	cfg.SMTPPort = getEnv("SMTP_PORT", "587")
	cfg.SMTPUser = getEnv("SMTP_USER", "")
	cfg.SMTPPass = getEnv("SMTP_PASS", "")
	cfg.AlertEmailTo = getEnv("ALERT_EMAIL_TO", "")

	cfg.AnthropicAPIKey = getEnv("ANTHROPIC_API_KEY", "")
	cfg.AnthropicModel  = getEnv("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001")

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(v)
}
