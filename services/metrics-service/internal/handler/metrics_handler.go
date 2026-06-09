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

// GetConfig retourne la configuration actuelle (URLs Prometheus/Grafana).
// GET /api/v1/metrics/config
func (h *MetricsHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// SetConfig met à jour les URLs Prometheus et/ou Grafana.
// POST /api/v1/metrics/config
// Body: { "prometheus_url": "...", "grafana_url": "..." }
func (h *MetricsHandler) SetConfig(c *gin.Context) {
	var body struct {
		PrometheusURL string `json:"prometheus_url"`
		GrafanaURL    string `json:"grafana_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetConfig(c.Request.Context(), body.PrometheusURL, body.GrafanaURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "configuration mise à jour"})
}
