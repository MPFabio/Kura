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

	// CI/CD - GitHub
	GitHubToken         string
	GitHubWebhookSecret string

	// CI/CD - GitLab
	GitLabToken         string
	GitLabWebhookSecret string

	// CI/CD - Jenkins
	JenkinsURL      string
	JenkinsUsername string
	JenkinsToken    string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:  getEnv("PIPELINE_SERVICE_PORT", "8084"),
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

	// TTL du cache pour l'historique des exécutions (24h par défaut)
	cacheTTLStr := getEnv("PIPELINE_CACHE_TTL", "24h")
	cacheTTL, err := time.ParseDuration(cacheTTLStr)
	if err != nil {
		return nil, fmt.Errorf("PIPELINE_CACHE_TTL invalide: %v", err)
	}
	cfg.CacheTTL = cacheTTL

	// Kafka
	cfg.KafkaBrokers = getEnv("KAFKA_BROKERS", "localhost:9092")
	cfg.KafkaGroupID = getEnv("KAFKA_GROUP_ID", "pipeline-service")

	// GitHub
	cfg.GitHubToken = getEnv("GITHUB_TOKEN", "")
	cfg.GitHubWebhookSecret = getEnv("GITHUB_WEBHOOK_SECRET", "")

	// GitLab
	cfg.GitLabToken = getEnv("GITLAB_TOKEN", "")
	cfg.GitLabWebhookSecret = getEnv("GITLAB_WEBHOOK_SECRET", "")

	// Jenkins
	cfg.JenkinsURL = getEnv("JENKINS_URL", "")
	cfg.JenkinsUsername = getEnv("JENKINS_USERNAME", "")
	cfg.JenkinsToken = getEnv("JENKINS_TOKEN", "")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
