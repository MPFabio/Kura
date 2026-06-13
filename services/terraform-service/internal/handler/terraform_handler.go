package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/configstore"
	"github.com/modulops/terraform-service/internal/models"
	"github.com/modulops/terraform-service/internal/service"
)

// TerraformHandler gère les requêtes HTTP liées à Terraform.
type TerraformHandler struct {
	svc         *service.TerraformService
	syncService *service.SyncService
	cfg         *config.Config
	cfgStore    *configstore.Client
}

// NewTerraformHandler crée un nouveau handler Terraform.
func NewTerraformHandler(svc *service.TerraformService, syncService *service.SyncService, cfg *config.Config) *TerraformHandler {
	return &TerraformHandler{
		svc:         svc,
		syncService: syncService,
		cfg:         cfg,
		cfgStore:    configstore.New(cfg.AuthServiceURL, "terraform"),
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
	// Récupérer project_id (optionnel) pour associer l'état au projet
	projectID := c.PostForm("project_id")

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
	stateFile, err := h.svc.ParseStateFile(ctx, name, stateData, projectID)
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
		Name      string          `json:"name" binding:"required"`
		State     json.RawMessage `json:"state" binding:"required"`
		ProjectID string          `json:"project_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parser et stocker
	stateFile, err := h.svc.ParseStateFile(ctx, req.Name, req.State, req.ProjectID)
	if err != nil {
		log.Printf("Erreur UploadStateFileJSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stateFile)
}

// ListStateFiles retourne la liste des fichiers d'état, filtrés par project_id si fourni.
func (h *TerraformHandler) ListStateFiles(c *gin.Context) {
	ctx := c.Request.Context()

	// Récupérer project_id depuis les query params
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id requis"})
		return
	}

	stateFiles, err := h.svc.ListStateFiles(ctx, projectID)
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
// Query param "method" : "fine" (tofu plan -refresh-only via dépôt GitHub),
// "fast" (détecteurs existants par type de ressource), ou "auto" (défaut :
// fine si un dépôt GitHub est configuré sur la source, sinon fast).
func (h *TerraformHandler) DetectDrift(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	method := c.DefaultQuery("method", "auto")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id requis"})
		return
	}
	if method != "auto" && method != "fine" && method != "fast" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method invalide (auto, fine ou fast)"})
		return
	}

	// Récupérer la source associée à cet état (credentials + config GitHub)
	var credentialsJSON string
	var providerType string
	var source *models.StateSource

	if h.syncService != nil {
		sources, err := h.syncService.ListSources(ctx)
		if err == nil {
			for _, s := range sources {
				if s.StateFileID == id {
					providerType = s.Type
					sourceCopy := *s
					if err := h.syncService.DecryptCredentials(&sourceCopy); err == nil {
						source = &sourceCopy
						credentialsJSON = sourceCopy.Config.GCPCredentialsJSON
					} else {
						source = s
					}
					break
				}
			}
		}
	}

	// Resynchroniser le tfstate local depuis le backend distant avant de détecter
	// le drift : sans ça, "expected" provient de la dernière sync planifiée et peut
	// être périmé (ex: un terraform apply récent non encore reflété), produisant un
	// faux drift entre le tfstate local et l'état réel de l'infrastructure.
	if h.syncService != nil && source != nil {
		if err := h.syncService.SyncStateSync(ctx, source.ID); err != nil {
			log.Printf("Avertissement: resynchronisation du tfstate avant détection de drift échouée pour %s: %v", source.ID, err)
		}
	}

	hasGitHubRepo := source != nil && source.Config.GitHubOwner != "" && source.Config.GitHubRepo != ""

	if method == "auto" {
		if hasGitHubRepo {
			method = "fine"
		} else {
			method = "fast"
		}
	}

	if method == "fine" {
		if !hasGitHubRepo {
			c.JSON(http.StatusBadRequest, gin.H{"error": "aucun dépôt GitHub configuré pour cette source (mode fine)"})
			return
		}

		githubToken := h.cfgStore.GetOrFallback(ctx, "github_token", "")
		envCreds := map[string]string{}
		if providerType == "gcp" && credentialsJSON != "" {
			envCreds["GOOGLE_CREDENTIALS"] = credentialsJSON
		}

		results, err := h.svc.DetectDriftFine(ctx, id, source, githubToken, envCreds)
		if err != nil {
			log.Printf("Erreur DetectDriftFine pour id %s: %v", id, err)
			if c.Query("method") == "auto" || c.Query("method") == "" {
				// Fallback en mode fast si le mode fine échoue en auto
				results, err = h.svc.DetectDrift(ctx, id, credentialsJSON, providerType)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"items": results})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": results})
		return
	}

	results, err := h.svc.DetectDrift(ctx, id, credentialsJSON, providerType)
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

// Les méthodes suivantes seront ajoutées dans un nouveau fichier handler pour les sources
