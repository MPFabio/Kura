package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/modulops/k8s-service/internal/config"
)

// ValkeyClient encapsule le client Valkey.
type ValkeyClient struct {
	client *redis.Client
	cfg    *config.Config
}

// NewValkeyClient initialise un nouveau client Valkey.
func NewValkeyClient(cfg *config.Config) (*ValkeyClient, error) {
	if cfg.ValkeyAddr == "" {
		return nil, fmt.Errorf("ValkeyAddr ne peut pas être vide")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.ValkeyAddr,
		Password: cfg.ValkeyPassword,
		DB:       cfg.ValkeyDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à Valkey: %w", err)
	}

	return &ValkeyClient{
		client: rdb,
		cfg:    cfg,
	}, nil
}

// Get récupère une valeur du cache.
func (r *ValkeyClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set stocke une valeur dans le cache avec TTL.
func (r *ValkeyClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete supprime une clé du cache.
func (r *ValkeyClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Keys retourne toutes les clés correspondant au pattern.
func (r *ValkeyClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Close ferme la connexion Valkey.
func (r *ValkeyClient) Close() error {
	return r.client.Close()
}
