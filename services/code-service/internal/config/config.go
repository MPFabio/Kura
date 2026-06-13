package config

import "os"

type Config struct {
	ServerPort     string
	Environment    string
	AuthServiceURL string

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("CODE_SERVICE_PORT", "8088"),
		Environment:    getEnv("ENV", "development"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
