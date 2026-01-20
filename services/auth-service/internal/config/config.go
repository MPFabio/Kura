package config

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	// Serveur
	ServerPort  string
	Environment string
	LogLevel    string

	// Base de données
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration
	JWTIssuer     string

	// OAuth2 (optionnel)
	OAuth2GoogleClientID     string
	OAuth2GoogleClientSecret string
	OAuth2GitHubClientID     string
	OAuth2GitHubClientSecret string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:  getEnv("AUTH_SERVICE_PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "modulops"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "modulops"),
		PostgresDB:       getEnv("POSTGRES_DB", "modulops"),

		JWTIssuer: "modulops-auth-service",
	}

	// JWT Secret (obligatoire)
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET doit être défini")
	}
	cfg.JWTSecret = jwtSecret

	// JWT Expiration
	jwtExpirationStr := getEnv("JWT_EXPIRATION", "24h")
	jwtExpiration, err := time.ParseDuration(jwtExpirationStr)
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION invalide: %v", err)
	}
	cfg.JWTExpiration = jwtExpiration

	// OAuth2 (optionnel)
	cfg.OAuth2GoogleClientID = getEnv("OAUTH2_GOOGLE_CLIENT_ID", "")
	cfg.OAuth2GoogleClientSecret = getEnv("OAUTH2_GOOGLE_CLIENT_SECRET", "")
	cfg.OAuth2GitHubClientID = getEnv("OAUTH2_GITHUB_CLIENT_ID", "")
	cfg.OAuth2GitHubClientSecret = getEnv("OAUTH2_GITHUB_CLIENT_SECRET", "")

	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDB)
}

func (c *Config) GetJWTKey() []byte {
	return []byte(c.JWTSecret)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Claims représente les claims JWT personnalisés
type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}
