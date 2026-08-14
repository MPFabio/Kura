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
	OpenBaoAddr    string
	OpenBaoToken   string
	MountPath      string
	ValkeyAddr     string
	ValkeyPassword string
	ValkeyDB       int

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	redisDB := 0
	if v := os.Getenv("VALKEY_DB"); v != "" {
		fmt.Sscanf(v, "%d", &redisDB)
	}

	valkeyHost := getEnv("VALKEY_HOST", "localhost")
	valkeyPort := getEnv("VALKEY_PORT", "6379")

	return &Config{
		ServerPort:     getEnv("VAULT_SERVICE_PORT", "8087"),
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		OpenBaoAddr:    getEnv("OPENBAO_ADDR", "http://openbao:8200"),
		OpenBaoToken:   getEnv("OPENBAO_TOKEN", ""),
		MountPath:      getEnv("OPENBAO_MOUNT_PATH", "secret"),
		ValkeyAddr:     fmt.Sprintf("%s:%s", valkeyHost, valkeyPort),
		ValkeyPassword: getEnv("VALKEY_PASSWORD", ""),
		ValkeyDB:       redisDB,

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
