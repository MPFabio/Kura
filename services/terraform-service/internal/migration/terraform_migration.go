package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/modulops/terraform-service/internal/models"
)

// Cache définit l'interface minimale du cache pour la migration
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// SimpleTerraformMigration migre les terraform states existants vers un projet par défaut
func SimpleTerraformMigration(ctx context.Context, cache Cache, defaultProjectID string) error {
	if defaultProjectID == "" {
		// Ne pas migrer si aucun project_id n'est fourni
		log.Println("ℹ️  Aucun project_id fourni, migration des terraform states ignorée")
		return nil
	}

	log.Println("🔄 Démarrage de la migration simple des terraform states...")

	// Récupérer tous les terraform states depuis Redis
	keys, err := cache.Keys(ctx, "terraform:state:*")
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération des clés: %w", err)
	}

	if len(keys) == 0 {
		log.Println("ℹ️  Aucun terraform state trouvé dans le cache, migration non nécessaire")
		return nil
	}

	log.Printf("📋 %d terraform state(s) trouvé(s), migration vers le projet %s...", len(keys), defaultProjectID)

	migratedCount := 0
	for _, key := range keys {
		// Charger le state
		cached, err := cache.Get(ctx, key)
		if err != nil || cached == "" {
			log.Printf("⚠️  State %s non trouvé dans le cache", key)
			continue
		}

		var stateFile models.StateFile
		if err := json.Unmarshal([]byte(cached), &stateFile); err != nil {
			log.Printf("⚠️  Erreur lors du parsing du state %s: %v", key, err)
			continue
		}

		// Si le state a déjà un project_id, le laisser tel quel
		if stateFile.ProjectID != "" {
			log.Printf("ℹ️  State %s a déjà un project_id (%s), ignoré", stateFile.ID, stateFile.ProjectID)
			continue
		}

		// Assigner le project_id
		stateFile.ProjectID = defaultProjectID

		// Sauvegarder avec le nouveau format (incluant project_id dans la clé)
		newCacheKey := fmt.Sprintf("terraform:state:%s:%s", stateFile.ProjectID, stateFile.ID)
		stateJSON, err := json.Marshal(stateFile)
		if err != nil {
			log.Printf("⚠️  Erreur lors de la sérialisation du state %s: %v", stateFile.ID, err)
			continue
		}

		// Utiliser un TTL de 30 jours comme dans le service
		stateTTL := 30 * 24 * time.Hour
		if err := cache.Set(ctx, newCacheKey, string(stateJSON), stateTTL); err != nil {
			log.Printf("⚠️  Erreur lors de la sauvegarde du state %s: %v", stateFile.ID, err)
			continue
		}

		migratedCount++
		log.Printf("✅ State %s migré vers le projet %s", stateFile.ID, stateFile.ProjectID)
	}

	log.Printf("✅ Migration terminée: %d terraform state(s) migré(s)", migratedCount)
	return nil
}
