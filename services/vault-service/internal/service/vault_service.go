package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	vault "github.com/openbao/openbao/api/v2"
	"github.com/redis/go-redis/v9"

	"github.com/modulops/vault-service/internal/config"
	"github.com/modulops/vault-service/internal/configstore"
	"github.com/modulops/vault-service/internal/models"
)

const cacheTTL = 30 * time.Second

type VaultService struct {
	client    *vault.Client
	rdb       *redis.Client
	cfgStore  *configstore.Client
	cfg       *config.Config
	mountPath string
}

func New(cfg *config.Config, rdb *redis.Client) (*VaultService, error) {
	cs := configstore.New(cfg.AuthServiceURL, "vault")
	ctx := context.Background()

	addr := cs.GetOrFallback(ctx, "vault_addr", cfg.VaultAddr)
	token := cs.GetOrFallback(ctx, "vault_token", cfg.VaultToken)
	mountPath := cs.GetOrFallback(ctx, "vault_mount_path", cfg.MountPath)

	vcfg := vault.DefaultConfig()
	vcfg.Address = addr

	client, err := vault.NewClient(vcfg)
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}
	client.SetToken(token)

	return &VaultService{
		client:    client,
		rdb:       rdb,
		cfgStore:  cs,
		cfg:       cfg,
		mountPath: mountPath,
	}, nil
}

// GetConfig retourne la config Vault actuelle (token masqué).
func (s *VaultService) GetConfig(ctx context.Context) (map[string]string, error) {
	addr := s.cfgStore.GetOrFallback(ctx, "vault_addr", s.cfg.VaultAddr)
	mount := s.cfgStore.GetOrFallback(ctx, "vault_mount_path", s.cfg.MountPath)
	tokenSet := s.cfgStore.GetOrFallback(ctx, "vault_token", s.cfg.VaultToken) != ""
	linked := "false"
	if tokenSet {
		linked = "true"
	}
	return map[string]string{
		"vault_addr":       addr,
		"vault_mount_path": mount,
		"vault_token":      "***",
		"linked":           linked,
	}, nil
}

// SetConfig met à jour addr, token et/ou mount path dans Postgres.
func (s *VaultService) SetConfig(ctx context.Context, addr, token, mountPath string) error {
	kv := map[string]string{}
	if addr != "" {
		kv["vault_addr"] = addr
	}
	if token != "" {
		kv["vault_token"] = token
	}
	if mountPath != "" {
		kv["vault_mount_path"] = mountPath
	}
	if len(kv) == 0 {
		return nil
	}
	if err := s.cfgStore.SetMany(ctx, kv); err != nil {
		return err
	}
	// Reconfigurer le client live
	if addr != "" {
		vcfg := vault.DefaultConfig()
		vcfg.Address = addr
		client, err := vault.NewClient(vcfg)
		if err != nil {
			return err
		}
		if token != "" {
			client.SetToken(token)
		} else {
			client.SetToken(s.cfgStore.GetOrFallback(ctx, "vault_token", s.cfg.VaultToken))
		}
		s.client = client
	} else if token != "" {
		s.client.SetToken(token)
	}
	if mountPath != "" {
		s.mountPath = mountPath
	}
	return nil
}

// Status retourne l'état de santé de Vault.
func (s *VaultService) Status(ctx context.Context) (*models.VaultStatus, error) {
	health, err := s.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &models.VaultStatus{
		Initialized: health.Initialized,
		Sealed:      health.Sealed,
		Version:     health.Version,
		ClusterName: health.ClusterName,
	}, nil
}

// ListSecrets liste les clés d'un path donné.
func (s *VaultService) ListSecrets(ctx context.Context, path string) ([]string, error) {
	fullPath := s.kvMetaPath(path)
	secret, err := s.client.Logical().ListWithContext(ctx, fullPath)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}
	raw, ok := secret.Data["keys"]
	if !ok {
		return []string{}, nil
	}
	keys, ok := raw.([]interface{})
	if !ok {
		return []string{}, nil
	}
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}

// GetSecret lit un secret KV v2.
func (s *VaultService) GetSecret(ctx context.Context, path string) (*models.Secret, error) {
	cacheKey := fmt.Sprintf("vault:secret:%s", path)
	if s.rdb != nil {
		if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
			var sec models.Secret
			if json.Unmarshal([]byte(cached), &sec) == nil {
				return &sec, nil
			}
		}
	}

	fullPath := s.kvDataPath(path)
	raw, err := s.client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("secret non trouvé: %s", path)
	}

	data, _ := raw.Data["data"].(map[string]interface{})
	meta, _ := raw.Data["metadata"].(map[string]interface{})

	sec := &models.Secret{
		Path: path,
		Data: data,
		Metadata: models.SecretMetadata{
			Path:    path,
			Version: intFromMeta(meta, "version"),
		},
	}

	if s.rdb != nil {
		if b, err := json.Marshal(sec); err == nil {
			s.rdb.Set(ctx, cacheKey, b, cacheTTL)
		}
	}
	return sec, nil
}

// WriteSecret crée ou met à jour un secret KV v2.
func (s *VaultService) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	fullPath := s.kvDataPath(path)
	_, err := s.client.Logical().WriteWithContext(ctx, fullPath, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return err
	}
	if s.rdb != nil {
		s.rdb.Del(ctx, fmt.Sprintf("vault:secret:%s", path))
	}
	return nil
}

// DeleteSecret supprime la dernière version d'un secret.
func (s *VaultService) DeleteSecret(ctx context.Context, path string) error {
	fullPath := s.kvDataPath(path)
	_, err := s.client.Logical().DeleteWithContext(ctx, fullPath)
	if err != nil {
		return err
	}
	if s.rdb != nil {
		s.rdb.Del(ctx, fmt.Sprintf("vault:secret:%s", path))
	}
	return nil
}

// kvDataPath construit le chemin KV v2 pour lire/écrire les données.
func (s *VaultService) kvDataPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("%s/data/%s", s.mountPath, path)
}

// kvMetaPath construit le chemin KV v2 pour lister les métadonnées.
func (s *VaultService) kvMetaPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return fmt.Sprintf("%s/metadata", s.mountPath)
	}
	return fmt.Sprintf("%s/metadata/%s", s.mountPath, path)
}

func intFromMeta(meta map[string]interface{}, key string) int {
	if meta == nil {
		return 0
	}
	switch v := meta[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}
