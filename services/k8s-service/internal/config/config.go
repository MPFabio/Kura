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

	// Authentification interne (appels service-à-service, ex: Semaphore)
	InternalAPISecret string

	// Kubernetes
	KubeconfigPath string
	InCluster      bool
	K8sAPITimeout  time.Duration
	K8sMaxRetries  int

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTL      time.Duration

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("K8S_SERVICE_PORT", "8081"),
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		KubeconfigPath: getEnv("KUBECONFIG_PATH", ""),

		InternalAPISecret: getEnv("INTERNAL_API_SECRET", ""),

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
	}

	// In-cluster ou kubeconfig
	inClusterStr := getEnv("K8S_INCLUSTER", "false")
	inCluster, err := strconv.ParseBool(inClusterStr)
	if err != nil {
		return nil, fmt.Errorf("K8S_INCLUSTER invalide: %v", err)
	}
	cfg.InCluster = inCluster

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
	cacheTTLStr := getEnv("K8S_CACHE_TTL", "30s")
	cacheTTL, err := time.ParseDuration(cacheTTLStr)
	if err != nil {
		return nil, fmt.Errorf("K8S_CACHE_TTL invalide: %v", err)
	}
	cfg.CacheTTL = cacheTTL

	// Timeout API Kubernetes (défaut: 30s, plus long en prod pour les connexions distantes)
	apiTimeoutStr := getEnv("K8S_API_TIMEOUT", "30s")
	apiTimeout, err := time.ParseDuration(apiTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("K8S_API_TIMEOUT invalide: %v", err)
	}
	cfg.K8sAPITimeout = apiTimeout

	// Nombre de tentatives pour les requêtes Kubernetes
	maxRetriesStr := getEnv("K8S_MAX_RETRIES", "3")
	maxRetries, err := strconv.Atoi(maxRetriesStr)
	if err != nil {
		return nil, fmt.Errorf("K8S_MAX_RETRIES invalide: %v", err)
	}
	cfg.K8sMaxRetries = maxRetries

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
