package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/models"
)

// ClusterCache définit l'interface du cache pour les clusters.
type ClusterCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// ClusterService gère les clusters Kubernetes.
type ClusterService struct {
	cache           ClusterCache
	cfg             *config.Config
	clusters        map[string]*models.Cluster // Stockage en mémoire (pourrait être remplacé par PostgreSQL)
	activeClusterID string
}

// NewClusterService crée un nouveau service de gestion de clusters.
func NewClusterService(cache ClusterCache, cfg *config.Config) *ClusterService {
	cs := &ClusterService{
		cache:    cache,
		cfg:      cfg,
		clusters: make(map[string]*models.Cluster),
	}

	// Charger les clusters depuis le cache au démarrage
	// Utiliser un contexte avec timeout pour éviter les blocages
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("🔄 Initialisation du ClusterService, chargement depuis Redis...")
	cs.loadClustersFromCache(ctx)

	return cs
}

// CreateCluster crée un nouveau cluster.
func (s *ClusterService) CreateCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error) {
	// Générer un ID unique
	cluster.ID = fmt.Sprintf("cluster-%d", time.Now().UnixNano())
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()

	// Si c'est le premier cluster ou si is_active est true, le rendre actif
	if len(s.clusters) == 0 || cluster.IsActive {
		s.activeClusterID = cluster.ID
		// Désactiver les autres clusters
		for _, c := range s.clusters {
			c.IsActive = false
		}
		cluster.IsActive = true
	} else {
		cluster.IsActive = false
	}

	s.clusters[cluster.ID] = cluster

	// Sauvegarder dans le cache (pas d'expiration pour les configurations critiques)
	cacheKey := fmt.Sprintf("k8s:cluster:%s:%s", cluster.ProjectID, cluster.ID)
	clusterJSON, _ := json.Marshal(cluster)
	_ = s.cache.Set(ctx, cacheKey, string(clusterJSON), 0) // 0 = pas d'expiration

	// Sauvegarder la liste des clusters
	s.saveClustersList(ctx)

	return cluster, nil
}

// GetCluster récupère un cluster par son ID.
func (s *ClusterService) GetCluster(ctx context.Context, id string) (*models.Cluster, error) {
	// Fallback sur le stockage en mémoire (le cache utilise maintenant project_id)
	if cluster, exists := s.clusters[id]; exists {
		return cluster, nil
	}

	// Essayer de trouver dans tous les projets (pour compatibilité)
	for _, cluster := range s.clusters {
		if cluster.ID == id {
			return cluster, nil
		}
	}

	return nil, fmt.Errorf("cluster non trouvé: %s", id)
}

// ListClusters retourne la liste de tous les clusters (pour compatibilité).
func (s *ClusterService) ListClusters(ctx context.Context) ([]*models.Cluster, error) {
	result := make([]*models.Cluster, 0, len(s.clusters))
	for _, cluster := range s.clusters {
		// Ne pas exposer le kubeconfig complet dans la liste
		clusterCopy := *cluster
		if len(clusterCopy.Kubeconfig) > 100 {
			clusterCopy.Kubeconfig = clusterCopy.Kubeconfig[:100] + "..."
		}
		result = append(result, &clusterCopy)
	}
	return result, nil
}

// ListClustersByProject retourne la liste des clusters d'un projet.
func (s *ClusterService) ListClustersByProject(ctx context.Context, projectID string) ([]*models.Cluster, error) {
	result := make([]*models.Cluster, 0)
	for _, cluster := range s.clusters {
		if cluster.ProjectID == projectID {
			// Ne pas exposer le kubeconfig complet dans la liste
			clusterCopy := *cluster
			if len(clusterCopy.Kubeconfig) > 100 {
				clusterCopy.Kubeconfig = clusterCopy.Kubeconfig[:100] + "..."
			}
			result = append(result, &clusterCopy)
		}
	}
	return result, nil
}

