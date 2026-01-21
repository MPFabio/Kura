package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Serveur
	ServerPort  string
	Environment string
	LogLevel    string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTL      time.Duration

	// Kafka (pour futur usage)
	KafkaBrokers string
	KafkaGroupID string

	// Chiffrement des credentials
	EncryptionKey string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:  getEnv("TERRAFORM_SERVICE_PORT", "8082"),
		Environment: getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	// Redis
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	cfg.RedisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", "")

	redisDBStr := getEnv("REDIS_DB", "0")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		return nil, fmt.Errorf("REDIS_DB invalide: %v", err)
	}
	cfg.RedisDB = redisDB

	// TTL du cache
	cacheTTLStr := getEnv("TERRAFORM_CACHE_TTL", "5m")
	cacheTTL, err := time.ParseDuration(cacheTTLStr)
	if err != nil {
		return nil, fmt.Errorf("TERRAFORM_CACHE_TTL invalide: %v", err)
	}
	cfg.CacheTTL = cacheTTL

	// Kafka
	cfg.KafkaBrokers = getEnv("KAFKA_BROKERS", "localhost:9092")
	cfg.KafkaGroupID = getEnv("KAFKA_GROUP_ID", "terraform-service")

	// Clé de chiffrement (pour les credentials sensibles)
	cfg.EncryptionKey = getEnv("TERRAFORM_ENCRYPTION_KEY", "")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
