package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/models"
	"github.com/modulops/terraform-service/internal/s3"
	azureStorage "github.com/modulops/terraform-service/internal/storage/azure"
	gcpStorage "github.com/modulops/terraform-service/internal/storage/gcp"
)

// #region agent log
func debugLog(location, message string, data map[string]interface{}) {
	logEntry := map[string]interface{}{
		"sessionId": "debug-session",
		"runId":     "run1",
		"location":  location,
		"message":   message,
		"data":      data,
		"timestamp": time.Now().UnixMilli(),
	}
	if jsonData, err := json.Marshal(logEntry); err == nil {
		// Écrire dans /tmp qui est accessible même dans distroless
		if f, err := os.OpenFile("/tmp/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.WriteString(string(jsonData) + "\n")
			f.Close()
		}
		// Aussi logger via le logger standard pour voir dans docker logs
		log.Printf("[DEBUG] %s: %s - %+v", location, message, data)
	}
}

// #endregion

// SyncService gère la synchronisation des états Terraform depuis différentes sources.
type SyncService struct {
	terraformService *TerraformService
	cache            Cache
	cfg              *config.Config
	sources          map[string]*models.StateSource // ID source -> source
	jobs             map[string]*models.SyncJob     // ID job -> job
	mu               sync.RWMutex
	syncTicker       *time.Ticker
	stopChan         chan struct{}
	encryptionKey    []byte // Pour chiffrer les credentials sensibles
}

// NewSyncService crée un nouveau service de synchronisation.
func NewSyncService(terraformService *TerraformService, cache Cache, cfg *config.Config) *SyncService {
	// Utiliser une clé de chiffrement fixe depuis la config ou une valeur par défaut
	// IMPORTANT: Cette clé doit être la même à chaque démarrage pour pouvoir déchiffrer les credentials existants
	encryptionKeyStr := cfg.EncryptionKey
	if encryptionKeyStr == "" {
		encryptionKeyStr = "default-encryption-key-32-bytes!!" // ⚠️ Clé par défaut pour le développement
		log.Printf("⚠️  Utilisation d'une clé de chiffrement par défaut. En production, définissez TERRAFORM_ENCRYPTION_KEY")
	}

	// Convertir la clé en bytes (doit faire exactement 32 bytes pour AES-256)
	key := []byte(encryptionKeyStr)
	if len(key) < 32 {
		// Padding avec des zéros si trop court
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		// Tronquer si trop long
		key = key[:32]
	}

	ss := &SyncService{
		terraformService: terraformService,
		cache:            cache,
		cfg:              cfg,
		sources:          make(map[string]*models.StateSource),
		jobs:             make(map[string]*models.SyncJob),
		encryptionKey:    key,
		stopChan:         make(chan struct{}),
	}

	// Charger les sources depuis le cache au démarrage
	ss.loadSourcesFromCache(context.Background())

	// Démarrer le scheduler de synchronisation
	ss.startScheduler()

	return ss
}