// UpdateCluster met à jour un cluster.
func (s *ClusterService) UpdateCluster(ctx context.Context, id string, cluster *models.Cluster) (*models.Cluster, error) {
	existing, exists := s.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster non trouvé: %s", id)
	}

	// Mettre à jour les champs
	existing.Name = cluster.Name
	existing.Description = cluster.Description
	existing.Endpoint = cluster.Endpoint
	if cluster.Kubeconfig != "" {
		existing.Kubeconfig = cluster.Kubeconfig
	}
	existing.UpdatedAt = time.Now()

	// Gérer le statut actif
	if cluster.IsActive && s.activeClusterID != id {
		// Désactiver les autres clusters
		for _, c := range s.clusters {
			c.IsActive = false
		}
		existing.IsActive = true
		s.activeClusterID = id
	} else if !cluster.IsActive && s.activeClusterID == id {
		existing.IsActive = false
		s.activeClusterID = ""
	}

	// Sauvegarder dans le cache (pas d'expiration pour les configurations critiques)
	cacheKey := fmt.Sprintf("k8s:cluster:%s:%s", existing.ProjectID, id)
	clusterJSON, _ := json.Marshal(existing)
	_ = s.cache.Set(ctx, cacheKey, string(clusterJSON), 0) // 0 = pas d'expiration

	s.saveClustersList(ctx)

	return existing, nil
}

// DeleteCluster supprime un cluster.
func (s *ClusterService) DeleteCluster(ctx context.Context, id string) error {
	if _, exists := s.clusters[id]; !exists {
		return fmt.Errorf("cluster non trouvé: %s", id)
	}

	// Ne pas supprimer le cluster actif
	if s.activeClusterID == id {
		return fmt.Errorf("impossible de supprimer le cluster actif. Activez un autre cluster d'abord")
	}

	delete(s.clusters, id)

	// Supprimer du cache
	cacheKey := fmt.Sprintf("k8s:cluster:%s", id)
	_ = s.cache.Delete(ctx, cacheKey)

	s.saveClustersList(ctx)

	return nil
}

// SetActiveCluster définit le cluster actif.
func (s *ClusterService) SetActiveCluster(ctx context.Context, id string) error {
	cluster, exists := s.clusters[id]
	if !exists {
		return fmt.Errorf("cluster non trouvé: %s", id)
	}

	// Désactiver tous les clusters
	for _, c := range s.clusters {
		c.IsActive = false
	}

	// Activer le cluster sélectionné
	cluster.IsActive = true
	s.activeClusterID = id
	cluster.UpdatedAt = time.Now()

	// Sauvegarder
	cacheKey := fmt.Sprintf("k8s:cluster:%s:%s", cluster.ProjectID, id)
	clusterJSON, _ := json.Marshal(cluster)
	_ = s.cache.Set(ctx, cacheKey, string(clusterJSON), 24*time.Hour)

	s.saveClustersList(ctx)

	return nil
}

// GetActiveCluster retourne le cluster actif.
func (s *ClusterService) GetActiveCluster(ctx context.Context) (*models.Cluster, error) {
	if s.activeClusterID == "" {
		return nil, fmt.Errorf("aucun cluster actif")
	}

	return s.GetCluster(ctx, s.activeClusterID)
}

// SaveKubeconfigToFile sauvegarde un kubeconfig dans un fichier temporaire.
func (s *ClusterService) SaveKubeconfigToFile(ctx context.Context, clusterID string) (string, error) {
	cluster, err := s.GetCluster(ctx, clusterID)
	if err != nil {
		return "", err
	}

	return s.saveKubeconfigToTempFile(cluster)
}

