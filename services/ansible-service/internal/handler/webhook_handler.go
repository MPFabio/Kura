package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleWebhook reçoit les webhooks d'exécution Ansible (Semaphore ou autre
// orchestrateur) et en accuse réception après journalisation.
func (h *AnsibleHandler) HandleWebhook(c *gin.Context) {
	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	eventType, _ := data["event"].(string)
	if eventType == "" {
		eventType = "unknown"
	}
	jobID := data["job_id"]
	if jobID == nil {
		jobID = data["id"]
	}
	status := data["status"]

	log.Printf("Webhook reçu: event=%s, job_id=%v, status=%v", eventType, jobID, status)
	webhooksReceivedTotal.WithLabelValues(eventType).Inc()

	c.JSON(http.StatusOK, gin.H{
		"status": "received",
		"event":  eventType,
		"job_id": jobID,
	})
}
