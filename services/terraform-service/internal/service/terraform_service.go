package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/modulops/terraform-service/internal/client"
	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/drift/fine"
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

	// Charger depuis Valkey si l'interface Keys est disponible
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

// États possibles d'une détection de drift lancée en tâche de fond.
const (
	driftStatusRunning = "running"
	driftStatusDone    = "done"
	driftStatusError   = "error"
)

// driftMaxDuration borne la durée d'une détection. Au-delà, une détection
// encore marquée "running" est tenue pour perdue (redémarrage du service en
// plein plan) et une nouvelle peut être lancée.
const driftMaxDuration = 15 * time.Minute

// DetectDriftFine détecte les dérives en exécutant `tofu plan -refresh-only`
// contre les fichiers .tf source récupérés depuis Forgejo/Codeberg et le tfstate stocké.
// Retourne une erreur si la source n'a pas de dépôt Forgejo/Codeberg configuré.
func (s *TerraformService) DetectDriftFine(ctx context.Context, stateFileID string, source *models.StateSource, forgejoToken string, envCreds map[string]string) ([]*models.DriftResult, error) {
	if source == nil || source.Config.ForgejoOwner == "" || source.Config.ForgejoRepo == "" {
		return nil, fmt.Errorf("aucun dépôt Forgejo/Codeberg configuré pour cette source")
	}

	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return nil, err
	}
	if stateFile.State == nil {
		return nil, fmt.Errorf("état vide")
	}

	stateJSON, err := json.Marshal(stateFile.State)
	if err != nil {
		return nil, fmt.Errorf("sérialisation du tfstate: %w", err)
	}

	fj := client.NewForgejoClient(source.Config.ForgejoURL, forgejoToken)
	ref := source.Config.ForgejoRef
	if ref == "" {
		ref = "main"
	}
	moduleDir := strings.Trim(source.Config.ForgejoPath, "/")
	tfFiles, err := fj.FetchTFFiles(source.Config.ForgejoOwner, source.Config.ForgejoRepo, source.Config.ForgejoPath, ref)
	if err != nil {
		return nil, fmt.Errorf("récupération des fichiers .tf: %w", err)
	}

	// Récupère les fichiers annexes référencés via file("${path.module}/...")
	// (ex: scripts cloud-init) qui se trouvent hors du répertoire du module.
	for _, refPath := range fine.ExtractModuleFileRefs(tfFiles, moduleDir) {
		extra, err := fj.FetchFile(source.Config.ForgejoOwner, source.Config.ForgejoRepo, refPath, ref)
		if err != nil {
			return nil, fmt.Errorf("récupération du fichier référencé %s: %w", refPath, err)
		}
		tfFiles = append(tfFiles, extra)
	}

	outputValues := make(map[string]interface{}, len(stateFile.State.Outputs))
	for name, out := range stateFile.State.Outputs {
		outputValues[name] = out.Value
	}

	runner := fine.NewRunner(s.cfg.TofuPath)
	plan, err := runner.Run(ctx, fine.RunInput{
		TFFiles:   tfFiles,
		ModuleDir: moduleDir,
		StateJSON: stateJSON,
		EnvCreds:  envCreds,
		Outputs:   outputValues,
	})
	if err != nil {
		return nil, fmt.Errorf("exécution tofu plan: %w", err)
	}

	results := fine.ParsePlan(plan, stateFile.State.Resources, time.Now())

	s.persistDriftResults(ctx, stateFile, results)

	return results, nil
}

// persistDriftResults met à jour le stateFile avec les résultats de drift et le persiste en cache/mémoire.
func (s *TerraformService) persistDriftResults(ctx context.Context, stateFile *models.StateFile, results []*models.DriftResult) {
	stateFile.DriftResults = results
	stateFile.LastChecked = time.Now()
	s.persistStateFile(ctx, stateFile)
}

// persistStateFile écrit l'état en cache et en mémoire.
func (s *TerraformService) persistStateFile(ctx context.Context, stateFile *models.StateFile) {
	stateTTL := 30 * 24 * time.Hour
	stateJSON, err := json.Marshal(stateFile)
	if err == nil {
		_ = s.cache.Set(ctx, fmt.Sprintf("terraform:state:%s", stateFile.ID), string(stateJSON), stateTTL)
		if stateFile.ProjectID != "" {
			_ = s.cache.Set(ctx, fmt.Sprintf("terraform:state:%s:%s", stateFile.ProjectID, stateFile.ID), string(stateJSON), stateTTL)
		}
	}

	s.states[stateFile.ID] = stateFile
}

// DeleteStateFile supprime un fichier d'état.
func (s *TerraformService) DeleteStateFile(ctx context.Context, id string) error {
	// Récupérer l'état pour connaître project_id (clé Valkey peut être terraform:state:projectID:id)
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

	// État pas en mémoire : supprimer du cache Valkey (clé avec project_id) pour qu’il disparaisse à la prochaine liste
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

// MarkDriftRunning signale qu'une détection démarre, afin que l'interface
// puisse afficher l'avancement sans maintenir de requête ouverte.
//
// Retourne false si une détection est déjà en cours pour cet état : les plans
// étant sérialisés, en empiler plusieurs ne ferait qu'allonger l'attente.
func (s *TerraformService) MarkDriftRunning(ctx context.Context, stateFileID string) (bool, error) {
	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return false, err
	}
	// Une détection démarrée il y a plus longtemps que le plafond d'exécution
	// est considérée comme perdue (redémarrage du service en cours de plan),
	// sans quoi l'état resterait bloqué sur "running" indéfiniment.
	if stateFile.DriftStatus == driftStatusRunning && time.Since(stateFile.DriftStartedAt) < driftMaxDuration {
		return false, nil
	}
	stateFile.DriftStatus = driftStatusRunning
	stateFile.DriftStartedAt = time.Now()
	stateFile.DriftError = ""
	s.persistStateFile(ctx, stateFile)
	return true, nil
}

// MarkDriftFinished enregistre l'issue d'une détection lancée en tâche de fond.
func (s *TerraformService) MarkDriftFinished(ctx context.Context, stateFileID string, runErr error) {
	stateFile, err := s.GetStateFile(ctx, stateFileID)
	if err != nil {
		return
	}
	if runErr != nil {
		stateFile.DriftStatus = driftStatusError
		stateFile.DriftError = runErr.Error()
	} else {
		stateFile.DriftStatus = driftStatusDone
		stateFile.DriftError = ""
	}
	s.persistStateFile(ctx, stateFile)
}
