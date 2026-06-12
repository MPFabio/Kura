package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort     string
	Environment    string
	LogLevel       string
	AuthServiceURL string
	VaultAddr      string
	VaultToken     string
	MountPath      string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	redisDB := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		fmt.Sscanf(v, "%d", &redisDB)
	}

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")

	return &Config{
		ServerPort:     getEnv("VAULT_SERVICE_PORT", "8087"),
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		VaultAddr:      getEnv("VAULT_ADDR", "http://vault:8200"),
		VaultToken:     getEnv("VAULT_TOKEN", ""),
		MountPath:      getEnv("VAULT_MOUNT_PATH", "secret"),
		RedisAddr:      fmt.Sprintf("%s:%s", redisHost, redisPort),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        redisDB,

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
