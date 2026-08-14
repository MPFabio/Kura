package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/modulops/pipeline-service/internal/config"
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

// RPush ajoute des éléments à une liste.
func (r *ValkeyClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LRange récupère une plage d'éléments d'une liste.
func (r *ValkeyClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LTrim conserve une plage d'éléments dans une liste.
func (r *ValkeyClient) LTrim(ctx context.Context, key string, start, stop int64) error {
	return r.client.LTrim(ctx, key, start, stop).Err()
}

// Expire définit un TTL sur une clé.
func (r *ValkeyClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// Incr incrémente un compteur.
func (r *ValkeyClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// HSet définit un champ dans un hash.
func (r *ValkeyClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet récupère un champ d'un hash.
func (r *ValkeyClient) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// HGetAll récupère tous les champs d'un hash.
func (r *ValkeyClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// Close ferme la connexion Valkey.
func (r *ValkeyClient) Close() error {
	return r.client.Close()
}
