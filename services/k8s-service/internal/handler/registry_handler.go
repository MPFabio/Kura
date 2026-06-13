package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/modulops/k8s-service/internal/service"
)

// RegistryHandler gère les requêtes HTTP liées au registre OCI interne (Zot).
type RegistryHandler struct {
	svc *service.RegistryService
}

// NewRegistryHandler crée un nouveau handler pour le registre OCI.
func NewRegistryHandler(svc *service.RegistryService) *RegistryHandler {
	return &RegistryHandler{svc: svc}
}

// ListRepositories liste les dépôts du registre.
func (h *RegistryHandler) ListRepositories(c *gin.Context) {
	items, err := h.svc.ListRepositories(c.Request.Context())
	if err != nil {
		log.Printf("Erreur ListRepositories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetRepository retourne le détail d'un dépôt (tags, statut Cosign).
func (h *RegistryHandler) GetRepository(c *gin.Context) {
	name := strings.TrimPrefix(c.Param("name"), "/")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nom de dépôt requis"})
		return
	}

	detail, err := h.svc.GetRepositoryDetail(c.Request.Context(), name)
	if err != nil {
		log.Printf("Erreur GetRepository(%s): %v", name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}
