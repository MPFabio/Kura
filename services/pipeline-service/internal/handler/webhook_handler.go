package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"
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

// verifyGitHubSignature valide la signature HMAC-SHA256 envoyée par GitHub.
// GitHub inclut l'en-tête X-Hub-Signature-256: sha256=<hex>
// Si le secret n'est pas configuré, la vérification est ignorée (mode développement).
func (h *WebhookHandler) verifyGitHubSignature(body []byte, signature string) bool {
	if h.cfg.GitHubWebhookSecret == "" {
		return true
	}
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	expectedMAC := hmac.New(sha256.New, []byte(h.cfg.GitHubWebhookSecret))
	expectedMAC.Write(body)
	expected := "sha256=" + hex.EncodeToString(expectedMAC.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// verifyGitLabToken valide le token secret envoyé par GitLab.
// GitLab inclut l'en-tête X-Gitlab-Token: <secret>
func (h *WebhookHandler) verifyGitLabToken(token string) bool {
	if h.cfg.GitLabWebhookSecret == "" {
		return true
	}
	return hmac.Equal([]byte(token), []byte(h.cfg.GitLabWebhookSecret))
}

// verifyForgejoSignature valide la signature HMAC-SHA256 envoyée par Forgejo.
// Forgejo inclut l'en-tête X-Hub-Signature-256: sha256=<hex> (format identique à GitHub).
// Si le secret n'est pas configuré, la vérification est ignorée (mode développement).
func (h *WebhookHandler) verifyForgejoSignature(body []byte, signature string) bool {
	if h.cfg.ForgejoWebhookSecret == "" {
		return true
	}
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	expectedMAC := hmac.New(sha256.New, []byte(h.cfg.ForgejoWebhookSecret))
	expectedMAC.Write(body)
	expected := "sha256=" + hex.EncodeToString(expectedMAC.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// HandleGitHub gère les webhooks GitHub Actions
// POST /api/v1/pipeline/webhooks/github
func (h *WebhookHandler) HandleGitHub(c *gin.Context) {
	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("github", "error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "lecture du corps impossible"})
		return
	}

	if !h.verifyGitHubSignature(body, c.GetHeader("X-Hub-Signature-256")) {
		metrics.WebhooksReceivedTotal.WithLabelValues("github", "unauthorized").Inc()
		log.Printf("⚠️ Webhook GitHub rejeté : signature HMAC invalide")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "signature invalide"})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), models.ProviderGitHub, body)
	metrics.WebhookProcessingDuration.WithLabelValues("github").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("github", "error").Inc()
		log.Printf("Erreur webhook GitHub: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "traitement du webhook échoué"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "lecture du corps impossible"})
		return
	}

	if !h.verifyGitLabToken(c.GetHeader("X-Gitlab-Token")) {
		metrics.WebhooksReceivedTotal.WithLabelValues("gitlab", "unauthorized").Inc()
		log.Printf("⚠️ Webhook GitLab rejeté : token invalide")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalide"})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), models.ProviderGitLab, body)
	metrics.WebhookProcessingDuration.WithLabelValues("gitlab").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("gitlab", "error").Inc()
		log.Printf("Erreur webhook GitLab: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "traitement du webhook échoué"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "lecture du corps impossible"})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), models.ProviderJenkins, body)
	metrics.WebhookProcessingDuration.WithLabelValues("jenkins").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("jenkins", "error").Inc()
		log.Printf("Erreur webhook Jenkins: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "traitement du webhook échoué"})
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

// HandleForgejo gère les webhooks Forgejo Actions
// POST /api/v1/pipeline/webhooks/forgejo
func (h *WebhookHandler) HandleForgejo(c *gin.Context) {
	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("forgejo", "error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "lecture du corps impossible"})
		return
	}

	if !h.verifyForgejoSignature(body, c.GetHeader("X-Hub-Signature-256")) {
		metrics.WebhooksReceivedTotal.WithLabelValues("forgejo", "unauthorized").Inc()
		log.Printf("⚠️ Webhook Forgejo rejeté : signature HMAC invalide")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "signature invalide"})
		return
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), models.ProviderForgejo, body)
	metrics.WebhookProcessingDuration.WithLabelValues("forgejo").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues("forgejo", "error").Inc()
		log.Printf("Erreur webhook Forgejo: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "traitement du webhook échoué"})
		return
	}
	metrics.WebhooksReceivedTotal.WithLabelValues("forgejo", "ok").Inc()
	if run == nil {
		c.JSON(http.StatusOK, gin.H{"message": "événement ignoré"})
		return
	}
	metrics.RunsStoredTotal.WithLabelValues("forgejo", string(run.Status)).Inc()

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "lecture du corps impossible"})
		return
	}

	xGitHub := c.GetHeader("X-GitHub-Event")
	xGitLab := c.GetHeader("X-Gitlab-Event")
	xForgejo := c.GetHeader("X-Forgejo-Event")

	provider := h.svc.DetectProvider(xGitHub, xGitLab, xForgejo)
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "impossible de détecter le provider (headers X-GitHub-Event, X-Gitlab-Event ou X-Forgejo-Event requis). Utilisez /webhooks/github, /webhooks/gitlab, /webhooks/forgejo ou /webhooks/jenkins"})
		return
	}

	// Validation de la signature selon le provider détecté
	if provider == models.ProviderGitHub {
		if !h.verifyGitHubSignature(body, c.GetHeader("X-Hub-Signature-256")) {
			metrics.WebhooksReceivedTotal.WithLabelValues("github", "unauthorized").Inc()
			log.Printf("⚠️ Webhook GitHub (generic) rejeté : signature HMAC invalide")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "signature invalide"})
			return
		}
	} else if provider == models.ProviderGitLab {
		if !h.verifyGitLabToken(c.GetHeader("X-Gitlab-Token")) {
			metrics.WebhooksReceivedTotal.WithLabelValues("gitlab", "unauthorized").Inc()
			log.Printf("⚠️ Webhook GitLab (generic) rejeté : token invalide")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalide"})
			return
		}
	} else if provider == models.ProviderForgejo {
		if !h.verifyForgejoSignature(body, c.GetHeader("X-Hub-Signature-256")) {
			metrics.WebhooksReceivedTotal.WithLabelValues("forgejo", "unauthorized").Inc()
			log.Printf("⚠️ Webhook Forgejo (generic) rejeté : signature HMAC invalide")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "signature invalide"})
			return
		}
	}

	run, err := h.svc.ProcessWebhook(c.Request.Context(), provider, body)
	metrics.WebhookProcessingDuration.WithLabelValues(string(provider)).Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.WebhooksReceivedTotal.WithLabelValues(string(provider), "error").Inc()
		log.Printf("Erreur webhook %s: %v", provider, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "traitement du webhook échoué"})
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
