// Package config charge la configuration du service Ansible depuis l'environnement.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort  string
	Environment string
	LogLevel    string

	AuthServiceURL string
	K8sServiceURL  string
	CodeServiceURL string

	SemaphoreURL       string
	SemaphoreAPIToken  string
	SemaphoreProjectID int

	InternalAPISecret string

	ValkeyAddr     string
	ValkeyPassword string
	ValkeyDB       int
	CacheTTL       time.Duration

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("ANSIBLE_SERVICE_PORT", "8083"),
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		K8sServiceURL:  getEnv("K8S_SERVICE_URL", "http://k8s-service:8081"),
		CodeServiceURL: getEnv("CODE_SERVICE_URL", "http://code-service:8088"),

		SemaphoreURL:      getEnv("SEMAPHORE_URL", ""),
		SemaphoreAPIToken: getEnv("SEMAPHORE_API_TOKEN", ""),

		InternalAPISecret: getEnv("INTERNAL_API_SECRET", ""),

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
	}

	projectID, err := strconv.Atoi(getEnv("SEMAPHORE_PROJECT_ID", "1"))
	if err != nil {
		return nil, fmt.Errorf("SEMAPHORE_PROJECT_ID invalide: %v", err)
	}
	cfg.SemaphoreProjectID = projectID

	valkeyHost := getEnv("VALKEY_HOST", "localhost")
	valkeyPort := getEnv("VALKEY_PORT", "6379")
	cfg.ValkeyAddr = fmt.Sprintf("%s:%s", valkeyHost, valkeyPort)
	cfg.ValkeyPassword = getEnv("VALKEY_PASSWORD", "")

	db, err := strconv.Atoi(getEnv("VALKEY_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("VALKEY_DB invalide: %v", err)
	}
	cfg.ValkeyDB = db

	ttlSeconds, err := strconv.Atoi(getEnv("ANSIBLE_CACHE_TTL", "300"))
	if err != nil {
		return nil, fmt.Errorf("ANSIBLE_CACHE_TTL invalide: %v", err)
	}
	cfg.CacheTTL = time.Duration(ttlSeconds) * time.Second

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
