package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/service"
)

// WebhookHandler gère les webhooks pour les mises à jour d'états.
type WebhookHandler struct {
	syncService *service.SyncService
	cfg         *config.Config
}

// NewWebhookHandler crée un nouveau handler de webhooks.
func NewWebhookHandler(syncService *service.SyncService, cfg *config.Config) *WebhookHandler {
	return &WebhookHandler{
		syncService: syncService,
		cfg:         cfg,
	}
}

// HandleStateUpdated gère un webhook indiquant qu'un état a été mis à jour.
func (h *WebhookHandler) HandleStateUpdated(c *gin.Context) {
	ctx := c.Request.Context()

	var payload struct {
		SourceID  string `json:"source_id" binding:"required"`
		Bucket    string `json:"bucket,omitempty"`
		Key       string `json:"key,omitempty"`
		VersionID string `json:"version_id,omitempty"`
		ETag      string `json:"etag,omitempty"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Déclencher une synchronisation immédiate
	job, err := h.syncService.SyncState(ctx, payload.SourceID)
	if err != nil {
		log.Printf("Erreur lors de la synchronisation via webhook pour source %s: %v", payload.SourceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Synchronisation déclenchée via webhook pour source %s (job: %s)", payload.SourceID, job.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "synchronisation déclenchée",
		"job_id":  job.ID,
	})
}

// HandleS3Event gère un événement S3 (compatible avec les notifications S3).
func (h *WebhookHandler) HandleS3Event(c *gin.Context) {
	ctx := c.Request.Context()

	var payload struct {
		Records []struct {
			S3 struct {
				Bucket struct {
					Name string `json:"name"`
				} `json:"bucket"`
				Object struct {
					Key       string `json:"key"`
					ETag      string `json:"eTag"`
					Size      int64  `json:"size"`
					VersionID string `json:"versionId,omitempty"`
				} `json:"object"`
			} `json:"s3"`
		} `json:"Records"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Traiter chaque événement
	for _, record := range payload.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		// Chercher les sources qui correspondent à ce bucket/key
		sources, err := h.syncService.ListSources(ctx)
		if err != nil {
			log.Printf("Erreur lors de la récupération des sources: %v", err)
			continue
		}

		for _, source := range sources {
			if source.Type == "s3" &&
				source.Config.S3Bucket == bucket &&
				source.Config.S3Key == key &&
				source.Enabled {
				// Déclencher une synchronisation
				job, err := h.syncService.SyncState(ctx, source.ID)
				if err != nil {
					log.Printf("Erreur lors de la synchronisation pour source %s: %v", source.ID, err)
				} else {
					log.Printf("✅ Synchronisation déclenchée via événement S3 pour source %s (job: %s)", source.ID, job.ID)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "événements traités"})
}
