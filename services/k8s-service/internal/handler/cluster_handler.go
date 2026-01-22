package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/models"
	"github.com/modulops/k8s-service/internal/service"
)

// ClusterHandler gère les requêtes HTTP liées aux clusters Kubernetes.
type ClusterHandler struct {
	clusterService *service.ClusterService
	cfg            *config.Config
}

// NewClusterHandler crée un nouveau handler pour les clusters.
func NewClusterHandler(clusterService *service.ClusterService, cfg *config.Config) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterService,
		cfg:            cfg,
	}
}

// CreateClusterRequest représente une requête de création de cluster.
type CreateClusterRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	Kubeconfig  string `json:"kubeconfig" binding:"required"`
	IsActive    bool   `json:"is_active,omitempty"`
}

// CreateCluster crée un nouveau cluster.
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster := &models.Cluster{
		Name:        req.Name,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		Kubeconfig:  req.Kubeconfig,
		IsActive:    req.IsActive,
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

// ListClusters retourne la liste de tous les clusters.
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	ctx := c.Request.Context()

	clusters, err := h.clusterService.ListClusters(ctx)
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
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	Kubeconfig  string `json:"kubeconfig,omitempty"`
	IsActive    bool   `json:"is_active,omitempty"`
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

	cluster := &models.Cluster{
		Name:        req.Name,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		Kubeconfig:  req.Kubeconfig,
		IsActive:    req.IsActive,
	}

	updated, err := h.clusterService.UpdateCluster(ctx, id, cluster)
	if err != nil {
		log.Printf("Erreur UpdateCluster pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

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
