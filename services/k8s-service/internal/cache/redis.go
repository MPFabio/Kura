package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/modulops/k8s-service/internal/config"
)

// RedisClient encapsule le client Redis.
type RedisClient struct {
	client *redis.Client
	cfg    *config.Config
}

// NewRedisClient initialise un nouveau client Redis.
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("RedisAddr ne peut pas être vide")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à Redis: %w", err)
	}

	return &RedisClient{
		client: rdb,
		cfg:    cfg,
	}, nil
}

// Get récupère une valeur du cache.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set stocke une valeur dans le cache avec TTL.
func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete supprime une clé du cache.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Keys retourne toutes les clés correspondant au pattern.
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Close ferme la connexion Redis.
func (r *RedisClient) Close() error {
	return r.client.Close()
}

