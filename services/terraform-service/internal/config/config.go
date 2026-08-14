package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Serveur
	ServerPort     string
	Environment    string
	LogLevel       string
	AuthServiceURL string

	// Valkey
	ValkeyAddr     string
	ValkeyPassword string
	ValkeyDB       int
	CacheTTL       time.Duration

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

	// Drift "fine" : chemin du binaire OpenTofu (vide = chemin par défaut)
	TofuPath string

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("TERRAFORM_SERVICE_PORT", "8082"),
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
	}

	// Valkey
	valkeyHost := getEnv("VALKEY_HOST", "localhost")
	valkeyPort := getEnv("VALKEY_PORT", "6379")
	cfg.ValkeyAddr = fmt.Sprintf("%s:%s", valkeyHost, valkeyPort)
	cfg.ValkeyPassword = getEnv("VALKEY_PASSWORD", "")

	redisDBStr := getEnv("VALKEY_DB", "0")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		return nil, fmt.Errorf("VALKEY_DB invalide: %v", err)
	}
	cfg.ValkeyDB = redisDB

	// TTL du cache
	cacheTTLStr := getEnv("TERRAFORM_CACHE_TTL", "5m")
	cacheTTL, err := time.ParseDuration(cacheTTLStr)
	if err != nil {
		return nil, fmt.Errorf("TERRAFORM_CACHE_TTL invalide: %v", err)
	}
	cfg.CacheTTL = cacheTTL

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

	// Drift "fine" : binaire OpenTofu
	cfg.TofuPath = getEnv("TOFU_PATH", "")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
