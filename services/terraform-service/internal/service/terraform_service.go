package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/modulops/terraform-service/internal/config"
	gcpDrift "github.com/modulops/terraform-service/internal/drift/gcp"
	"github.com/modulops/terraform-service/internal/models"
	"github.com/modulops/terraform-service/internal/parser"
	"github.com/modulops/terraform-service/internal/storage"
)

// Cache définit l'interface minimale du cache utilisée par le service.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// TerraformService contient la logique métier autour de Terraform.
type TerraformService struct {
	parser        *parser.TFStateParser
	cache         Cache
	cfg           *config.Config
	states        map[string]*models.StateFile // Stockage en mémoire (pourrait être remplacé par PostgreSQL)
	backendWriter storage.BackendWriter        // optionnel : persistance tfstate dans un bucket S3
}

// NewTerraformService crée un nouveau service Terraform.
// Si backendWriter est non nil et cfg.StateBackend == "s3", chaque état uploadé est aussi persisté dans le bucket.
func NewTerraformService(cache Cache, cfg *config.Config, backendWriter storage.BackendWriter) *TerraformService {
	return &TerraformService{
		parser:        parser.NewTFStateParser(),
		cache:         cache,
		cfg:           cfg,
		states:        make(map[string]*models.StateFile),
		backendWriter: backendWriter,
	}
}

// ParseStateFile parse un fichier tfstate et le stocke.
// projectID optionnel : si fourni, associe l'état au projet.
func (s *TerraformService) ParseStateFile(ctx context.Context, name string, stateData []byte, projectID string) (*models.StateFile, error) {
	return s.ParseStateFileWithID(ctx, "", name, stateData, projectID)
}

// ParseStateFileWithID parse un fichier tfstate et le stocke avec un ID spécifique.
// Si l'ID existe déjà, l'état est mis à jour.
func (s *TerraformService) ParseStateFileWithID(ctx context.Context, stateFileID, name string, stateData []byte, projectID string) (*models.StateFile, error) {
	// Parser le tfstate
	state, err := s.parser.ParseStateFromBytes(stateData)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du parsing: %w", err)
	}

	// Valider l'état
	if err := s.parser.ValidateState(state); err != nil {
		return nil, fmt.Errorf("état invalide: %w", err)
	}

	var stateFile *models.StateFile
	var isUpdate bool

	// Vérifier si l'état existe déjà
	if stateFileID != "" {
		if existing, err := s.GetStateFile(ctx, stateFileID); err == nil {
			stateFile = existing
			isUpdate = true
		}
	}

	// Créer ou mettre à jour le fichier d'état
	if !isUpdate {
		// Nouvel état
		// Si l'ID commence par "temp-", générer un nouvel ID permanent
		if strings.HasPrefix(stateFileID, "temp-") {
			stateFileID = fmt.Sprintf("%s-%d", name, time.Now().Unix())
		} else if stateFileID == "" {
			stateFileID = fmt.Sprintf("%s-%d", name, time.Now().Unix())
		}
		stateFile = &models.StateFile{
			ID:         stateFileID,
			Name:       name,
			State:      state,
			ProjectID:  projectID,
			UploadedAt: time.Now(),
		}
	} else {
		// Mise à jour
		stateFile.State = state
		stateFile.UploadedAt = time.Now()
		if name != "" {
			stateFile.Name = name
		}
	}

	// Stocker en mémoire (pourrait être remplacé par PostgreSQL)
	s.states[stateFile.ID] = stateFile

	// Mettre en cache avec un TTL très long (30 jours) pour les états Terraform
	// Les états sont des données importantes qui ne doivent pas expirer rapidement
	// Utiliser project_id dans la clé si disponible
	var cacheKey string
	if stateFile.ProjectID != "" {
		cacheKey = fmt.Sprintf("terraform:state:%s:%s", stateFile.ProjectID, stateFile.ID)
	} else {
		cacheKey = fmt.Sprintf("terraform:state:%s", stateFile.ID)
	}
	stateJSON, err := json.Marshal(stateFile)
	if err == nil {
		// Utiliser un TTL de 30 jours pour les états Terraform (beaucoup plus long que le TTL par défaut)
		stateTTL := 30 * 24 * time.Hour
		_ = s.cache.Set(ctx, cacheKey, string(stateJSON), stateTTL)
	}

	// Persister dans le backend S3 si configuré (tfstate dans un bucket)
	if s.backendWriter != nil && s.cfg.StateBackend == "s3" && len(stateData) > 0 {
		key := s.cfg.S3KeyPrefix + "/" + stateFile.ID + ".tfstate"
		if stateFile.ProjectID != "" {
			key = s.cfg.S3KeyPrefix + "/" + stateFile.ProjectID + "/" + stateFile.ID + ".tfstate"
		}
		if err := s.backendWriter.PutStateFile(ctx, s.cfg.S3Bucket, key, stateData); err != nil {
			log.Printf("⚠️  Persistance tfstate vers S3 ignorée: %v", err)
		}
	}

	return stateFile, nil
}

