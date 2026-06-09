package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/configstore"
)

// ConfigHandler gère la configuration persistante du terraform-service.
type ConfigHandler struct {
	cfgStore *configstore.Client
	cfg      *config.Config
}

// NewConfigHandler crée un handler de configuration.
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		cfgStore: configstore.New(cfg.AuthServiceURL, "terraform"),
		cfg:      cfg,
	}
}

// GetConfig retourne la configuration actuelle du backend S3.
// GET /api/v1/terraform/config
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	ctx := context.Background()
	bucket := h.cfgStore.GetOrFallback(ctx, "s3_bucket", h.cfg.S3Bucket)
	prefix := h.cfgStore.GetOrFallback(ctx, "s3_key_prefix", h.cfg.S3KeyPrefix)
	region := h.cfgStore.GetOrFallback(ctx, "s3_region", h.cfg.S3Region)
	endpoint := h.cfgStore.GetOrFallback(ctx, "s3_endpoint", h.cfg.S3Endpoint)
	backend := h.cfgStore.GetOrFallback(ctx, "state_backend", h.cfg.StateBackend)
	encKey := h.cfgStore.GetOrFallback(ctx, "encryption_key", h.cfg.EncryptionKey)
	encKeyMasked := ""
	if encKey != "" {
		encKeyMasked = "***"
	}
	c.JSON(http.StatusOK, gin.H{
		"state_backend":  backend,
		"s3_bucket":      bucket,
		"s3_key_prefix":  prefix,
		"s3_region":      region,
		"s3_endpoint":    endpoint,
		"encryption_key": encKeyMasked,
	})
}

// SetConfig met à jour la configuration du backend S3.
// POST /api/v1/terraform/config
func (h *ConfigHandler) SetConfig(c *gin.Context) {
	var body struct {
		StateBackend  string `json:"state_backend"`
		S3Bucket      string `json:"s3_bucket"`
		S3KeyPrefix   string `json:"s3_key_prefix"`
		S3Region      string `json:"s3_region"`
		S3Endpoint    string `json:"s3_endpoint"`
		S3AccessKeyID string `json:"s3_access_key_id"`
		S3SecretKey   string `json:"s3_secret_key"`
		EncryptionKey string `json:"encryption_key"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	kv := map[string]string{}
	if body.StateBackend != "" {
		kv["state_backend"] = body.StateBackend
	}
	if body.S3Bucket != "" {
		kv["s3_bucket"] = body.S3Bucket
	}
	if body.S3KeyPrefix != "" {
		kv["s3_key_prefix"] = body.S3KeyPrefix
	}
	if body.S3Region != "" {
		kv["s3_region"] = body.S3Region
	}
	if body.S3Endpoint != "" {
		kv["s3_endpoint"] = body.S3Endpoint
	}
	if body.S3AccessKeyID != "" {
		kv["s3_access_key_id"] = body.S3AccessKeyID
	}
	if body.S3SecretKey != "" {
		kv["s3_secret_key"] = body.S3SecretKey
	}
	if body.EncryptionKey != "" {
		kv["encryption_key"] = body.EncryptionKey
	}
	if len(kv) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aucun champ à mettre à jour"})
		return
	}
	ctx := context.Background()
	if err := h.cfgStore.SetMany(ctx, kv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "configuration terraform mise à jour"})
}