// AddSource ajoute une nouvelle source de synchronisation.
// Si l'état n'existe pas encore et que la source est de type S3, l'état est créé depuis S3.
func (s *SyncService) AddSource(ctx context.Context, source *models.StateSource) (*models.StateSource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Générer un ID si nécessaire
	if source.ID == "" {
		source.ID = fmt.Sprintf("source-%d", time.Now().UnixNano())
	}

	// Si l'état n'existe pas encore (ID commence par "temp-"), créer l'état depuis la source cloud
	if strings.HasPrefix(source.StateFileID, "temp-") {
		// Décrypter temporairement pour la synchronisation initiale
		sourceCopy := *source
		if err := s.decryptCredentials(&sourceCopy); err == nil {
			var stateData []byte
			var fileName string
			var err error

			switch source.Type {
			case "s3":
				s3Client, err := s3.NewClient(
					sourceCopy.Config.S3Region,
					sourceCopy.Config.S3Endpoint,
					sourceCopy.Config.AWSAccessKeyID,
					sourceCopy.Config.AWSSecretAccessKey,
				)
				if err == nil {
					stateData, err = s3Client.GetStateFile(ctx, sourceCopy.Config.S3Bucket, sourceCopy.Config.S3Key)
					if err == nil {
						keyParts := strings.Split(sourceCopy.Config.S3Key, "/")
						fileName = keyParts[len(keyParts)-1]
						if fileName == "" {
							fileName = "terraform.tfstate"
						}
					}
				}
			case "gcp":
				gcpClient, err := gcpStorage.NewClient(sourceCopy.Config.GCPCredentialsJSON)
				if err == nil {
					stateData, err = gcpClient.GetStateFile(ctx, sourceCopy.Config.GCPBucket, sourceCopy.Config.GCPObjectName)
					if err == nil {
						keyParts := strings.Split(sourceCopy.Config.GCPObjectName, "/")
						fileName = keyParts[len(keyParts)-1]
						if fileName == "" {
							fileName = "terraform.tfstate"
						}
					}
				}
			case "azure":
				azureClient, err := azureStorage.NewClient(
					sourceCopy.Config.AzureAccountName,
					sourceCopy.Config.AzureAccountKey,
					sourceCopy.Config.AzureConnectionString,
				)
				if err == nil {
					stateData, err = azureClient.GetStateFile(ctx, sourceCopy.Config.AzureContainer, sourceCopy.Config.AzureBlobName)
					if err == nil {
						keyParts := strings.Split(sourceCopy.Config.AzureBlobName, "/")
						fileName = keyParts[len(keyParts)-1]
						if fileName == "" {
							fileName = "terraform.tfstate"
						}
					}
				}
			}

			if err == nil && len(stateData) > 0 {
				// Extraire le nom depuis l'ID temporaire (format: temp-nom-1234567890)
				parts := strings.Split(source.StateFileID, "-")
				var stateName string
				if len(parts) >= 3 {
					// Reconstruire le nom (tout sauf "temp" et le timestamp)
					stateName = strings.Join(parts[1:len(parts)-1], "-")
				} else {
					// Fallback: utiliser le nom du fichier
					stateName = strings.TrimSuffix(fileName, ".tfstate")
				}

				// Générer un ID permanent basé sur le nom
				stateFileID := fmt.Sprintf("%s-%d", stateName, time.Now().Unix())
				// Créer l'état
				stateFile, err := s.terraformService.ParseStateFileWithID(ctx, stateFileID, stateName, stateData, "")
				if err != nil {
					log.Printf("⚠️  Erreur lors de la création initiale de l'état depuis %s: %v", source.Type, err)
				} else {
					// Mettre à jour l'ID de la source avec l'ID réel de l'état
					source.StateFileID = stateFile.ID
					log.Printf("✅ État créé depuis %s: %s (ID: %s)", source.Type, stateFile.Name, stateFile.ID)
				}
			} else if err != nil {
				log.Printf("⚠️  Erreur lors de la récupération du tfstate depuis %s: %v", source.Type, err)
			}
		}
	} else {
		// Vérifier si l'état existe, sinon essayer de le créer depuis la source
		_, err := s.terraformService.GetStateFile(ctx, source.StateFileID)
		if err != nil {
			// L'état n'existe pas, essayer de le créer depuis la source
			log.Printf("🔄 État %s non trouvé, tentative de création depuis la source %s...", source.StateFileID, source.Type)
			// Utiliser la même logique que ci-dessus mais avec l'ID existant
			sourceCopy := *source
			if err := s.decryptCredentials(&sourceCopy); err == nil {
				var stateData []byte
				var err error

				switch source.Type {
				case "s3":
					s3Client, err := s3.NewClient(
						sourceCopy.Config.S3Region,
						sourceCopy.Config.S3Endpoint,
						sourceCopy.Config.AWSAccessKeyID,
						sourceCopy.Config.AWSSecretAccessKey,
					)
					if err == nil {
						stateData, err = s3Client.GetStateFile(ctx, sourceCopy.Config.S3Bucket, sourceCopy.Config.S3Key)
					}
				case "gcp":
					gcpClient, err := gcpStorage.NewClient(sourceCopy.Config.GCPCredentialsJSON)
					if err == nil {
						stateData, err = gcpClient.GetStateFile(ctx, sourceCopy.Config.GCPBucket, sourceCopy.Config.GCPObjectName)
					}
				case "azure":
					azureClient, err := azureStorage.NewClient(
						sourceCopy.Config.AzureAccountName,
						sourceCopy.Config.AzureAccountKey,
						sourceCopy.Config.AzureConnectionString,
					)
					if err == nil {
						stateData, err = azureClient.GetStateFile(ctx, sourceCopy.Config.AzureContainer, sourceCopy.Config.AzureBlobName)
					}
				}

				if err == nil && len(stateData) > 0 {
					var fileName string
					switch source.Type {
					case "s3":
						keyParts := strings.Split(sourceCopy.Config.S3Key, "/")
						fileName = keyParts[len(keyParts)-1]
					case "gcp":
						keyParts := strings.Split(sourceCopy.Config.GCPObjectName, "/")
						fileName = keyParts[len(keyParts)-1]
					case "azure":
						keyParts := strings.Split(sourceCopy.Config.AzureBlobName, "/")
						fileName = keyParts[len(keyParts)-1]
					}
					if fileName == "" {
						fileName = "terraform.tfstate"
					}
					stateName := strings.TrimSuffix(fileName, ".tfstate")
					_, err = s.terraformService.ParseStateFileWithID(ctx, source.StateFileID, stateName, stateData, "")
					if err != nil {
						log.Printf("⚠️  Erreur lors de la création de l'état depuis %s: %v", source.Type, err)
					} else {
						log.Printf("✅ État créé depuis %s: %s", source.Type, source.StateFileID)
					}
				}
			}
		}
	}

	// Chiffrer les credentials sensibles
	if err := s.encryptCredentials(source); err != nil {
		return nil, fmt.Errorf("erreur lors du chiffrement des credentials: %w", err)
	}

	source.CreatedAt = time.Now()
	source.UpdatedAt = time.Now()

	s.sources[source.ID] = source

	// Sauvegarder dans le cache
	s.saveSourceToCache(ctx, source)

	return source, nil
}

