package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modulops/k8s-service/internal/service"
)

// DiscoveryHandler gère les requêtes HTTP liées à l'auto-découverte des
// applications et composants du cluster client.
type DiscoveryHandler struct {
	svc *service.DiscoveryService
}

// NewDiscoveryHandler crée un nouveau handler de découverte.
func NewDiscoveryHandler(svc *service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{svc: svc}
}

// GetReport godoc
// GET /api/v1/k8s/discovery
func (h *DiscoveryHandler) GetReport(c *gin.Context) {
	report, err := h.svc.GetReport(c.Request.Context())
	if err != nil {
		log.Printf("Erreur GetReport (découverte): %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}
