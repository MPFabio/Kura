package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modulops/vault-service/internal/service"
)

type VaultHandler struct {
	svc *service.VaultService
}

func New(svc *service.VaultService) *VaultHandler {
	return &VaultHandler{svc: svc}
}

// GetStatus retourne l'état de santé de Vault.
func (h *VaultHandler) GetStatus(c *gin.Context) {
	status, err := h.svc.Status(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// ListSecrets liste les clés d'un path.
// GET /api/v1/vault/secrets?path=kura
func (h *VaultHandler) ListSecrets(c *gin.Context) {
	path := c.Query("path")
	keys, err := h.svc.ListSecrets(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": path, "keys": keys})
}

// GetSecret lit un secret (valeurs masquées, clés visibles).
// GET /api/v1/vault/secrets/:path
func (h *VaultHandler) GetSecret(c *gin.Context) {
	path := c.Param("path")
	// param inclut le slash initial — on retire le préfixe "/"
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	secret, err := h.svc.GetSecret(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, secret)
}

// WriteSecret crée ou met à jour un secret.
// POST /api/v1/vault/secrets/:path  { "data": { "key": "value" } }
func (h *VaultHandler) WriteSecret(c *gin.Context) {
	path := c.Param("path")
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	var req struct {
		Data map[string]interface{} `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.WriteSecret(c.Request.Context(), path, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "secret écrit", "path": path})
}

// DeleteSecret supprime un secret.
// DELETE /api/v1/vault/secrets/:path
func (h *VaultHandler) DeleteSecret(c *gin.Context) {
	path := c.Param("path")
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if err := h.svc.DeleteSecret(c.Request.Context(), path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "secret supprimé", "path": path})
}