// GetSource récupère une source par son ID.
func (s *SyncService) GetSource(ctx context.Context, id string) (*models.StateSource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	source, exists := s.sources[id]
	if !exists {
		return nil, fmt.Errorf("source non trouvée: %s", id)
	}

	// Décrypter les credentials pour l'affichage (optionnel, ou ne pas les exposer)
	sourceCopy := *source
	if err := s.decryptCredentials(&sourceCopy); err != nil {
		log.Printf("⚠️  Erreur lors du déchiffrement: %v", err)
	}

	return &sourceCopy, nil
}

// ListSources retourne toutes les sources.
func (s *SyncService) ListSources(ctx context.Context) ([]*models.StateSource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.StateSource, 0, len(s.sources))
	for _, source := range s.sources {
		sourceCopy := *source
		// Ne pas exposer les credentials dans la liste
		sourceCopy.Config.AWSAccessKeyID = ""
		sourceCopy.Config.AWSSecretAccessKey = ""
		sourceCopy.Config.TerraformCloudToken = ""
		result = append(result, &sourceCopy)
	}

	return result, nil
}

// UpdateSource met à jour une source de synchronisation existante.
func (s *SyncService) UpdateSource(ctx context.Context, id string, updatedSource *models.StateSource) (*models.StateSource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Vérifier que la source existe
	existingSource, exists := s.sources[id]
	if !exists {
		return nil, fmt.Errorf("source non trouvée: %s", id)
	}

	// Mettre à jour les champs (conserver l'ID et les timestamps)
	updatedSource.ID = id
	updatedSource.CreatedAt = existingSource.CreatedAt
	updatedSource.UpdatedAt = time.Now()

	// Si de nouveaux credentials sont fournis, les chiffrer
	// Sinon, conserver les credentials existants (déjà chiffrés)
	if updatedSource.Config.AWSSecretAccessKey != "" {
		// Nouveau credential fourni, le chiffrer
		if err := s.encryptCredentials(updatedSource); err != nil {
			return nil, fmt.Errorf("erreur lors du chiffrement des credentials: %w", err)
		}
	} else {
		// Conserver le credential existant
		updatedSource.Config.AWSSecretAccessKey = existingSource.Config.AWSSecretAccessKey
	}

	if updatedSource.Config.AzureAccountKey != "" {
		// Nouveau credential fourni, le chiffrer
		if err := s.encryptCredentials(updatedSource); err != nil {
			return nil, fmt.Errorf("erreur lors du chiffrement des credentials: %w", err)
		}
	} else {
		// Conserver les credentials existants
		updatedSource.Config.AzureAccountKey = existingSource.Config.AzureAccountKey
		updatedSource.Config.AzureConnectionString = existingSource.Config.AzureConnectionString
	}

	if updatedSource.Config.GCPCredentialsJSON != "" {
		// Nouveau credential fourni, le chiffrer
		if err := s.encryptCredentials(updatedSource); err != nil {
			return nil, fmt.Errorf("erreur lors du chiffrement des credentials: %w", err)
		}
	} else {
		// Conserver le credential existant
		updatedSource.Config.GCPCredentialsJSON = existingSource.Config.GCPCredentialsJSON
	}

	if updatedSource.Config.TerraformCloudToken != "" {
		// Nouveau token fourni, le chiffrer
		if err := s.encryptCredentials(updatedSource); err != nil {
			return nil, fmt.Errorf("erreur lors du chiffrement des credentials: %w", err)
		}
	} else {
		// Conserver le token existant
		updatedSource.Config.TerraformCloudToken = existingSource.Config.TerraformCloudToken
	}

	// Mettre à jour dans la map
	s.sources[id] = updatedSource

	// Sauvegarder dans le cache
	s.saveSourceToCache(ctx, updatedSource)

	return updatedSource, nil
}

// DeleteSource supprime une source de synchronisation.
func (s *SyncService) DeleteSource(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Vérifier que la source existe
	if _, exists := s.sources[id]; !exists {
		return fmt.Errorf("source non trouvée: %s", id)
	}

	// Supprimer de la map
	delete(s.sources, id)

	// Supprimer du cache
	s.deleteSourceFromCache(ctx, id)

	// Supprimer les jobs associés
	for jobID, job := range s.jobs {
		if job.SourceID == id {
			delete(s.jobs, jobID)
		}
	}

	return nil
}

