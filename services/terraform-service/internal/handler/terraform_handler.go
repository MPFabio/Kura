package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/service"
)

// TerraformHandler gère les requêtes HTTP liées à Terraform.
type TerraformHandler struct {
	svc *service.TerraformService
	cfg *config.Config
}

// NewTerraformHandler crée un nouveau handler Terraform.
func NewTerraformHandler(svc *service.TerraformService, cfg *config.Config) *TerraformHandler {
	return &TerraformHandler{
		svc: svc,
		cfg: cfg,
	}
}

// UploadStateFile traite l'upload d'un fichier tfstate.
func (h *TerraformHandler) UploadStateFile(c *gin.Context) {
	ctx := c.Request.Context()

	// Récupérer le nom du fichier (optionnel)
	name := c.PostForm("name")
	if name == "" {
		name = "terraform.tfstate"
	}

	// Récupérer le fichier
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fichier requis"})
		return
	}

	// Ouvrir le fichier
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "impossible d'ouvrir le fichier"})
		return
	}
	defer f.Close()

	// Lire le contenu
	stateData := make([]byte, file.Size)
	if _, err := f.Read(stateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "impossible de lire le fichier"})
		return
	}

	// Parser et stocker
	stateFile, err := h.svc.ParseStateFile(ctx, name, stateData)
	if err != nil {
		log.Printf("Erreur UploadStateFile: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stateFile)
}

// UploadStateFileJSON traite l'upload d'un tfstate via JSON.
func (h *TerraformHandler) UploadStateFileJSON(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Name  string          `json:"name" binding:"required"`
		State json.RawMessage `json:"state" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parser et stocker
	stateFile, err := h.svc.ParseStateFile(ctx, req.Name, req.State)
	if err != nil {
		log.Printf("Erreur UploadStateFileJSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stateFile)
}

// ListStateFiles retourne la liste de tous les fichiers d'état.
func (h *TerraformHandler) ListStateFiles(c *gin.Context) {
	ctx := c.Request.Context()

	stateFiles, err := h.svc.ListStateFiles(ctx)
	if err != nil {
		log.Printf("Erreur ListStateFiles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": stateFiles,
	})
}

// GetStateFile retourne un fichier d'état par son ID.
func (h *TerraformHandler) GetStateFile(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	stateFile, err := h.svc.GetStateFile(ctx, id)
	if err != nil {
		log.Printf("Erreur GetStateFile pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stateFile)
}

// GetStateSummary retourne un résumé d'un état Terraform.
func (h *TerraformHandler) GetStateSummary(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	summary, err := h.svc.GetStateSummary(ctx, id)
	if err != nil {
		log.Printf("Erreur GetStateSummary pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetResources retourne toutes les ressources d'un état.
func (h *TerraformHandler) GetResources(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	resources, err := h.svc.GetResources(ctx, id)
	if err != nil {
		log.Printf("Erreur GetResources pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": resources,
	})
}

// GetOutputs retourne toutes les sorties d'un état.
func (h *TerraformHandler) GetOutputs(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	outputs, err := h.svc.GetOutputs(ctx, id)
	if err != nil {
		log.Printf("Erreur GetOutputs pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, outputs)
}

// GetResourceByAddress retourne une ressource spécifique par son adresse.
func (h *TerraformHandler) GetResourceByAddress(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	address := c.Param("address")

	if id == "" || address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id et address requis"})
		return
	}

	resource, err := h.svc.GetResourceByAddress(ctx, id, address)
	if err != nil {
		log.Printf("Erreur GetResourceByAddress pour id %s, address %s: %v", id, address, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// DetectDrift détecte les dérives pour un état.
func (h *TerraformHandler) DetectDrift(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	results, err := h.svc.DetectDrift(ctx, id)
	if err != nil {
		log.Printf("Erreur DetectDrift pour id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": results,
	})
}

// DeleteStateFile supprime un fichier d'état.
func (h *TerraformHandler) DeleteStateFile(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}

	if err := h.svc.DeleteStateFile(ctx, id); err != nil {
		log.Printf("Erreur DeleteStateFile pour id %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "fichier d'état supprimé"})
}
