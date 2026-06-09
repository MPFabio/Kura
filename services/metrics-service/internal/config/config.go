package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	ServerPort    string
	Environment   string
	LogLevel      string

	AuthServiceURL string

	PrometheusURL string
	GrafanaURL    string
	CacheTTL      time.Duration

	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("METRICS_SERVICE_PORT", "8086"),
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		PrometheusURL:  getEnv("PROMETHEUS_URL", "http://prometheus:9090"),
		GrafanaURL:     getEnv("GRAFANA_URL", "http://grafana:3000"),
	}

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	cfg.RedisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", "")

	ttl, err := time.ParseDuration(getEnv("METRICS_CACHE_TTL", "30s"))
	if err != nil {
		return nil, fmt.Errorf("METRICS_CACHE_TTL invalide: %v", err)
	}
	cfg.CacheTTL = ttl

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