// SyncState synchronise un état depuis sa source.
func (s *SyncService) SyncState(ctx context.Context, sourceID string) (*models.SyncJob, error) {
	source, err := s.GetSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	// Créer un job de synchronisation
	job := &models.SyncJob{
		ID:          fmt.Sprintf("job-%d", time.Now().UnixNano()),
		StateFileID: source.StateFileID,
		SourceID:    sourceID,
		Status:      "running",
		StartedAt:   time.Now(),
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	// Exécuter la synchronisation en arrière-plan
	go s.executeSync(ctx, source, job)

	return job, nil
}

// executeSync exécute la synchronisation.
func (s *SyncService) executeSync(ctx context.Context, source *models.StateSource, job *models.SyncJob) {
	// #region agent log
	debugLog("sync_service.go:414", "executeSync entry", map[string]interface{}{
		"sourceId": source.ID, "sourceType": source.Type, "stateFileID": source.StateFileID, "jobId": job.ID,
	})
	// #endregion
	defer func() {
		job.CompletedAt = time.Now()
		if job.Status == "running" {
			job.Status = "success"
		}
		// #region agent log
		debugLog("sync_service.go:423", "executeSync exit", map[string]interface{}{
			"jobId": job.ID, "status": job.Status, "error": job.Error, "message": job.Message,
		})
		// #endregion
		s.mu.Lock()
		s.jobs[job.ID] = job
		s.mu.Unlock()
	}()

	// Décrypter les credentials
	// #region agent log
	debugLog("sync_service.go:426", "before decryptCredentials", map[string]interface{}{
		"sourceId": source.ID, "hasGCPCreds": source.Config.GCPCredentialsJSON != "",
	})
	// #endregion
	if err := s.decryptCredentials(source); err != nil {
		// #region agent log
		debugLog("sync_service.go:430", "decryptCredentials failed", map[string]interface{}{
			"sourceId": source.ID, "error": err.Error(),
		})
		// #endregion
		job.Status = "failed"
		job.Error = fmt.Sprintf("erreur lors du déchiffrement: %v", err)
		return
	}
	// #region agent log
	debugLog("sync_service.go:432", "after decryptCredentials success", map[string]interface{}{
		"sourceId": source.ID, "hasGCPCreds": source.Config.GCPCredentialsJSON != "",
	})
	// #endregion

	switch source.Type {
	case "s3":
		if err := s.syncFromS3(ctx, source, job); err != nil {
			job.Status = "failed"
			job.Error = err.Error()
			return
		}
	case "azure":
		if err := s.syncFromAzure(ctx, source, job); err != nil {
			job.Status = "failed"
			job.Error = err.Error()
			return
		}
	case "gcp":
		// #region agent log
		debugLog("sync_service.go:445", "before syncFromGCP", map[string]interface{}{
			"sourceId": source.ID, "bucket": source.Config.GCPBucket, "objectName": source.Config.GCPObjectName,
		})
		// #endregion
		if err := s.syncFromGCP(ctx, source, job); err != nil {
			// #region agent log
			debugLog("sync_service.go:448", "syncFromGCP failed", map[string]interface{}{
				"sourceId": source.ID, "error": err.Error(),
			})
			// #endregion
			job.Status = "failed"
			job.Error = err.Error()
			return
		}
		// #region agent log
		debugLog("sync_service.go:455", "syncFromGCP success", map[string]interface{}{
			"sourceId": source.ID, "jobMessage": job.Message,
		})
		// #endregion
	case "terraform_cloud":
		job.Status = "failed"
		job.Error = "synchronisation depuis Terraform Cloud non encore implémentée"
		return
	default:
		job.Status = "failed"
		job.Error = fmt.Sprintf("type de source non supporté: %s", source.Type)
		return
	}

	// Mettre à jour la source
	source.LastSync = time.Now()
	if source.Config.SyncInterval != "" {
		if duration, err := time.ParseDuration(source.Config.SyncInterval); err == nil {
			source.NextSync = time.Now().Add(duration)
		}
	}
	source.UpdatedAt = time.Now()

	s.mu.Lock()
	s.sources[source.ID] = source
	s.mu.Unlock()

	s.saveSourceToCache(ctx, source)

	job.Message = "Synchronisation réussie"
}

// syncFromS3 synchronise depuis S3.
func (s *SyncService) syncFromS3(ctx context.Context, source *models.StateSource, job *models.SyncJob) error {
	config := source.Config

	// Créer le client S3
	s3Client, err := s3.NewClient(
		config.S3Region,
		config.S3Endpoint,
		config.AWSAccessKeyID,
		config.AWSSecretAccessKey,
	)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du client S3: %w", err)
	}

	// Tester la connexion
	if err := s3Client.TestConnection(ctx, config.S3Bucket); err != nil {
		return fmt.Errorf("erreur de connexion S3: %w", err)
	}

	// Récupérer le fichier tfstate
	stateData, err := s3Client.GetStateFile(ctx, config.S3Bucket, config.S3Key)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du fichier: %w", err)
	}

	// Parser et mettre à jour l'état (utiliser l'ID existant pour la mise à jour)
	stateFile, err := s.terraformService.ParseStateFileWithID(ctx, source.StateFileID, "", stateData, "")
	if err != nil {
		return fmt.Errorf("erreur lors du parsing: %w", err)
	}

	job.Message = fmt.Sprintf("État synchronisé: %s (version %d, %d ressources)",
		stateFile.Name, stateFile.State.Version, len(stateFile.State.Resources))

	return nil
}

