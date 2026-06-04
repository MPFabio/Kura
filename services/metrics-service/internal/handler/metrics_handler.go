package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modulops/metrics-service/internal/service"
)

type MetricsHandler struct {
	svc *service.MetricsService
}

func New(svc *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{svc: svc}
}

// GetHealth godoc
// GET /api/v1/metrics/health
// Retourne l'état up/down + goroutines de chaque service Kura.
func (h *MetricsHandler) GetHealth(c *gin.Context) {
	result, err := h.svc.GetHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetServices godoc
// GET /api/v1/metrics/services
// Retourne les métriques détaillées (goroutines, CPU, mémoire) par service.
func (h *MetricsHandler) GetServices(c *gin.Context) {
	result, err := h.svc.GetServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetOverview godoc
// GET /api/v1/metrics/overview
// Retourne les KPIs globaux : services up/down, goroutines totales, mémoire totale.
func (h *MetricsHandler) GetOverview(c *gin.Context) {
	result, err := h.svc.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
