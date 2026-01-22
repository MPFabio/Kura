package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/models"
	"github.com/modulops/terraform-service/internal/s3"
	"github.com/modulops/terraform-service/internal/service"
)

// SourceHandler gère les requêtes HTTP liées aux sources de synchronisation.
type SourceHandler struct {
	syncService *service.SyncService
	cfg         *config.Config
}

// NewSourceHandler crée un nouveau handler de sources.
func NewSourceHandler(syncService *service.SyncService, cfg *config.Config) *SourceHandler {
	return &SourceHandler{
		syncService: syncService,
		cfg:         cfg,
	}
}

// AddSource ajoute une nouvelle source de synchronisation.
func (h *SourceHandler) AddSource(c *gin.Context) {
	ctx := c.Request.Context()

	var source models.StateSource
	if err := c.ShouldBindJSON(&source); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation basique
	if source.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type de source requis"})
		return
	}

	if source.StateFileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state_file_id requis"})
		return
	}

	// Validation spécifique selon le type
	if source.Type == "s3" {
		if source.Config.S3Bucket == "" || source.Config.S3Key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "s3_bucket et s3_key requis pour une source S3"})
			return
		}
		// Nettoyer les espaces
		source.Config.S3Bucket = strings.TrimSpace(source.Config.S3Bucket)
		source.Config.S3Key = strings.TrimSpace(source.Config.S3Key)
		if source.Config.S3Region == "" {
			source.Config.S3Region = "us-east-1" // Valeur par défaut
		} else {
			source.Config.S3Region = strings.TrimSpace(source.Config.S3Region)
		}
		if source.Config.S3Endpoint != "" {
			source.Config.S3Endpoint = strings.TrimSpace(source.Config.S3Endpoint)
		}
	} else if source.Type == "gcp" {
		// Nettoyer les espaces pour GCP
		source.Config.GCPBucket = strings.TrimSpace(source.Config.GCPBucket)
		source.Config.GCPObjectName = strings.TrimSpace(source.Config.GCPObjectName)
	} else if source.Type == "azure" {
		// Nettoyer les espaces pour Azure
		source.Config.AzureAccountName = strings.TrimSpace(source.Config.AzureAccountName)
		source.Config.AzureContainer = strings.TrimSpace(source.Config.AzureContainer)
		source.Config.AzureBlobName = strings.TrimSpace(source.Config.AzureBlobName)
	}

	createdSource, err := h.syncService.AddSource(ctx, &source)
	if err != nil {
		log.Printf("Erreur AddSource: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, createdSource)
}

// ListSources retourne toutes les sources.
func (h *SourceHandler) ListSources(c *gin.Context) {
	ctx := c.Request.Context()

	sources, err := h.syncService.ListSources(ctx)
	if err != nil {
		log.Printf("Erreur ListSources: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": sources,
	})
}

// GetSource retourne une source par son ID.
func (h *SourceHandler) GetSource(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	source, err := h.syncService.GetSource(ctx, id)
	if err != nil {
		log.Printf("Erreur GetSource pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, source)
}

// SyncSource synchronise un état depuis sa source.
func (h *SourceHandler) SyncSource(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	job, err := h.syncService.SyncState(ctx, id)
	if err != nil {
		log.Printf("Erreur SyncSource pour id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// UpdateSource met à jour une source de synchronisation.
func (h *SourceHandler) UpdateSource(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	var source models.StateSource
	if err := c.ShouldBindJSON(&source); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation basique
	if source.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type de source requis"})
		return
	}

	updatedSource, err := h.syncService.UpdateSource(ctx, id, &source)
	if err != nil {
		log.Printf("Erreur UpdateSource pour id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedSource)
}

// DeleteSource supprime une source de synchronisation.
func (h *SourceHandler) DeleteSource(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	if err := h.syncService.DeleteSource(ctx, id); err != nil {
		log.Printf("Erreur DeleteSource pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "source supprimée avec succès"})
}

// TestS3Connection teste la connexion à S3.
func (h *SourceHandler) TestS3Connection(c *gin.Context) {
	var req struct {
		Bucket            string `json:"bucket" binding:"required"`
		Region            string `json:"region" binding:"required"`
		Endpoint          string `json:"endpoint,omitempty"`
		AWSAccessKeyID    string `json:"aws_access_key_id,omitempty"`
		AWSSecretAccessKey string `json:"aws_secret_access_key,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Créer un client S3 temporaire pour tester
	ctx := c.Request.Context()
	s3Client, err := s3.NewClient(req.Region, req.Endpoint, req.AWSAccessKeyID, req.AWSSecretAccessKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("erreur lors de la création du client S3: %v", err)})
		return
	}

	// Tester la connexion
	if err := s3Client.TestConnection(ctx, req.Bucket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "connexion S3 réussie"})
}