// syncFromAzure synchronise depuis Azure Blob Storage.
func (s *SyncService) syncFromAzure(ctx context.Context, source *models.StateSource, job *models.SyncJob) error {
	config := source.Config

	// Créer le client Azure
	azureClient, err := azureStorage.NewClient(
		config.AzureAccountName,
		config.AzureAccountKey,
		config.AzureConnectionString,
	)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du client Azure: %w", err)
	}

	// Tester la connexion
	if err := azureClient.TestConnection(ctx, config.AzureContainer); err != nil {
		return fmt.Errorf("erreur de connexion Azure: %w", err)
	}

	// Récupérer le fichier tfstate
	stateData, err := azureClient.GetStateFile(ctx, config.AzureContainer, config.AzureBlobName)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du fichier: %w", err)
	}

	// Parser et mettre à jour l'état
	stateFile, err := s.terraformService.ParseStateFileWithID(ctx, source.StateFileID, "", stateData, "")
	if err != nil {
		return fmt.Errorf("erreur lors du parsing: %w", err)
	}

	job.Message = fmt.Sprintf("État synchronisé depuis Azure: %s (version %d, %d ressources)",
		stateFile.Name, stateFile.State.Version, len(stateFile.State.Resources))

	return nil
}

// syncFromGCP synchronise depuis GCP Cloud Storage.
func (s *SyncService) syncFromGCP(ctx context.Context, source *models.StateSource, job *models.SyncJob) error {
	// #region agent log
	debugLog("sync_service.go:555", "syncFromGCP entry", map[string]interface{}{
		"sourceId": source.ID, "stateFileID": source.StateFileID, "bucket": source.Config.GCPBucket, "objectName": source.Config.GCPObjectName,
	})
	// #endregion
	config := source.Config

	// Nettoyer les espaces dans le bucket et objectName
	config.GCPBucket = strings.TrimSpace(config.GCPBucket)
	config.GCPObjectName = strings.TrimSpace(config.GCPObjectName)

	// Si l'objectName commence par le nom du bucket, le retirer (correction d'erreur utilisateur)
	if strings.HasPrefix(config.GCPObjectName, config.GCPBucket+"/") {
		config.GCPObjectName = strings.TrimPrefix(config.GCPObjectName, config.GCPBucket+"/")
	}

	// #region agent log
	debugLog("sync_service.go:562", "after trim bucket/object", map[string]interface{}{
		"sourceId": source.ID, "bucket": config.GCPBucket, "objectName": config.GCPObjectName,
	})
	// #endregion

	// Créer le client GCP
	gcpClient, err := gcpStorage.NewClient(config.GCPCredentialsJSON)
	if err != nil {
		// #region agent log
		debugLog("sync_service.go:568", "NewClient failed", map[string]interface{}{
			"sourceId": source.ID, "error": err.Error(),
		})
		// #endregion
		return fmt.Errorf("erreur lors de la création du client GCP: %w", err)
	}

	// Tester la connexion
	if err := gcpClient.TestConnection(ctx, config.GCPBucket); err != nil {
		// #region agent log
		debugLog("sync_service.go:575", "TestConnection failed", map[string]interface{}{
			"sourceId": source.ID, "bucket": config.GCPBucket, "bucketLen": len(config.GCPBucket), "error": err.Error(),
		})
		// #endregion
		return fmt.Errorf("erreur de connexion GCP: %w", err)
	}

	// Récupérer le fichier tfstate
	// #region agent log
	debugLog("sync_service.go:570", "before GetStateFile", map[string]interface{}{
		"sourceId": source.ID, "bucket": config.GCPBucket, "objectName": config.GCPObjectName,
	})
	// #endregion
	stateData, err := gcpClient.GetStateFile(ctx, config.GCPBucket, config.GCPObjectName)
	if err != nil {
		// #region agent log
		debugLog("sync_service.go:573", "GetStateFile failed", map[string]interface{}{
			"sourceId": source.ID, "error": err.Error(),
		})
		// #endregion
		return fmt.Errorf("erreur lors de la récupération du fichier: %w", err)
	}
	// #region agent log
	debugLog("sync_service.go:575", "after GetStateFile success", map[string]interface{}{
		"sourceId": source.ID, "stateDataSize": len(stateData),
	})
	// #endregion

	// Déterminer le nom de l'état
	var stateName string
	// #region agent log
	debugLog("sync_service.go:576", "before stateName extraction", map[string]interface{}{
		"sourceId": source.ID, "stateFileID": source.StateFileID, "isTemp": strings.HasPrefix(source.StateFileID, "temp-"),
	})
	// #endregion
	if strings.HasPrefix(source.StateFileID, "temp-") {
		// Extraire le nom depuis l'ID temporaire (format: temp-nom-1234567890)
		parts := strings.Split(source.StateFileID, "-")
		if len(parts) >= 3 {
			stateName = strings.Join(parts[1:len(parts)-1], "-")
		} else {
			// Fallback: utiliser le nom du fichier
			keyParts := strings.Split(config.GCPObjectName, "/")
			fileName := keyParts[len(keyParts)-1]
			stateName = strings.TrimSuffix(fileName, ".tfstate")
			if stateName == "" {
				stateName = "terraform-state"
			}
		}
	} else {
		// Vérifier si l'état existe pour obtenir son nom
		if existing, err := s.terraformService.GetStateFile(ctx, source.StateFileID); err == nil {
			stateName = existing.Name
		} else {
			// Nouvel état, extraire le nom du chemin
			keyParts := strings.Split(config.GCPObjectName, "/")
			fileName := keyParts[len(keyParts)-1]
			stateName = strings.TrimSuffix(fileName, ".tfstate")
			if stateName == "" {
				stateName = "terraform-state"
			}
		}
	}
	// Nettoyer les espaces dans le stateName
	stateName = strings.TrimSpace(stateName)

	// #region agent log
	debugLog("sync_service.go:603", "after stateName extraction", map[string]interface{}{
		"sourceId": source.ID, "stateName": stateName,
	})
	// #endregion

	// Parser et mettre à jour l'état
	// #region agent log
	debugLog("sync_service.go:606", "before ParseStateFileWithID", map[string]interface{}{
		"sourceId": source.ID, "stateFileID": source.StateFileID, "stateName": stateName, "stateDataSize": len(stateData),
	})
	// #endregion
	stateFile, err := s.terraformService.ParseStateFileWithID(ctx, source.StateFileID, stateName, stateData, "")
	if err != nil {
		// #region agent log
		debugLog("sync_service.go:610", "ParseStateFileWithID failed", map[string]interface{}{
			"sourceId": source.ID, "error": err.Error(),
		})
		// #endregion
		return fmt.Errorf("erreur lors du parsing: %w", err)
	}
	// #region agent log
	debugLog("sync_service.go:616", "after ParseStateFileWithID success", map[string]interface{}{
		"sourceId": source.ID, "stateFileID": stateFile.ID, "stateFileName": stateFile.Name, "resourceCount": len(stateFile.State.Resources), "oldStateFileID": source.StateFileID,
	})
	// #endregion

	// Détecter le drift automatiquement après synchronisation
	if stateFile != nil && stateFile.ID != "" {
		// #region agent log
		debugLog("sync_service.go:627", "before drift detection", map[string]interface{}{
			"sourceId": source.ID, "stateFileID": stateFile.ID, "sourceType": source.Type,
		})
		// #endregion

		// Décrypter les credentials pour la détection de drift
		sourceCopy := *source
		if err := s.decryptCredentials(&sourceCopy); err == nil {
			var credentialsJSON string
			if sourceCopy.Type == "gcp" && sourceCopy.Config.GCPCredentialsJSON != "" {
				credentialsJSON = sourceCopy.Config.GCPCredentialsJSON
			}

			// Effectuer la détection de drift
			driftResults, driftErr := s.terraformService.DetectDrift(ctx, stateFile.ID, credentialsJSON, sourceCopy.Type)
			if driftErr != nil {
				// Log l'erreur mais ne fait pas échouer la synchronisation
				// #region agent log
				debugLog("sync_service.go:642", "drift detection failed", map[string]interface{}{
					"sourceId": source.ID, "stateFileID": stateFile.ID, "error": driftErr.Error(),
				})
				// #endregion
			} else {
				// #region agent log
				driftCount := 0
				for _, r := range driftResults {
					if r.Status == "drifted" || r.Status == "missing" {
						driftCount++
					}
				}
				debugLog("sync_service.go:651", "drift detection completed", map[string]interface{}{
					"sourceId": source.ID, "stateFileID": stateFile.ID, "driftCount": driftCount, "totalResources": len(driftResults),
				})
				// #endregion
			}
		}
	}

	// Si l'ID était temporaire, mettre à jour la source avec le nouvel ID permanent
	if strings.HasPrefix(source.StateFileID, "temp-") && source.StateFileID != stateFile.ID {
		// #region agent log
		debugLog("sync_service.go:620", "updating source with new stateFileID", map[string]interface{}{
			"sourceId": source.ID, "oldStateFileID": source.StateFileID, "newStateFileID": stateFile.ID,
		})
		// #endregion
		// Mettre à jour la source avec le nouvel ID seulement si l'ID a changé
		if source.StateFileID != stateFile.ID {
			source.StateFileID = stateFile.ID
			s.mu.Lock()
			s.sources[source.ID] = source
			s.saveSourceToCache(ctx, source)
			s.mu.Unlock()
			// #region agent log
			debugLog("sync_service.go:627", "source updated with new stateFileID", map[string]interface{}{
				"sourceId": source.ID, "newStateFileID": source.StateFileID,
			})
			// #endregion
		} else {
			// #region agent log
			debugLog("sync_service.go:633", "stateFileID unchanged, no update needed", map[string]interface{}{
				"sourceId": source.ID, "stateFileID": source.StateFileID,
			})
			// #endregion
		}
	}

	job.Message = fmt.Sprintf("État synchronisé depuis GCP: %s (version %d, %d ressources)",
		stateFile.Name, stateFile.State.Version, len(stateFile.State.Resources))

	return nil
}

