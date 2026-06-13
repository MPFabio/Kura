package handler

import (
	"net/http"
	"strconv"

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

// GetPlatformConfig indique si l'observabilité interne de la plateforme Kura
// (santé/metrics de ses propres microservices) est exposée dans cet environnement.
// GET /api/v1/metrics/platform-config
func (h *MetricsHandler) GetPlatformConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"internal_observability_enabled": h.svc.InternalObservabilityEnabled(),
	})
}

// RequireInternalObservability bloque l'accès aux endpoints d'observabilité
// interne de Kura lorsque cette fonctionnalité est désactivée (mode SaaS).
func (h *MetricsHandler) RequireInternalObservability() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.svc.InternalObservabilityEnabled() {
			c.JSON(http.StatusNotFound, gin.H{"error": "observabilité interne désactivée"})
			c.Abort()
			return
		}
		c.Next()
	}
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

// GetLogs godoc
// GET /api/v1/metrics/logs?service=auth-service&search=error&limit=200
// Retourne les dernières lignes de logs depuis Loki, filtrées par service et/ou texte.
func (h *MetricsHandler) GetLogs(c *gin.Context) {
	service := c.Query("service")
	search := c.Query("search")
	limit, _ := strconv.Atoi(c.Query("limit"))

	result, err := h.svc.GetLogs(c.Request.Context(), service, search, limit)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

// GetLogServices godoc
// GET /api/v1/metrics/logs/services
// Retourne la liste des services dont les logs sont disponibles dans Loki.
func (h *MetricsHandler) GetLogServices(c *gin.Context) {
	result, err := h.svc.GetLogServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

// SearchTraces godoc
// GET /api/v1/metrics/traces?service=auth-service&min_duration_ms=100&limit=50
// Retourne les traces récentes depuis Tempo, filtrées par service et/ou durée minimale.
func (h *MetricsHandler) SearchTraces(c *gin.Context) {
	service := c.Query("service")
	minDuration, _ := strconv.Atoi(c.Query("min_duration_ms"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	result, err := h.svc.SearchTraces(c.Request.Context(), service, minDuration, limit)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

// GetTrace godoc
// GET /api/v1/metrics/traces/:traceID
// Retourne le détail complet d'une trace depuis Tempo.
func (h *MetricsHandler) GetTrace(c *gin.Context) {
	traceID := c.Param("traceID")

	result, err := h.svc.GetTrace(c.Request.Context(), traceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", result)
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
