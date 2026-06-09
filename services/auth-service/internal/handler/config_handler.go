package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modulops/auth-service/internal/repository"
)

// ConfigHandler expose les endpoints internes de gestion des configs de service.
// Ces routes sont réservées au réseau interne (pas exposées via Kong).
type ConfigHandler struct {
	repo *repository.Repository
}

func NewConfigHandler(repo *repository.Repository) *ConfigHandler {
	return &ConfigHandler{repo: repo}
}

// GetServiceConfig retourne toutes les clés d'un service.
// GET /internal/config/:service
func (h *ConfigHandler) GetServiceConfig(c *gin.Context) {
	service := c.Param("service")
	configs, err := h.repo.GetServiceConfigs(service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"service": service, "config": configs})
}

// GetServiceKey retourne une clé spécifique d'un service.
// GET /internal/config/:service/:key
func (h *ConfigHandler) GetServiceKey(c *gin.Context) {
	service := c.Param("service")
	key := c.Param("key")
	value, err := h.repo.GetServiceConfig(service, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"service": service, "key": key, "value": value})
}

// SetServiceConfigs insère ou met à jour plusieurs clés d'un service.
// POST /internal/config/:service
// Body: { "key1": "value1", "key2": "value2" }
func (h *ConfigHandler) SetServiceConfigs(c *gin.Context) {
	service := c.Param("service")
	var kv map[string]string
	if err := c.ShouldBindJSON(&kv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.SetServiceConfigs(service, kv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"service": service, "updated": len(kv)})
}