// startScheduler démarre le scheduler de synchronisation automatique.
func (s *SyncService) startScheduler() {
	// Vérifier toutes les minutes pour les synchronisations programmées
	s.syncTicker = time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-s.syncTicker.C:
				s.checkAndSync()
			case <-s.stopChan:
				s.syncTicker.Stop()
				return
			}
		}
	}()
}

// checkAndSync vérifie et synchronise les sources programmées.
func (s *SyncService) checkAndSync() {
	ctx := context.Background()
	now := time.Now()

	s.mu.RLock()
	sourcesToSync := make([]*models.StateSource, 0)
	for _, source := range s.sources {
		if source.Enabled && source.Config.AutoSync {
			// Vérifier si c'est le moment de synchroniser
			if source.NextSync.IsZero() || now.After(source.NextSync) {
				sourcesToSync = append(sourcesToSync, source)
			}
		}
	}
	s.mu.RUnlock()

	// Synchroniser chaque source
	for _, source := range sourcesToSync {
		log.Printf("🔄 Synchronisation automatique de la source %s (%s)", source.ID, source.Type)
		if _, err := s.SyncState(ctx, source.ID); err != nil {
			log.Printf("❌ Erreur lors de la synchronisation de %s: %v", source.ID, err)
		}
	}
}

// Stop arrête le service de synchronisation.
func (s *SyncService) Stop() {
	close(s.stopChan)
}

