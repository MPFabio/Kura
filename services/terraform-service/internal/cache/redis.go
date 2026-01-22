package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisClient est un wrapper autour du client Redis.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient crée un nouveau client Redis.
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Tester la connexion
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à Redis: %w", err)
	}

	return &RedisClient{
		client: client,
	}, nil
}

// Get récupère une valeur depuis Redis.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set stocke une valeur dans Redis avec un TTL.
func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete supprime une clé de Redis.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Keys retourne toutes les clés correspondant au pattern.
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Close ferme la connexion Redis.
func (r *RedisClient) Close() error {
	return r.client.Close()
}
