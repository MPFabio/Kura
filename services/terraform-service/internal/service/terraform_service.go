package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/models"
	"github.com/modulops/terraform-service/internal/parser"
)

// Cache définit l'interface minimale du cache utilisée par le service.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// TerraformService contient la logique métier autour de Terraform.
type TerraformService struct {
	parser   *parser.TFStateParser
	cache    Cache
	cfg      *config.Config
	states   map[string]*models.StateFile // Stockage en mémoire (pourrait être remplacé par PostgreSQL)
}

// NewTerraformService crée un nouveau service Terraform.
func NewTerraformService(cache Cache, cfg *config.Config) *TerraformService {
	return &TerraformService{
		parser: parser.NewTFStateParser(),
		cache:  cache,
		cfg:    cfg,
		states: make(map[string]*models.StateFile),
	}
}

// ParseStateFile parse un fichier tfstate et le stocke.
func (s *TerraformService) ParseStateFile(ctx context.Context, name string, stateData []byte) (*models.StateFile, error) {
	// Parser le tfstate
	state, err := s.parser.ParseStateFromBytes(stateData)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du parsing: %w", err)
	}

	// Valider l'état
	if err := s.parser.ValidateState(state); err != nil {
		return nil, fmt.Errorf("état invalide: %w", err)
	}

	// Créer ou mettre à jour le fichier d'état
	stateFile := &models.StateFile{
		ID:         fmt.Sprintf("%s-%d", name, time.Now().Unix()),
		Name:       name,
		State:      state,
		UploadedAt: time.Now(),
	}

	// Stocker en mémoire (pourrait être remplacé par PostgreSQL)
	s.states[stateFile.ID] = stateFile

	// Mettre en cache
	cacheKey := fmt.Sprintf("terraform:state:%s", stateFile.ID)
	stateJSON, err := json.Marshal(stateFile)
	if err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(stateJSON), s.cfg.CacheTTL)
	}

	return stateFile, nil
}

// GetStateFile récupère un fichier d'état par son ID.
func (s *TerraformService) GetStateFile(ctx context.Context, id string) (*models.StateFile, error) {
	// Vérifier le cache
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

	return nil, fmt.Errorf("fichier d'état non trouvé: %s", id)
}

// ListStateFiles retourne la liste de tous les fichiers d'état.
func (s *TerraformService) ListStateFiles(ctx context.Context) ([]*models.StateFile, error) {
	result := make([]*models.StateFile, 0, len(s.states))
	for _, stateFile := range s.states {
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
// Cette implémentation est basique et peut être étendue avec des providers réels.
func (s *TerraformService) DetectDrift(ctx context.Context, stateFileID string) ([]*models.DriftResult, error) {
	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return nil, err
	}

	if stateFile.State == nil {
		return nil, fmt.Errorf("état vide")
	}

	results := make([]*models.DriftResult, 0)

	// Parcourir toutes les ressources
	for _, resource := range stateFile.State.Resources {
		resourceAddr := s.parser.BuildResourceAddress(&resource)

		// Pour l'instant, on simule une détection de drift basique
		// Dans une implémentation complète, on interrogerait les providers réels
		// (AWS, GCP, Azure, etc.) pour comparer avec l'état réel

		driftResult := &models.DriftResult{
			ResourceAddress: resourceAddr,
			ResourceType:    resource.Type,
			Status:          "unknown", // "in_sync", "drifted", "missing"
			DetectedAt:      time.Now(),
			Message:         "Détection de drift non implémentée pour ce type de ressource",
		}

		// Exemple : si la ressource n'a pas d'instances, elle est considérée comme manquante
		if len(resource.Instances) == 0 {
			driftResult.Status = "missing"
			driftResult.Message = "Aucune instance trouvée pour cette ressource"
		} else {
			// Par défaut, on considère qu'il n'y a pas de drift
			// (ceci nécessiterait une vérification réelle avec les providers)
			driftResult.Status = "unknown"
		}

		results = append(results, driftResult)
	}

	// Mettre à jour le timestamp de dernière vérification
	stateFile.LastChecked = time.Now()

	return results, nil
}

// DeleteStateFile supprime un fichier d'état.
func (s *TerraformService) DeleteStateFile(ctx context.Context, id string) error {
	// Supprimer du cache
	cacheKey := fmt.Sprintf("terraform:state:%s", id)
	_ = s.cache.Delete(ctx, cacheKey)

	// Supprimer du stockage
	if _, exists := s.states[id]; !exists {
		return fmt.Errorf("fichier d'état non trouvé: %s", id)
	}

	delete(s.states, id)
	return nil
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