// encryptCredentials chiffre les credentials sensibles.
func (s *SyncService) encryptCredentials(source *models.StateSource) error {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Chiffrer AWS Secret Access Key
	if source.Config.AWSSecretAccessKey != "" {
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		ciphertext := gcm.Seal(nonce, nonce, []byte(source.Config.AWSSecretAccessKey), nil)
		source.Config.AWSSecretAccessKey = base64.StdEncoding.EncodeToString(ciphertext)
	}

	// Chiffrer Azure credentials
	if source.Config.AzureAccountKey != "" {
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		ciphertext := gcm.Seal(nonce, nonce, []byte(source.Config.AzureAccountKey), nil)
		source.Config.AzureAccountKey = base64.StdEncoding.EncodeToString(ciphertext)
	}
	if source.Config.AzureConnectionString != "" {
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		ciphertext := gcm.Seal(nonce, nonce, []byte(source.Config.AzureConnectionString), nil)
		source.Config.AzureConnectionString = base64.StdEncoding.EncodeToString(ciphertext)
	}

	// Chiffrer GCP credentials
	if source.Config.GCPCredentialsJSON != "" {
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		ciphertext := gcm.Seal(nonce, nonce, []byte(source.Config.GCPCredentialsJSON), nil)
		source.Config.GCPCredentialsJSON = base64.StdEncoding.EncodeToString(ciphertext)
	}

	// Chiffrer Terraform Cloud Token
	if source.Config.TerraformCloudToken != "" {
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		ciphertext := gcm.Seal(nonce, nonce, []byte(source.Config.TerraformCloudToken), nil)
		source.Config.TerraformCloudToken = base64.StdEncoding.EncodeToString(ciphertext)
	}

	return nil
}

// DecryptCredentials déchiffre les credentials d'une source (méthode publique).
func (s *SyncService) DecryptCredentials(source *models.StateSource) error {
	return s.decryptCredentials(source)
}

