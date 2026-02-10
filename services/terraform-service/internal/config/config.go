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

	// Drift worker : intervalle entre chaque vérification (ex. 1h)
	DriftWorkerInterval time.Duration

	// Chiffrement des credentials
	EncryptionKey string

	// Backend tfstate (S3 / MinIO) : persistance des états dans un bucket
	StateBackend  string // "s3" pour activer
	S3Bucket      string
	S3KeyPrefix   string
	S3Region      string
	S3Endpoint    string // ex. http://minio:9000 pour MinIO
	S3AccessKeyID string
	S3SecretKey   string
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

	// Drift worker
	driftIntervalStr := getEnv("TERRAFORM_DRIFT_WORKER_INTERVAL", "1h")
	if d, err := time.ParseDuration(driftIntervalStr); err == nil && d > 0 {
		cfg.DriftWorkerInterval = d
	}

	// Clé de chiffrement (pour les credentials sensibles)
	cfg.EncryptionKey = getEnv("TERRAFORM_ENCRYPTION_KEY", "")

	// Backend tfstate S3 / MinIO
	cfg.StateBackend = getEnv("TERRAFORM_STATE_BACKEND", "")
	cfg.S3Bucket = getEnv("AWS_S3_BUCKET", "kura-tfstate")
	cfg.S3KeyPrefix = getEnv("AWS_S3_KEY_PREFIX", "tfstate")
	cfg.S3Region = getEnv("AWS_S3_REGION", "us-east-1")
	cfg.S3Endpoint = getEnv("S3_ENDPOINT", "")
	cfg.S3AccessKeyID = getEnv("AWS_ACCESS_KEY_ID", "")
	cfg.S3SecretKey = getEnv("AWS_SECRET_ACCESS_KEY", "")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
