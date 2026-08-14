// Package handler expose les routes HTTP du service Ansible (Gin).
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/modulops/ansible-service/internal/service"
)

type AnsibleHandler struct {
	svc *service.AnsibleService
}

func New(svc *service.AnsibleService) *AnsibleHandler {
	return &AnsibleHandler{svc: svc}
}

// emptyPaginated est la réponse renvoyée quand Semaphore n'est pas configuré.
func emptyPaginated() gin.H {
	return gin.H{"count": 0, "results": []any{}, "next": nil, "previous": nil}
}

// ── Configuration ────────────────────────────────────────────────────────────

func (h *AnsibleHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.GetConfig(c.Request.Context()))
}

func (h *AnsibleHandler) SetConfig(c *gin.Context) {
	var payload struct {
		SemaphoreURL       string `json:"semaphore_url"`
		Token              string `json:"token"`
		SemaphoreProjectID any    `json:"semaphore_project_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	projectID := 1
	switch v := payload.SemaphoreProjectID.(type) {
	case float64:
		projectID = int(v)
	case string:
		if id, err := strconv.Atoi(v); err == nil {
			projectID = id
		}
	}

	result, err := h.svc.SetConfig(c.Request.Context(), payload.SemaphoreURL, payload.Token, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── Jobs ─────────────────────────────────────────────────────────────────────

func (h *AnsibleHandler) GetJobs(c *gin.Context) {
	pageSize := queryInt(c, "page_size", 20)
	result, err := h.svc.GetJobs(c.Request.Context(), pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusOK, emptyPaginated())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnsibleHandler) GetJobHistory(c *gin.Context) {
	pageSize := queryInt(c, "page_size", 50)
	result, err := h.svc.GetJobHistory(c.Request.Context(), pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusOK, emptyPaginated())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnsibleHandler) GetJob(c *gin.Context) {
	jobID, ok := paramInt(c, "job_id")
	if !ok {
		return
	}
	includeStdout := c.Query("include_stdout") == "true"
	job, err := h.svc.GetJob(c.Request.Context(), jobID, includeStdout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Job " + c.Param("job_id") + " non trouvé"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// ── Inventaires ──────────────────────────────────────────────────────────────

func (h *AnsibleHandler) GetInventories(c *gin.Context) {
	result, err := h.svc.GetInventories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusOK, emptyPaginated())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnsibleHandler) GetInventory(c *gin.Context) {
	inventoryID, ok := paramInt(c, "inventory_id")
	if !ok {
		return
	}
	inventory, err := h.svc.GetInventory(c.Request.Context(), inventoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if inventory == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Inventaire " + c.Param("inventory_id") + " non trouvé"})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (h *AnsibleHandler) GetInventoryHosts(c *gin.Context) {
	inventoryID, ok := paramInt(c, "inventory_id")
	if !ok {
		return
	}
	result, err := h.svc.GetInventoryHosts(c.Request.Context(), inventoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusOK, emptyPaginated())
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── Templates ────────────────────────────────────────────────────────────────

func (h *AnsibleHandler) GetJobTemplates(c *gin.Context) {
	result, err := h.svc.GetJobTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusOK, emptyPaginated())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnsibleHandler) GetJobTemplate(c *gin.Context) {
	templateID, ok := paramInt(c, "template_id")
	if !ok {
		return
	}
	template, err := h.svc.GetJobTemplate(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if template == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Template " + c.Param("template_id") + " non trouvé"})
		return
	}
	c.JSON(http.StatusOK, template)
}

func (h *AnsibleHandler) GetJobTemplatePlaybook(c *gin.Context) {
	templateID, ok := paramInt(c, "template_id")
	if !ok {
		return
	}
	source, err := h.svc.GetTemplatePlaybookSource(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if source == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Playbook introuvable pour ce template"})
		return
	}
	c.JSON(http.StatusOK, source)
}

func (h *AnsibleHandler) LaunchJobTemplate(c *gin.Context) {
	templateID, ok := paramInt(c, "template_id")
	if !ok {
		return
	}
	var extraVars map[string]any
	_ = c.ShouldBindJSON(&extraVars) // corps optionnel

	job, err := h.svc.LaunchJobTemplate(c.Request.Context(), templateID, extraVars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Impossible de lancer le template " + c.Param("template_id")})
		return
	}
	// Champ "job" = id, au format de la réponse de lancement AWX attendue par le frontend.
	job["job"] = job["id"]
	jobsLaunchedTotal.WithLabelValues(strconv.Itoa(templateID)).Inc()
	c.JSON(http.StatusOK, job)
}

// ── Projets et playbooks ─────────────────────────────────────────────────────

func (h *AnsibleHandler) GetProjects(c *gin.Context) {
	result, err := h.svc.GetProjects(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Service Ansible non disponible. Vérifiez la configuration."})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnsibleHandler) GetProject(c *gin.Context) {
	projectID, ok := paramInt(c, "project_id")
	if !ok {
		return
	}
	project, err := h.svc.GetProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Projet " + c.Param("project_id") + " non trouvé"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *AnsibleHandler) GetProjectPlaybooks(c *gin.Context) {
	projectID, ok := paramInt(c, "project_id")
	if !ok {
		return
	}
	// Semaphore ne liste pas les playbooks d'un projet : ils sont attachés aux
	// templates. On retourne une liste vide, comme le faisait le client Python.
	c.JSON(http.StatusOK, gin.H{"project_id": projectID, "playbooks": []any{}})
}

func (h *AnsibleHandler) AnalyzePlaybook(c *gin.Context) {
	var payload struct {
		PlaybookContent string `json:"playbook_content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h.svc.AnalyzePlaybook(payload.PlaybookContent))
}

// ── Stubs de compatibilité AWX (credentials, organisations) ──────────────────

func (h *AnsibleHandler) GetCredentials(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.GetCredentials(c.Request.Context()))
}

func (h *AnsibleHandler) GetOrganizations(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.GetOrganizations(c.Request.Context()))
}

// GetCredential : la liste des credentials est vide avec Semaphore, un id précis est donc introuvable.
func (h *AnsibleHandler) GetCredential(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"detail": "Credential " + c.Param("credential_id") + " non trouvé"})
}

// GetOrganization : seule l'organisation par défaut existe côté Semaphore.
func (h *AnsibleHandler) GetOrganization(c *gin.Context) {
	if c.Param("organization_id") == "1" {
		c.JSON(http.StatusOK, gin.H{"id": 1, "name": "Default"})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"detail": "Organisation " + c.Param("organization_id") + " non trouvée"})
}

// NotSupported répond 501 pour les opérations d'écriture AWX sans équivalent
// Semaphore (credentials et organisations sont gérés dans Semaphore lui-même).
func (h *AnsibleHandler) NotSupported(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"detail": "Opération non supportée avec le backend Semaphore : gérez cette ressource dans Semaphore.",
	})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func queryInt(c *gin.Context, name string, fallback int) int {
	if raw := c.Query(name); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 1 && v <= 100 {
			return v
		}
	}
	return fallback
}

func paramInt(c *gin.Context, name string) (int, bool) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": name + " doit être un entier"})
		return 0, false
	}
	return v, true
}
