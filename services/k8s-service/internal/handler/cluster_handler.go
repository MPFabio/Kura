package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/models"
	"github.com/modulops/k8s-service/internal/service"
)

// invalidator est implémenté par les handlers qui mettent en cache un client
// Kubernetes lié au cluster actif (K8sHandler, TerminalHandler).
type invalidator interface {
	Invalidate()
}

// ClusterHandler gère les requêtes HTTP liées aux clusters Kubernetes.
type ClusterHandler struct {
	clusterService *service.ClusterService
	cfg            *config.Config
	invalidators   []invalidator
}

// NewClusterHandler crée un nouveau handler pour les clusters.
func NewClusterHandler(clusterService *service.ClusterService, cfg *config.Config) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterService,
		cfg:            cfg,
	}
}

// SetInvalidators enregistre les handlers dont le client Kubernetes mis en
// cache doit être invalidé lorsqu'un cluster est modifié, activé ou supprimé.
func (h *ClusterHandler) SetInvalidators(invalidators ...invalidator) {
	h.invalidators = invalidators
}

// invalidateClients force la recréation des clients Kubernetes mis en cache
// par les autres handlers (K8sHandler, TerminalHandler).
func (h *ClusterHandler) invalidateClients() {
	for _, inv := range h.invalidators {
		inv.Invalidate()
	}
}

// CreateClusterRequest représente une requête de création de cluster.
type CreateClusterRequest struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description,omitempty"`
	Endpoint         string `json:"endpoint,omitempty"`
	Kubeconfig       string `json:"kubeconfig" binding:"required"`
	ProjectID        string `json:"project_id" binding:"required"`
	ClusterType      string `json:"cluster_type,omitempty"`      // generic | gke | aks | eks | proxmox
	CloudCredentials string `json:"cloud_credentials,omitempty"` // JSON : clé GCP (gcp_sa_key), Azure (tenant_id, client_id, client_secret), AWS (access_key_id, secret_access_key)
	IsActive         bool   `json:"is_active,omitempty"`
}

