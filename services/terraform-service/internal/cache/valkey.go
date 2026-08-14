package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/redis/go-redis/v9"
)

// ValkeyClient est un wrapper autour du client Valkey.
type ValkeyClient struct {
	client *redis.Client
}

// NewValkeyClient crée un nouveau client Valkey.
func NewValkeyClient(cfg *config.Config) (*ValkeyClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.ValkeyAddr,
		Password: cfg.ValkeyPassword,
		DB:       cfg.ValkeyDB,
	})

	// Tester la connexion
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à Valkey: %w", err)
	}

	return &ValkeyClient{
		client: client,
	}, nil
}

// Get récupère une valeur depuis Valkey.
func (r *ValkeyClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set stocke une valeur dans Valkey avec un TTL.
func (r *ValkeyClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete supprime une clé de Valkey.
func (r *ValkeyClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Keys retourne toutes les clés correspondant au pattern.
func (r *ValkeyClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Close ferme la connexion Valkey.
func (r *ValkeyClient) Close() error {
	return r.client.Close()
}
