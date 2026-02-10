package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/modulops/pipeline-service/internal/config"
	"github.com/modulops/pipeline-service/internal/models"
	"github.com/modulops/pipeline-service/internal/service"
)

// PipelineHandler gère les requêtes HTTP pour les pipelines
type PipelineHandler struct {
	svc *service.PipelineService
	cfg *config.Config
}

// NewPipelineHandler crée un nouveau handler
func NewPipelineHandler(svc *service.PipelineService, cfg *config.Config) *PipelineHandler {
	return &PipelineHandler{svc: svc, cfg: cfg}
}

// ListRuns liste les exécutions de pipelines
// GET /api/v1/pipeline/runs?provider=github&repository=org/repo&branch=main&limit=50
func (h *PipelineHandler) ListRuns(c *gin.Context) {
	provider := c.Query("provider")
	repository := c.Query("repository")
	branch := c.Query("branch")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	runs, err := h.svc.ListRuns(c.Request.Context(), provider, repository, branch, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"runs":  runs,
		"count": len(runs),
	})
}

// GetRun récupère une exécution par ID
// GET /api/v1/pipeline/runs/:id
func (h *PipelineHandler) GetRun(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	run, err := h.svc.GetRun(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "run non trouvé"})
		return
	}

	c.JSON(http.StatusOK, run)
}

// GetAggregatedStatus récupère le statut agrégé
// GET /api/v1/pipeline/aggregated?provider=github&repository=org/repo&branch=main
func (h *PipelineHandler) GetAggregatedStatus(c *gin.Context) {
	provider := c.Query("provider")
	repository := c.Query("repository")
	branch := c.DefaultQuery("branch", "main")

	if provider == "" || repository == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider et repository requis"})
		return
	}

	agg, err := h.svc.GetAggregatedStatus(c.Request.Context(), provider, repository, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if agg == nil {
		c.JSON(http.StatusOK, gin.H{
			"repository": repository,
			"branch":     branch,
			"provider":   provider,
			"message":    "aucune donnée agrégée",
		})
		return
	}

	c.JSON(http.StatusOK, agg)
}

// ListProviders retourne les providers supportés
// GET /api/v1/pipeline/providers
func (h *PipelineHandler) ListProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"providers": []gin.H{
			{"id": string(models.ProviderGitHub), "name": "GitHub Actions"},
			{"id": string(models.ProviderGitLab), "name": "GitLab CI"},
			{"id": string(models.ProviderJenkins), "name": "Jenkins"},
		},
	})
}

// GetConfig retourne la config actuelle (sans token)
// GET /api/v1/pipeline/config
func (h *PipelineHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// SetConfig enregistre la config (token + repos) depuis l'UI
// POST /api/v1/pipeline/config
func (h *PipelineHandler) SetConfig(c *gin.Context) {
	var body struct {
		GitHubToken *string  `json:"github_token"`
		GitHubRepos []string `json:"github_repos"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format invalide"})
		return
	}

	token := ""
	if body.GitHubToken != nil {
		token = *body.GitHubToken
	}
	repos := body.GitHubRepos

	if err := h.svc.SetConfig(c.Request.Context(), token, repos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cfg, _ := h.svc.GetConfig(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"message": "configuration enregistrée", "config": cfg})
}

// SyncGitHub déclenche une synchronisation manuelle depuis l'API GitHub
// POST /api/v1/pipeline/sync
func (h *PipelineHandler) SyncGitHub(c *gin.Context) {
	count, err := h.svc.SyncFromGitHub(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "synchronisation terminée",
		"runs":    count,
	})
}
