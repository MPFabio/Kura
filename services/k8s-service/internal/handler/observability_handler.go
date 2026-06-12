package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/modulops/k8s-service/internal/service"
)

// ObservabilityHandler gère les requêtes HTTP liées à l'observabilité du
// projet (Prometheus/Loki/Tempo déployés dans le cluster du client).
type ObservabilityHandler struct {
	svc *service.ObservabilityService
}

// NewObservabilityHandler crée un nouveau handler pour l'observabilité projet.
func NewObservabilityHandler(svc *service.ObservabilityService) *ObservabilityHandler {
	return &ObservabilityHandler{svc: svc}
}

// GetOverview godoc
// GET /api/v1/k8s/observability/overview
func (h *ObservabilityHandler) GetOverview(c *gin.Context) {
	result, err := h.svc.GetOverview(c.Request.Context())
	if err != nil {
		log.Printf("Erreur GetOverview (observabilité projet): %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetServices godoc
// GET /api/v1/k8s/observability/services
func (h *ObservabilityHandler) GetServices(c *gin.Context) {
	result, err := h.svc.GetServiceMetrics(c.Request.Context())
	if err != nil {
		log.Printf("Erreur GetServiceMetrics (observabilité projet): %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetLogs godoc
// GET /api/v1/k8s/observability/logs?service=...&search=...&limit=...
func (h *ObservabilityHandler) GetLogs(c *gin.Context) {
	svc := c.Query("service")
	search := c.Query("search")
	limit, _ := strconv.Atoi(c.Query("limit"))

	result, err := h.svc.GetLogs(c.Request.Context(), svc, search, limit)
	if err != nil {
		log.Printf("Erreur GetLogs (observabilité projet): %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

// GetLogServices godoc
// GET /api/v1/k8s/observability/logs/services
func (h *ObservabilityHandler) GetLogServices(c *gin.Context) {
	result, err := h.svc.GetLogServices(c.Request.Context())
	if err != nil {
		log.Printf("Erreur GetLogServices (observabilité projet): %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

// SearchTraces godoc
// GET /api/v1/k8s/observability/traces?service=...&min_duration_ms=...&limit=...
func (h *ObservabilityHandler) SearchTraces(c *gin.Context) {
	svc := c.Query("service")
	minDuration, _ := strconv.Atoi(c.Query("min_duration_ms"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	result, err := h.svc.SearchTraces(c.Request.Context(), svc, minDuration, limit)
	if err != nil {
		log.Printf("Erreur SearchTraces (observabilité projet): %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

// grafanaProxyPrefix est le préfixe de route retiré avant de transmettre la
// requête au Grafana du cluster client.
const grafanaProxyPrefix = "/api/v1/k8s/observability/grafana"

// ProxyGrafana godoc
// GET /api/v1/k8s/observability/grafana/*path
// Relaie la requête vers le Grafana déployé dans le cluster client (kube-prometheus-stack),
// via un port-forward maintenu ouvert pour le chargement d'un dashboard en iframe.
func (h *ObservabilityHandler) ProxyGrafana(c *gin.Context) {
	proxy, err := h.svc.GrafanaProxyHandler(c.Request.Context())
	if err != nil {
		log.Printf("Erreur ProxyGrafana (observabilité projet): %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, grafanaProxyPrefix)
	if c.Request.URL.Path == "" {
		c.Request.URL.Path = "/"
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// GetTrace godoc
// GET /api/v1/k8s/observability/traces/:traceID
func (h *ObservabilityHandler) GetTrace(c *gin.Context) {
	traceID := c.Param("traceID")

	result, err := h.svc.GetTrace(c.Request.Context(), traceID)
	if err != nil {
		log.Printf("Erreur GetTrace (observabilité projet): %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", result)
}
