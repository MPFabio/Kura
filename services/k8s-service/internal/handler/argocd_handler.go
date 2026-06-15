package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/modulops/k8s-service/internal/models"
	"github.com/modulops/k8s-service/internal/service"
)

// ArgoCDHandler gère les requêtes HTTP liées à ArgoCD.
type ArgoCDHandler struct {
	svc         *service.ArgoCDService
	helmCatalog *service.HelmCatalogService
}

// NewArgoCDHandler crée un nouveau handler pour ArgoCD.
func NewArgoCDHandler(svc *service.ArgoCDService, helmCatalog *service.HelmCatalogService) *ArgoCDHandler {
	return &ArgoCDHandler{svc: svc, helmCatalog: helmCatalog}
}

// InstallArgoCD installe ArgoCD sur le cluster actif, puis amorce son auto-gestion GitOps
// (commit du manifest argo-cd dans le dépôt GitOps du projet, sur la branche choisie).
func (h *ArgoCDHandler) InstallArgoCD(c *gin.Context) {
	var req models.InstallArgoCDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// L'installation d'ArgoCD (application du manifest officiel + commit GitOps) peut
	// prendre plus de temps que le délai d'attente HTTP du client : on la détache du
	// contexte de la requête pour qu'une annulation côté client n'interrompe pas
	// l'opération en plein milieu, ce qui laisserait le cluster dans un état partiel.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := h.svc.Install(ctx, c.GetHeader("Authorization"), req.Branch, req.CreateBranchFrom); err != nil {
		log.Printf("Erreur InstallArgoCD: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetGitOpsBranches renvoie les informations (URL de clone, nom complet, branches) du
// dépôt GitOps du projet, pour peupler le sélecteur de branche côté frontend.
func (h *ArgoCDHandler) GetGitOpsBranches(c *gin.Context) {
	info, err := h.svc.GetGitOpsInfo(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		log.Printf("Erreur GetGitOpsBranches: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetStatus retourne l'état d'installation et de disponibilité d'ArgoCD.
func (h *ArgoCDHandler) GetStatus(c *gin.Context) {
	status, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		log.Printf("Erreur GetStatus ArgoCD: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// ListApplications liste les Applications ArgoCD.
func (h *ArgoCDHandler) ListApplications(c *gin.Context) {
	apps, err := h.svc.ListApplications(c.Request.Context())
	if err != nil {
		log.Printf("Erreur ListApplications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": apps})
}

// GetApplication retourne le détail d'une Application ArgoCD.
func (h *ArgoCDHandler) GetApplication(c *gin.Context) {
	name := c.Param("name")
	app, err := h.svc.GetApplication(c.Request.Context(), name)
	if err != nil {
		log.Printf("Erreur GetApplication: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
}

// CreateApplication crée une nouvelle Application ArgoCD.
func (h *ArgoCDHandler) CreateApplication(c *gin.Context) {
	var req models.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SourceType == "helm" {
		if req.Chart == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "chart est requis pour une source Helm"})
			return
		}
	} else if req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path est requis pour une source Git"})
		return
	}

	app, err := h.svc.CreateApplicationViaGitOps(c.Request.Context(), c.GetHeader("Authorization"), &req)
	if err != nil {
		log.Printf("Erreur CreateApplication: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, app)
}

// SyncApplication déclenche une synchronisation d'une Application ArgoCD.
func (h *ArgoCDHandler) SyncApplication(c *gin.Context) {
	name := c.Param("name")
	prune, _ := strconv.ParseBool(c.Query("prune"))

	if err := h.svc.SyncApplication(c.Request.Context(), name, prune); err != nil {
		log.Printf("Erreur SyncApplication: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// UpdateApplicationValues remplace les values Helm d'une Application ArgoCD et resynchronise.
func (h *ArgoCDHandler) UpdateApplicationValues(c *gin.Context) {
	name := c.Param("name")

	var req models.UpdateValuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateApplicationValues(c.Request.Context(), name, req.Values); err != nil {
		log.Printf("Erreur UpdateApplicationValues: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RefreshApplication force le rafraîchissement de l'état d'une Application ArgoCD.
func (h *ArgoCDHandler) RefreshApplication(c *gin.Context) {
	name := c.Param("name")

	if err := h.svc.RefreshApplication(c.Request.Context(), name); err != nil {
		log.Printf("Erreur RefreshApplication: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SearchHelmCharts recherche des charts Helm dans le catalogue ArtifactHub.
func (h *ArgoCDHandler) SearchHelmCharts(c *gin.Context) {
	query := c.Query("q")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	items, err := h.helmCatalog.SearchCharts(c.Request.Context(), query, page)
	if err != nil {
		log.Printf("Erreur SearchHelmCharts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// RollbackApplication revient à une révision précédente de l'historique d'une Application ArgoCD.
func (h *ArgoCDHandler) RollbackApplication(c *gin.Context) {
	name := c.Param("name")

	var req models.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RollbackApplication(c.Request.Context(), name, req.ID); err != nil {
		log.Printf("Erreur RollbackApplication: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeleteApplication supprime une Application ArgoCD et ses ressources associées.
func (h *ArgoCDHandler) DeleteApplication(c *gin.Context) {
	name := c.Param("name")

	if err := h.svc.DeleteApplication(c.Request.Context(), name); err != nil {
		log.Printf("Erreur DeleteApplication: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