// CreateCluster crée un nouveau cluster.
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation des endpoints en production
	isProduction := h.cfg.Environment == "production"
	if isProduction && req.Endpoint != "" {
		if strings.Contains(req.Endpoint, "127.0.0.1") ||
			strings.Contains(req.Endpoint, "localhost") ||
			strings.Contains(req.Endpoint, "host.docker.internal") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "endpoints locaux (127.0.0.1, localhost, host.docker.internal) interdits en production pour des raisons de sécurité",
			})
			return
		}
	}

	clusterType := req.ClusterType
	if clusterType == "" {
		clusterType = models.ClusterTypeGeneric
	}
	cluster := &models.Cluster{
		Name:             req.Name,
		Description:      req.Description,
		Endpoint:         req.Endpoint,
		Kubeconfig:       req.Kubeconfig,
		ProjectID:        req.ProjectID,
		ClusterType:      clusterType,
		CloudCredentials: req.CloudCredentials,
		IsActive:         req.IsActive,
	}

	created, err := h.clusterService.CreateCluster(ctx, cluster)
	if err != nil {
		log.Printf("Erreur CreateCluster: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Tester automatiquement la connexion après la création
	if created.Kubeconfig != "" {
		testResult, testErr := h.clusterService.TestClusterConnection(ctx, created)
		if testErr == nil && testResult != nil {
			log.Printf("Test de connexion pour cluster %s: Connected=%v, Error=%v", created.ID, testResult.Connected, testResult.Error)
			if !testResult.Connected && testResult.Error != "" {
				// Ne pas échouer la création, mais loguer l'avertissement
				log.Printf("⚠️  Avertissement: Le cluster %s a été créé mais la connexion a échoué: %s", created.ID, testResult.Error)
			}
		}
	}

	c.JSON(http.StatusCreated, created)
}

// GetCluster retourne un cluster par son ID.
func (h *ClusterHandler) GetCluster(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	cluster, err := h.clusterService.GetCluster(ctx, id)
	if err != nil {
		log.Printf("Erreur GetCluster pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

// GetClusterKubeconfig retourne le kubeconfig brut d'un cluster (route interne).
// Protégée par un secret partagé (header X-Internal-Token) car non authentifiée par JWT.
// Utilisé par les playbooks Ansible (Semaphore) pour obtenir un accès au cluster au moment de l'exécution.
func (h *ClusterHandler) GetClusterKubeconfig(c *gin.Context) {
	if h.cfg.InternalAPISecret == "" || c.GetHeader("X-Internal-Token") != h.cfg.InternalAPISecret {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non autorisé"})
		return
	}

	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	cluster, err := h.clusterService.GetCluster(ctx, id)
	if err != nil {
		log.Printf("Erreur GetClusterKubeconfig pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if cluster.Kubeconfig == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "kubeconfig manquant pour ce cluster"})
		return
	}

	kubeconfig, err := h.clusterService.GetPortableKubeconfig(ctx, cluster)
	if err != nil {
		log.Printf("Erreur GetPortableKubeconfig pour id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"kubeconfig": kubeconfig})
}

// ListClusters retourne la liste de tous les clusters.
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	ctx := c.Request.Context()

	// Récupérer project_id depuis les query params
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id requis"})
		return
	}

	clusters, err := h.clusterService.ListClustersByProject(ctx, projectID)
	if err != nil {
		log.Printf("Erreur ListClusters: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": clusters,
	})
}

// UpdateClusterRequest représente une requête de mise à jour de cluster.
type UpdateClusterRequest struct {
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Endpoint         string `json:"endpoint,omitempty"`
	Kubeconfig       string `json:"kubeconfig,omitempty"`
	ClusterType      string `json:"cluster_type,omitempty"`
	CloudCredentials string `json:"cloud_credentials,omitempty"`
	IsActive         bool   `json:"is_active,omitempty"`
}

// UpdateCluster met à jour un cluster.
func (h *ClusterHandler) UpdateCluster(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	var req UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation des endpoints en production
	isProduction := h.cfg.Environment == "production"
	if isProduction && req.Endpoint != "" {
		if strings.Contains(req.Endpoint, "127.0.0.1") ||
			strings.Contains(req.Endpoint, "localhost") ||
			strings.Contains(req.Endpoint, "host.docker.internal") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "endpoints locaux (127.0.0.1, localhost, host.docker.internal) interdits en production pour des raisons de sécurité",
			})
			return
		}
	}

	cluster := &models.Cluster{
		Name:             req.Name,
		Description:      req.Description,
		Endpoint:         req.Endpoint,
		Kubeconfig:       req.Kubeconfig,
		ClusterType:      req.ClusterType,
		CloudCredentials: req.CloudCredentials,
		IsActive:         req.IsActive,
	}

	updated, err := h.clusterService.UpdateCluster(ctx, id, cluster)
	if err != nil {
		log.Printf("Erreur UpdateCluster pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.invalidateClients()

	c.JSON(http.StatusOK, updated)
}

// DeleteCluster supprime un cluster.
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	if err := h.clusterService.DeleteCluster(ctx, id); err != nil {
		log.Printf("Erreur DeleteCluster pour id %s: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.invalidateClients()

	c.JSON(http.StatusOK, gin.H{"message": "cluster supprimé"})
}

// SetActiveCluster définit le cluster actif.
func (h *ClusterHandler) SetActiveCluster(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	if err := h.clusterService.SetActiveCluster(ctx, id); err != nil {
		log.Printf("Erreur SetActiveCluster pour id %s: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.invalidateClients()

	c.JSON(http.StatusOK, gin.H{"message": "cluster activé"})
}

// GetActiveCluster retourne le cluster actif.
func (h *ClusterHandler) GetActiveCluster(c *gin.Context) {
	ctx := c.Request.Context()

	cluster, err := h.clusterService.GetActiveCluster(ctx)
	if err != nil {
		log.Printf("Erreur GetActiveCluster: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

// TestClusterConnection teste la connexion à un cluster.
func (h *ClusterHandler) TestClusterConnection(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	cluster, err := h.clusterService.GetCluster(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	status := &models.ClusterStatus{
		ClusterID:   cluster.ID,
		Connected:   false,
		LastChecked: time.Now(),
	}

	// Vérifier si le kubeconfig est présent
	if cluster.Kubeconfig == "" {
		status.Error = "kubeconfig manquant"
		c.JSON(http.StatusOK, status)
		return
	}

	// Tester la connexion réelle en créant un client temporaire
	testResult, err := h.clusterService.TestClusterConnection(ctx, cluster)
	if err != nil {
		status.Error = err.Error()
		c.JSON(http.StatusOK, status)
		return
	}

	status.Connected = testResult.Connected
	status.Version = testResult.Version
	status.NodesCount = testResult.NodesCount
	status.Error = testResult.Error

	c.JSON(http.StatusOK, status)
}