// decryptCredentials déchiffre les credentials.
func (s *SyncService) decryptCredentials(source *models.StateSource) error {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Déchiffrer AWS Secret Access Key
	if source.Config.AWSSecretAccessKey != "" {
		ciphertext, err := base64.StdEncoding.DecodeString(source.Config.AWSSecretAccessKey)
		if err != nil {
			return err
		}
		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return fmt.Errorf("ciphertext trop court")
		}
		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return err
		}
		source.Config.AWSSecretAccessKey = string(plaintext)
	}

	// Déchiffrer Terraform Cloud Token
	if source.Config.TerraformCloudToken != "" {
		ciphertext, err := base64.StdEncoding.DecodeString(source.Config.TerraformCloudToken)
		if err != nil {
			return err
		}
		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return fmt.Errorf("ciphertext trop court")
		}
		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return err
		}
		source.Config.TerraformCloudToken = string(plaintext)
	}

	// Déchiffrer Azure credentials
	if source.Config.AzureAccountKey != "" {
		ciphertext, err := base64.StdEncoding.DecodeString(source.Config.AzureAccountKey)
		if err != nil {
			return err
		}
		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return fmt.Errorf("ciphertext trop court")
		}
		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return err
		}
		source.Config.AzureAccountKey = string(plaintext)
	}
	if source.Config.AzureConnectionString != "" {
		ciphertext, err := base64.StdEncoding.DecodeString(source.Config.AzureConnectionString)
		if err != nil {
			return err
		}
		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return fmt.Errorf("ciphertext trop court")
		}
		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return err
		}
		source.Config.AzureConnectionString = string(plaintext)
	}

	// Déchiffrer GCP credentials
	if source.Config.GCPCredentialsJSON != "" {
		// Vérifier si c'est déjà du JSON brut (non chiffré)
		var testJSON map[string]interface{}
		if json.Unmarshal([]byte(source.Config.GCPCredentialsJSON), &testJSON) == nil {
			// C'est déjà du JSON valide, pas besoin de déchiffrer
			// #region agent log
			debugLog("sync_service.go:953", "GCP credentials already in plaintext JSON", map[string]interface{}{
				"sourceId": source.ID,
			})
			// #endregion
			return nil
		}

		// Essayer de déchiffrer
		ciphertext, err := base64.StdEncoding.DecodeString(source.Config.GCPCredentialsJSON)
		if err != nil {
			return fmt.Errorf("erreur lors du décodage base64 des credentials GCP (peut-être déjà en texte brut?): %w", err)
		}
		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return fmt.Errorf("ciphertext trop court")
		}
		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return fmt.Errorf("erreur lors du déchiffrement des credentials GCP: %w", err)
		}
		source.Config.GCPCredentialsJSON = string(plaintext)
	}

	return nil
}

// saveSourceToCache sauvegarde une source dans le cache.
func (s *SyncService) saveSourceToCache(ctx context.Context, source *models.StateSource) {
	cacheKey := fmt.Sprintf("terraform:source:%s", source.ID)
	sourceJSON, err := json.Marshal(source)
	if err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(sourceJSON), 24*time.Hour)
	}

	// Sauvegarder aussi la liste des IDs
	s.saveSourcesList(ctx)
}

// loadSourcesFromCache charge les sources depuis le cache.
func (s *SyncService) loadSourcesFromCache(ctx context.Context) {
	listKey := "terraform:sources:list"
	idsJSON, err := s.cache.Get(ctx, listKey)
	if err != nil || idsJSON == "" {
		return
	}

	var ids []string
	if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil {
		return
	}

	for _, id := range ids {
		cacheKey := fmt.Sprintf("terraform:source:%s", id)
		cached, err := s.cache.Get(ctx, cacheKey)
		if err != nil || cached == "" {
			continue
		}

		var source models.StateSource
		if err := json.Unmarshal([]byte(cached), &source); err == nil {
			s.sources[source.ID] = &source
		}
	}
}

// saveSourcesList sauvegarde la liste des sources.
func (s *SyncService) saveSourcesList(ctx context.Context) {
	ids := make([]string, 0, len(s.sources))
	for id := range s.sources {
		ids = append(ids, id)
	}
	idsJSON, _ := json.Marshal(ids)
	_ = s.cache.Set(ctx, "terraform:sources:list", string(idsJSON), 0) // 0 = pas d'expiration pour les configurations
}

// deleteSourceFromCache supprime une source du cache.
func (s *SyncService) deleteSourceFromCache(ctx context.Context, id string) {
	cacheKey := fmt.Sprintf("terraform:source:%s", id)
	_ = s.cache.Delete(ctx, cacheKey)

	// Mettre à jour la liste des IDs
	s.saveSourcesList(ctx)
}
