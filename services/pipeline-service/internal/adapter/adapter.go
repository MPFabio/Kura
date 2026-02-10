package adapter

import (
	"io"

	"github.com/modulops/pipeline-service/internal/models"
)

// PipelineAdapter définit l'interface pour les adapters CI/CD
type PipelineAdapter interface {
	// ParseWebhook parse le payload d'un webhook et retourne un PipelineRun
	ParseWebhook(body io.Reader) (*models.PipelineRun, error)
	// Provider retourne le type de fournisseur
	Provider() models.Provider
}
