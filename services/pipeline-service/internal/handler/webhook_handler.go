package handler

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/modulops/pipeline-service/internal/config"
	"github.com/modulops/pipeline-service/internal/metrics"
	"github.com/modulops/pipeline-service/internal/models"
	"github.com/modulops/pipeline-service/internal/service"
)

// WebhookHandler gère les webhooks des providers CI/CD
type WebhookHandler struct {
	svc *service.PipelineService
	cfg *config.Config
}

// NewWebhookHandler crée un nouveau handler de webhooks
func NewWebhookHandler(svc *service.PipelineService, cfg *config.Config) *WebhookHandler {
	return &WebhookHandler{svc: svc, cfg: cfg}
}

// HandleGitHub gère les webhooks GitHub Actions
// POST /api/v1/pipeline/webhooks/github
func (h *WebhookHandler) HandleGitHub(c *gin.Context) {
	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("github", "error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), models.ProviderGitHub, body)
	metrics.WebhookProcessingDuration.WithLabelValues("github").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("github", "error").Inc()
		log.Printf("Erreur webhook GitHub: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	metrics.WebhooksReceivedTotal.WithLabelValues("github", "ok").Inc()
	if run == nil {
		c.JSON(http.StatusOK, gin.H{"message": "événement ignoré"})
		return
	}
	metrics.RunsStoredTotal.WithLabelValues("github", string(run.Status)).Inc()

	c.JSON(http.StatusOK, gin.H{
		"message": "webhook traité",
		"run_id":  run.ID,
		"status":  run.Status,
	})
}

// HandleGitLab gère les webhooks GitLab CI
// POST /api/v1/pipeline/webhooks/gitlab
func (h *WebhookHandler) HandleGitLab(c *gin.Context) {
	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("gitlab", "error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), models.ProviderGitLab, body)
	metrics.WebhookProcessingDuration.WithLabelValues("gitlab").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("gitlab", "error").Inc()
		log.Printf("Erreur webhook GitLab: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	metrics.WebhooksReceivedTotal.WithLabelValues("gitlab", "ok").Inc()
	if run == nil {
		c.JSON(http.StatusOK, gin.H{"message": "événement ignoré"})
		return
	}
	metrics.RunsStoredTotal.WithLabelValues("gitlab", string(run.Status)).Inc()

	c.JSON(http.StatusOK, gin.H{
		"message": "webhook traité",
		"run_id":  run.ID,
		"status":  run.Status,
	})
}

// HandleJenkins gère les webhooks Jenkins
// POST /api/v1/pipeline/webhooks/jenkins
func (h *WebhookHandler) HandleJenkins(c *gin.Context) {
	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("jenkins", "error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), models.ProviderJenkins, body)
	metrics.WebhookProcessingDuration.WithLabelValues("jenkins").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("jenkins", "error").Inc()
		log.Printf("Erreur webhook Jenkins: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	metrics.WebhooksReceivedTotal.WithLabelValues("jenkins", "ok").Inc()
	if run == nil {
		c.JSON(http.StatusOK, gin.H{"message": "événement ignoré"})
		return
	}
	metrics.RunsStoredTotal.WithLabelValues("jenkins", string(run.Status)).Inc()

	c.JSON(http.StatusOK, gin.H{
		"message": "webhook traité",
		"run_id":  run.ID,
		"status":  run.Status,
	})
}

// HandleGeneric détecte automatiquement le provider et route le webhook
// POST /api/v1/pipeline/webhooks
func (h *WebhookHandler) HandleGeneric(c *gin.Context) {
	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	xGitHub := c.GetHeader("X-GitHub-Event")
	xGitLab := c.GetHeader("X-Gitlab-Event")

	provider := h.svc.DetectProvider(xGitHub, xGitLab)
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "impossible de détecter le provider (headers X-GitHub-Event ou X-Gitlab-Event requis). Utilisez /webhooks/github, /webhooks/gitlab ou /webhooks/jenkins"})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), provider, body)
	metrics.WebhookProcessingDuration.WithLabelValues(string(provider)).Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues(string(provider), "error").Inc()
		log.Printf("Erreur webhook %s: %v", provider, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	metrics.WebhooksReceivedTotal.WithLabelValues(string(provider), "ok").Inc()
	if run == nil {
		c.JSON(http.StatusOK, gin.H{"message": "événement ignoré"})
		return
	}
	metrics.RunsStoredTotal.WithLabelValues(string(provider), string(run.Status)).Inc()

	c.JSON(http.StatusOK, gin.H{
		"message":  "webhook traité",
		"run_id":   run.ID,
		"status":   run.Status,
		"provider": provider,
	})
}
