package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/modulops/k8s-service/internal/models"
)

// Cache définit l'interface minimale du cache pour la migration
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

// SimpleClusterMigration migre les clusters existants vers un projet par défaut
// Cette version simplifiée assigne tous les clusters à un projet par défaut
func SimpleClusterMigration(ctx context.Context, cache Cache, defaultProjectID string) error {
	if defaultProjectID == "" {
		// Ne pas migrer si aucun project_id n'est fourni
		log.Println("ℹ️  Aucun project_id fourni, migration des clusters ignorée")
		return nil
	}

	log.Println("🔄 Démarrage de la migration simple des clusters...")

	// Récupérer tous les clusters depuis Redis (ancien format sans project_id)
	listKey := "k8s:clusters:list"
	idsJSON, err := cache.Get(ctx, listKey)
	if err != nil || idsJSON == "" {
		log.Println("ℹ️  Aucun cluster trouvé dans le cache, migration non nécessaire")
		return nil
	}

	var clusterIDs []string
	if err := json.Unmarshal([]byte(idsJSON), &clusterIDs); err != nil {
		return fmt.Errorf("erreur lors du parsing de la liste des clusters: %w", err)
	}

	log.Printf("📋 %d cluster(s) trouvé(s), migration vers le projet %s...", len(clusterIDs), defaultProjectID)

	migratedCount := 0
	for _, clusterID := range clusterIDs {
		// Charger le cluster (ancien format)
		oldCacheKey := fmt.Sprintf("k8s:cluster:%s", clusterID)
		cached, err := cache.Get(ctx, oldCacheKey)
		if err != nil || cached == "" {
			log.Printf("⚠️  Cluster %s non trouvé dans le cache", clusterID)
			continue
		}

		var cluster models.Cluster
		if err := json.Unmarshal([]byte(cached), &cluster); err != nil {
			log.Printf("⚠️  Erreur lors du parsing du cluster %s: %v", clusterID, err)
			continue
		}

		// Si le cluster a déjà un project_id, le laisser tel quel
		if cluster.ProjectID != "" {
			log.Printf("ℹ️  Cluster %s a déjà un project_id (%s), ignoré", clusterID, cluster.ProjectID)
			continue
		}

		// Assigner le project_id
		cluster.ProjectID = defaultProjectID

		// Sauvegarder avec le nouveau format (incluant project_id dans la clé)
		newCacheKey := fmt.Sprintf("k8s:cluster:%s:%s", cluster.ProjectID, cluster.ID)
		clusterJSON, err := json.Marshal(cluster)
		if err != nil {
			log.Printf("⚠️  Erreur lors de la sérialisation du cluster %s: %v", clusterID, err)
			continue
		}

		if err := cache.Set(ctx, newCacheKey, string(clusterJSON), 0); err != nil {
			log.Printf("⚠️  Erreur lors de la sauvegarde du cluster %s: %v", clusterID, err)
			continue
		}

		migratedCount++
		log.Printf("✅ Cluster %s migré vers le projet %s", clusterID, cluster.ProjectID)
	}

	log.Printf("✅ Migration terminée: %d cluster(s) migré(s)", migratedCount)
	return nil
}
