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
	LokiURL       string
	TempoURL      string
	CacheTTL      time.Duration

	// DeploymentMode vaut "saas" (par défaut) ou "self-hosted".
	// En mode "saas", l'observabilité interne de la plateforme Kura
	// (santé/metrics des microservices Kura) n'est pas exposée aux clients.
	DeploymentMode               string
	InternalObservabilityEnabled bool

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("METRICS_SERVICE_PORT", "8086"),
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		PrometheusURL:  getEnv("PROMETHEUS_URL", "http://prometheus:9090"),
		GrafanaURL:     getEnv("GRAFANA_URL", "http://grafana:3000"),
		LokiURL:        getEnv("LOKI_URL", "http://loki:3100"),
		TempoURL:       getEnv("TEMPO_URL", "http://tempo:3200"),

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
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

	cfg.DeploymentMode = getEnv("DEPLOYMENT_MODE", "saas")
	cfg.InternalObservabilityEnabled = getEnvBool("INTERNAL_OBSERVABILITY_ENABLED", cfg.DeploymentMode != "saas")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v == "true" || v == "1"
}
