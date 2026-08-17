package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modulops/metrics-service/internal/service"
)

// IncidentHandler reçoit les webhooks d'Alertmanager.
type IncidentHandler struct {
	svc *service.IncidentService
}

// NewIncidentHandler crée le handler.
func NewIncidentHandler(svc *service.IncidentService) *IncidentHandler {
	return &IncidentHandler{svc: svc}
}

// AlertWebhook ouvre une issue d'incident à partir d'une alerte.
//
// La route n'est pas exposée par Kong : Alertmanager appelle le service
// directement sur le réseau interne.
func (h *IncidentHandler) AlertWebhook(c *gin.Context) {
	var payload service.AlertmanagerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.svc.HandleAlert(c.Request.Context(), &payload)
	if err != nil {
		log.Printf("webhook alerte: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if url == "" {
		c.JSON(http.StatusOK, gin.H{"status": "ignoree"})
		return
	}

	log.Printf("incident consigne: %s", url)
	c.JSON(http.StatusCreated, gin.H{"issue": url})
}