// GetStateFile récupère un fichier d'état par son ID.
func (s *TerraformService) GetStateFile(ctx context.Context, id string) (*models.StateFile, error) {
	// Vérifier le cache avec la clé simple (terraform:state:<id>)
	cacheKey := fmt.Sprintf("terraform:state:%s", id)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var stateFile models.StateFile
		if err := json.Unmarshal([]byte(cached), &stateFile); err == nil {
			return &stateFile, nil
		}
	}

	// Fallback sur le stockage en mémoire
	if stateFile, exists := s.states[id]; exists {
		return stateFile, nil
	}

	// Recherche dans le cache avec le préfixe project_id (terraform:state:<projectID>:<id>)
	type KeysCache interface {
		Keys(ctx context.Context, pattern string) ([]string, error)
	}
	if cacheWithKeys, ok := s.cache.(KeysCache); ok {
		keys, err := cacheWithKeys.Keys(ctx, "terraform:state:*:"+id)
		if err == nil {
			for _, key := range keys {
				if cached, err := s.cache.Get(ctx, key); err == nil && cached != "" {
					var stateFile models.StateFile
					if err := json.Unmarshal([]byte(cached), &stateFile); err == nil {
						// Mettre en cache sous la clé simple pour les prochaines recherches
						_ = s.cache.Set(ctx, cacheKey, cached, 30*24*time.Hour)
						s.states[stateFile.ID] = &stateFile
						return &stateFile, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("fichier d'état non trouvé: %s", id)
}

// ListStateFiles retourne la liste des fichiers d'état, filtrés par project_id si fourni.
func (s *TerraformService) ListStateFiles(ctx context.Context, projectID string) ([]*models.StateFile, error) {
	resultMap := make(map[string]*models.StateFile)

	// Charger depuis la mémoire
	for _, stateFile := range s.states {
		// Filtrer par project_id si fourni
		if projectID != "" && stateFile.ProjectID != projectID {
			continue
		}
		resultMap[stateFile.ID] = stateFile
	}

	// Charger depuis Redis si l'interface Keys est disponible
	// Toujours utiliser le pattern "terraform:state:*" pour trouver les états quel que soit le format de clé
	// (terraform:state:id ou terraform:state:projectID:id), puis filtrer par project_id.
	type KeysCache interface {
		Keys(ctx context.Context, pattern string) ([]string, error)
	}
	if cacheWithKeys, ok := s.cache.(KeysCache); ok {
		keys, err := cacheWithKeys.Keys(ctx, "terraform:state:*")
		if err == nil {
			for _, key := range keys {
				cached, err := s.cache.Get(ctx, key)
				if err == nil && cached != "" {
					var stateFile models.StateFile
					if err := json.Unmarshal([]byte(cached), &stateFile); err == nil {
						// Filtrer par project_id si fourni : garder les états du projet ou sans project (ancien format)
						if projectID != "" && stateFile.ProjectID != projectID && stateFile.ProjectID != "" {
							continue
						}
						resultMap[stateFile.ID] = &stateFile
						s.states[stateFile.ID] = &stateFile
					}
				}
			}
		}
	}

	// Convertir en slice
	result := make([]*models.StateFile, 0, len(resultMap))
	for _, stateFile := range resultMap {
		result = append(result, stateFile)
	}

	return result, nil
}

// GetStateSummary retourne un résumé d'un état Terraform.
func (s *TerraformService) GetStateSummary(ctx context.Context, id string) (*models.StateSummary, error) {
	stateFile, err := s.GetStateFile(ctx, id)
	if err != nil {
		return nil, err
	}

	if stateFile.State == nil {
		return nil, fmt.Errorf("état vide")
	}

	summary := &models.StateSummary{
		ResourceCount: len(stateFile.State.Resources),
		OutputCount:   len(stateFile.State.Outputs),
		LastModified:  stateFile.UploadedAt,
		DriftCount:    0, // Sera rempli après détection de drift
	}

	return summary, nil
}

// DetectDrift détecte les dérives entre l'état Terraform et l'état réel.
// Cette méthode délègue à un détecteur spécifique selon le provider.
func (s *TerraformService) DetectDrift(ctx context.Context, stateFileID string, credentialsJSON string, providerType string) ([]*models.DriftResult, error) {
	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return nil, err
	}

	if stateFile.State == nil {
		return nil, fmt.Errorf("état vide")
	}

	var results []*models.DriftResult

	// Utiliser le détecteur approprié selon le provider
	switch providerType {
	case "gcp":
		detector, err := gcpDrift.NewDetector(credentialsJSON)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la création du détecteur GCP: %w", err)
		}
		results, err = detector.DetectDrift(ctx, stateFile)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la détection de drift GCP: %w", err)
		}
	case "aws", "s3":
		// TODO: Implémenter le détecteur AWS
		results = make([]*models.DriftResult, 0)
		for _, resource := range stateFile.State.Resources {
			resourceAddr := s.parser.BuildResourceAddress(&resource)
			results = append(results, &models.DriftResult{
				ResourceAddress: resourceAddr,
				ResourceType:    resource.Type,
				Status:          "unknown",
				DetectedAt:      time.Now(),
				Message:         "Détection de drift AWS non implémentée",
			})
		}
	case "azure":
		// TODO: Implémenter le détecteur Azure
		results = make([]*models.DriftResult, 0)
		for _, resource := range stateFile.State.Resources {
			resourceAddr := s.parser.BuildResourceAddress(&resource)
			results = append(results, &models.DriftResult{
				ResourceAddress: resourceAddr,
				ResourceType:    resource.Type,
				Status:          "unknown",
				DetectedAt:      time.Now(),
				Message:         "Détection de drift Azure non implémentée",
			})
		}
	default:
		// Fallback: détection basique
		results = make([]*models.DriftResult, 0)
		for _, resource := range stateFile.State.Resources {
			resourceAddr := s.parser.BuildResourceAddress(&resource)
			driftResult := &models.DriftResult{
				ResourceAddress: resourceAddr,
				ResourceType:    resource.Type,
				Status:          "unknown",
				DetectedAt:      time.Now(),
				Message:         fmt.Sprintf("Provider %s non supporté pour la détection de drift", providerType),
			}
			if len(resource.Instances) == 0 {
				driftResult.Status = "missing"
				driftResult.Message = "Aucune instance trouvée pour cette ressource"
			}
			results = append(results, driftResult)
		}
	}

	// Mettre à jour le stateFile avec les résultats
	stateFile.DriftResults = results
	stateFile.LastChecked = time.Now()

	stateTTL := 30 * 24 * time.Hour
	stateJSON, err := json.Marshal(stateFile)
	if err == nil {
		// Sauvegarder sous la clé simple
		_ = s.cache.Set(ctx, fmt.Sprintf("terraform:state:%s", stateFile.ID), string(stateJSON), stateTTL)
		// Sauvegarder aussi sous la clé avec project_id si disponible
		if stateFile.ProjectID != "" {
			_ = s.cache.Set(ctx, fmt.Sprintf("terraform:state:%s:%s", stateFile.ProjectID, stateFile.ID), string(stateJSON), stateTTL)
		}
	}

	// Mettre à jour en mémoire
	s.states[stateFile.ID] = stateFile

	return results, nil
}

// DeleteStateFile supprime un fichier d'état.
func (s *TerraformService) DeleteStateFile(ctx context.Context, id string) error {
	// Récupérer l'état pour connaître project_id (clé Redis peut être terraform:state:projectID:id)
	stateFile, exists := s.states[id]
	if exists {
		// Supprimer les deux formes de clé cache (avec et sans project_id)
		cacheKey := fmt.Sprintf("terraform:state:%s", id)
		_ = s.cache.Delete(ctx, cacheKey)
		if stateFile.ProjectID != "" {
			projectCacheKey := fmt.Sprintf("terraform:state:%s:%s", stateFile.ProjectID, id)
			_ = s.cache.Delete(ctx, projectCacheKey)
		}
		delete(s.states, id)
		return nil
	}

	// État pas en mémoire : supprimer du cache Redis (clé avec project_id) pour qu’il disparaisse à la prochaine liste
	if deleted := s.deleteStateFileFromCacheByID(ctx, id); deleted {
		return nil
	}
	return fmt.Errorf("fichier d'état non trouvé: %s", id)
}

// deleteStateFileFromCacheByID supprime du cache toutes les clés correspondant à cet id.
// Retourne true si au moins une clé a été supprimée.
func (s *TerraformService) deleteStateFileFromCacheByID(ctx context.Context, id string) bool {
	type KeysCache interface {
		Keys(ctx context.Context, pattern string) ([]string, error)
	}
	cacheWithKeys, ok := s.cache.(KeysCache)
	if !ok {
		return false
	}
	keys, err := cacheWithKeys.Keys(ctx, "terraform:state:*")
	if err != nil {
		return false
	}
	// Ne supprimer que les clés qui correspondent exactement à cet id (éviter id sous-chaîne)
	// Format: "terraform:state:id" ou "terraform:state:projectID:id" (exactement 3 ':' dans ce cas)
	var deleted bool
	for _, key := range keys {
		if key == "terraform:state:"+id || (strings.HasSuffix(key, ":"+id) && strings.Count(key, ":") == 3) {
			_ = s.cache.Delete(ctx, key)
			deleted = true
		}
	}
	return deleted
}

// GetResources retourne toutes les ressources d'un état.
func (s *TerraformService) GetResources(ctx context.Context, stateFileID string) ([]models.Resource, error) {
	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return nil, err
	}

	if stateFile.State == nil {
		return nil, fmt.Errorf("état vide")
	}

	return s.parser.ExtractResources(stateFile.State), nil
}

// GetOutputs retourne toutes les sorties d'un état.
func (s *TerraformService) GetOutputs(ctx context.Context, stateFileID string) (map[string]models.Output, error) {
	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return nil, err
	}

	if stateFile.State == nil {
		return nil, fmt.Errorf("état vide")
	}

	return s.parser.ExtractOutputs(stateFile.State), nil
}

// GetResourceByAddress retourne une ressource spécifique par son adresse.
func (s *TerraformService) GetResourceByAddress(ctx context.Context, stateFileID, address string) (*models.Resource, error) {
	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return nil, err
	}

	return s.parser.GetResourceByAddress(stateFile.State, address)
}