// TestClusterConnection teste la connexion à un cluster en utilisant son kubeconfig.
func (s *ClusterService) TestClusterConnection(ctx context.Context, cluster *models.Cluster) (*models.ClusterStatus, error) {
	status := &models.ClusterStatus{
		ClusterID:   cluster.ID,
		Connected:   false,
		LastChecked: time.Now(),
	}

	if cluster.Kubeconfig == "" {
		status.Error = "kubeconfig manquant"
		return status, nil
	}

	// Validation de l'endpoint en production
	isProduction := s.cfg.Environment == "production"
	if isProduction && cluster.Endpoint != "" {
		if strings.Contains(cluster.Endpoint, "127.0.0.1") ||
			strings.Contains(cluster.Endpoint, "localhost") ||
			strings.Contains(cluster.Endpoint, "host.docker.internal") {
			status.Error = "endpoints locaux interdits en production"
			return status, nil
		}
	}

	// Créer un contexte avec timeout pour la connexion
	timeout := 30 * time.Second
	if s.cfg.K8sAPITimeout > 0 {
		timeout = s.cfg.K8sAPITimeout
	}
	testCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Sauvegarder le kubeconfig dans un fichier temporaire
	kubeconfigPath, err := s.saveKubeconfigToTempFile(cluster)
	if err != nil {
		status.Error = fmt.Sprintf("erreur lors de la sauvegarde du kubeconfig: %v", err)
		return status, nil
	}
	defer os.Remove(kubeconfigPath) // Nettoyer le fichier temporaire

	// Créer un client Kubernetes temporaire avec validation TLS en production
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		status.Error = fmt.Sprintf("erreur lors du chargement du kubeconfig: %v", err)
		return status, nil
	}

	// Configurer les timeouts et la validation TLS
	restConfig.Timeout = timeout
	if isProduction {
		restConfig.Insecure = false // Forcer la validation TLS en production
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		status.Error = fmt.Sprintf("erreur lors de la création du client Kubernetes: %v", err)
		return status, nil
	}

	// Tester la connexion en obtenant la version du serveur avec retry
	var version *version.Info
	maxRetries := s.cfg.K8sMaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	for i := 0; i < maxRetries; i++ {
		version, err = clientset.Discovery().ServerVersion()
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second) // Backoff exponentiel
		}
	}

	if err != nil {
		status.Error = fmt.Sprintf("impossible de se connecter au cluster après %d tentatives: %v", maxRetries, err)
		return status, nil
	}

	// Obtenir le nombre de nodes
	nodes, err := clientset.CoreV1().Nodes().List(testCtx, metav1.ListOptions{})
	if err != nil {
		// Ne pas échouer si on ne peut pas lister les nodes, mais noter l'erreur
		status.Error = fmt.Sprintf("connexion OK mais erreur lors de la récupération des nodes: %v", err)
	} else {
		status.NodesCount = len(nodes.Items)
	}

	status.Connected = true
	status.Version = version.GitVersion

	return status, nil
}

// rewriteKubeconfigServerForDocker remplace localhost/127.0.0.1 par host.docker.internal
// dans les URLs server du kubeconfig pour que le conteneur puisse joindre un cluster sur l'hôte.
func rewriteKubeconfigServerForDocker(content string) string {
	out := content
	out = strings.ReplaceAll(out, "https://127.0.0.1:", "https://host.docker.internal:")
	out = strings.ReplaceAll(out, "http://127.0.0.1:", "http://host.docker.internal:")
	out = strings.ReplaceAll(out, "https://localhost:", "https://host.docker.internal:")
	out = strings.ReplaceAll(out, "http://localhost:", "http://host.docker.internal:")
	return out
}

