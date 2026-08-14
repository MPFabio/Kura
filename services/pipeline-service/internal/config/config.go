package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Serveur
	ServerPort  string
	Environment string
	LogLevel    string

	// Config store (auth-service)
	AuthServiceURL string

	// Valkey
	ValkeyAddr     string
	ValkeyPassword string
	ValkeyDB       int
	CacheTTL       time.Duration

	// CI/CD - GitHub
	GitHubToken         string
	GitHubRepos         []string // ex: ["owner/repo1", "owner/repo2"]
	GitHubWebhookSecret string

	// CI/CD - Forgejo
	ForgejoURL           string
	ForgejoToken         string
	ForgejoRepos         []string // ex: ["owner/repo1", "owner/repo2"]
	ForgejoWebhookSecret string

	// Tracing (OpenTelemetry)
	OTLPEndpoint string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("PIPELINE_SERVICE_PORT", "8084"),
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

	// TTL du cache pour l'historique des exécutions (24h par défaut)
	cacheTTLStr := getEnv("PIPELINE_CACHE_TTL", "24h")
	cacheTTL, err := time.ParseDuration(cacheTTLStr)
	if err != nil {
		return nil, fmt.Errorf("PIPELINE_CACHE_TTL invalide: %v", err)
	}
	cfg.CacheTTL = cacheTTL

	// GitHub
	cfg.GitHubToken = getEnv("GITHUB_TOKEN", "")
	// WEBHOOK_SECRET est le secret partagé pour valider les signatures HMAC de tous les providers
	webhookSecret := getEnv("WEBHOOK_SECRET", getEnv("GITHUB_WEBHOOK_SECRET", ""))
	cfg.GitHubWebhookSecret = webhookSecret
	if reposStr := getEnv("GITHUB_REPOS", ""); reposStr != "" {
		for _, r := range splitAndTrim(reposStr, ",") {
			if r != "" {
				cfg.GitHubRepos = append(cfg.GitHubRepos, r)
			}
		}
	}

	// Forgejo
	cfg.ForgejoURL = getEnv("FORGEJO_URL", "")
	cfg.ForgejoToken = getEnv("FORGEJO_TOKEN", "")
	cfg.ForgejoWebhookSecret = getEnv("FORGEJO_WEBHOOK_SECRET", webhookSecret)
	if reposStr := getEnv("FORGEJO_REPOS", ""); reposStr != "" {
		for _, r := range splitAndTrim(reposStr, ",") {
			if r != "" {
				cfg.ForgejoRepos = append(cfg.ForgejoRepos, r)
			}
		}
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for _, p := range strings.Split(s, sep) {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