// saveKubeconfigToTempFile sauvegarde un kubeconfig dans un fichier temporaire.
func (s *ClusterService) saveKubeconfigToTempFile(cluster *models.Cluster) (string, error) {
	// Créer un répertoire temporaire pour les kubeconfigs
	tmpDir := "/tmp/kubeconfigs"
	os.MkdirAll(tmpDir, 0755)

	// Déterminer le contenu du kubeconfig
	var kubeconfigContent string
	if cluster.Kubeconfig != "" {
		// Essayer de décoder en base64
		if decoded, err := base64.StdEncoding.DecodeString(cluster.Kubeconfig); err == nil {
			kubeconfigContent = string(decoded)
		} else {
			kubeconfigContent = cluster.Kubeconfig
		}
	}

	// En dev/Docker : réécrire localhost/127.0.0.1 -> host.docker.internal pour que
	// le conteneur puisse joindre un cluster tournant sur l'hôte (minikube, k3d, etc.)
	if s.cfg.Environment != "production" && kubeconfigContent != "" {
		kubeconfigContent = rewriteKubeconfigServerForDocker(kubeconfigContent)
	}

	// Sauvegarder dans un fichier temporaire
	filePath := filepath.Join(tmpDir, fmt.Sprintf("%s-config-%d", cluster.ID, time.Now().UnixNano()))
	if err := os.WriteFile(filePath, []byte(kubeconfigContent), 0600); err != nil {
		return "", fmt.Errorf("erreur lors de l'écriture du kubeconfig: %w", err)
	}

	return filePath, nil
}

// loadClustersFromCache charge les clusters depuis le cache Redis.
func (s *ClusterService) loadClustersFromCache(ctx context.Context) {
	// Charger la liste des IDs de clusters
	listKey := "k8s:clusters:list"
	idsJSON, err := s.cache.Get(ctx, listKey)
	if err != nil || idsJSON == "" {
		log.Printf("⚠️  Aucun cluster trouvé dans le cache Redis")
		return // Pas de clusters en cache
	}

	var ids []string
	if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil {
		log.Printf("⚠️  Erreur lors du parsing de la liste des clusters: %v", err)
		return
	}

	log.Printf("📦 Chargement de %d cluster(s) depuis Redis...", len(ids))

	// Charger chaque cluster depuis le cache
	// Note: Le cache utilise maintenant project_id, donc on doit scanner tous les projets
	// Pour l'instant, on charge depuis la liste des IDs (ancien format)
	for _, id := range ids {
		// Essayer l'ancien format (sans project_id) pour compatibilité
		cacheKey := fmt.Sprintf("k8s:cluster:%s", id)
		cached, err := s.cache.Get(ctx, cacheKey)
		if err != nil || cached == "" {
			// Si pas trouvé, essayer de scanner par projet (nécessiterait une liste de projets)
			log.Printf("⚠️  Cluster %s non trouvé dans le cache (ancien format)", id)
			continue
		}

		var cluster models.Cluster
		if err := json.Unmarshal([]byte(cached), &cluster); err != nil {
			log.Printf("⚠️  Erreur lors du parsing du cluster %s: %v", id, err)
			continue
		}

		// Si le cluster n'a pas de project_id, on le laisse vide
		// Il sera assigné lors de la création via l'API avec le project_id de l'utilisateur

		s.clusters[cluster.ID] = &cluster
		log.Printf("✅ Cluster chargé: %s (%s) - Projet: %s", cluster.ID, cluster.Name, cluster.ProjectID)
	}

	// Charger l'ID du cluster actif
	activeID, err := s.cache.Get(ctx, "k8s:clusters:active")
	if err == nil && activeID != "" {
		s.activeClusterID = activeID
		// S'assurer que le cluster actif est marqué comme actif
		if cluster, exists := s.clusters[activeID]; exists {
			cluster.IsActive = true
			log.Printf("✅ Cluster actif restauré: %s (%s)", cluster.ID, cluster.Name)
		} else {
			log.Printf("⚠️  Cluster actif %s non trouvé dans la mémoire", activeID)
		}
	}
}

// saveClustersList sauvegarde la liste des clusters dans le cache.
func (s *ClusterService) saveClustersList(ctx context.Context) {
	ids := make([]string, 0, len(s.clusters))
	for id := range s.clusters {
		ids = append(ids, id)
	}
	idsJSON, _ := json.Marshal(ids)
	// Pas d'expiration pour les configurations critiques
	_ = s.cache.Set(ctx, "k8s:clusters:list", string(idsJSON), 0)     // 0 = pas d'expiration
	_ = s.cache.Set(ctx, "k8s:clusters:active", s.activeClusterID, 0) // 0 = pas d'expiration
}
