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
	"github.com/modulops/k8s-service/internal/configstore"
	"github.com/modulops/k8s-service/internal/models"
)

// ClusterCache définit l'interface du cache pour les clusters.
type ClusterCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// ClusterService gère les clusters Kubernetes.
type ClusterService struct {
	cache           ClusterCache
	cfg             *config.Config
	cfgStore        *configstore.Client
	clusters        map[string]*models.Cluster
	activeClusterID string
}

// NewClusterService crée un nouveau service de gestion de clusters.
func NewClusterService(cache ClusterCache, cfg *config.Config) *ClusterService {
	cs := &ClusterService{
		cache:    cache,
		cfg:      cfg,
		cfgStore: configstore.New(cfg.AuthServiceURL, "k8s"),
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
	if cluster.ClusterType == "" {
		cluster.ClusterType = models.ClusterTypeGeneric
	}
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
		clusterCopy := *cluster
		if len(clusterCopy.Kubeconfig) > 100 {
			clusterCopy.Kubeconfig = clusterCopy.Kubeconfig[:100] + "..."
		}
		clusterCopy.CloudCredentials = "" // Ne pas exposer en liste
		result = append(result, &clusterCopy)
	}
	return result, nil
}

// ListClustersByProject retourne la liste des clusters d'un projet.
func (s *ClusterService) ListClustersByProject(ctx context.Context, projectID string) ([]*models.Cluster, error) {
	result := make([]*models.Cluster, 0)
	for _, cluster := range s.clusters {
		if cluster.ProjectID == projectID {
			clusterCopy := *cluster
			if len(clusterCopy.Kubeconfig) > 100 {
				clusterCopy.Kubeconfig = clusterCopy.Kubeconfig[:100] + "..."
			}
			clusterCopy.CloudCredentials = "" // Ne pas exposer en liste
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
	if cluster.ClusterType != "" {
		existing.ClusterType = cluster.ClusterType
	}
	if cluster.CloudCredentials != "" {
		existing.CloudCredentials = cluster.CloudCredentials
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

// WriteGCPCredentialsToTempFile écrit les credentials GCP du cluster dans un fichier temporaire
// pour que le plugin gke-gcloud-auth-plugin puisse s'authentifier. Retourne le chemin ou "" si non GKE/sans creds.
func (s *ClusterService) WriteGCPCredentialsToTempFile(cluster *models.Cluster) (string, error) {
	if cluster.ClusterType != models.ClusterTypeGKE || cluster.CloudCredentials == "" {
		return "", nil
	}
	tmpDir := "/tmp/kubeconfigs"
	os.MkdirAll(tmpDir, 0755)
	filePath := filepath.Join(tmpDir, fmt.Sprintf("gcp-sa-%s.json", cluster.ID))
	if err := os.WriteFile(filePath, []byte(cluster.CloudCredentials), 0600); err != nil {
		return "", fmt.Errorf("écriture des credentials GCP: %w", err)
	}
	return filePath, nil
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

// rewriteKubeconfigServerForDocker remplace localhost/127.0.0.1/0.0.0.0 par host.docker.internal
// dans les URLs server du kubeconfig pour que le conteneur puisse joindre un cluster sur l'hôte.
func rewriteKubeconfigServerForDocker(content string) string {
	out := content
	// 0.0.0.0 n'est pas une adresse valide pour se connecter (c'est pour bind), on la remplace aussi
	out = strings.ReplaceAll(out, "https://0.0.0.0:", "https://host.docker.internal:")
	out = strings.ReplaceAll(out, "http://0.0.0.0:", "http://host.docker.internal:")
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

	// Rejeter un kubeconfig vide pour éviter l'erreur client-go "no configuration has been provided"
	if strings.TrimSpace(kubeconfigContent) == "" {
		return "", fmt.Errorf("kubeconfig manquant ou vide pour le cluster %q: ajoutez ou collez le contenu du fichier kubeconfig (YAML) dans l'onglet Clusters", cluster.Name)
	}
	// Détecter un kubeconfig tronqué (ex. sortie de "kubectl config view" sans --raw : DATA+OMITTED)
	if strings.Contains(kubeconfigContent, "DATA+OMITTED") {
		return "", fmt.Errorf("kubeconfig invalide: les certificats sont masqués (DATA+OMITTED). Utilisez le contenu brut du fichier (Get-Content $env:USERPROFILE\\.kube\\config -Raw) ou kubectl config view --raw, puis collez-le dans l'onglet Clusters")
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

// loadClustersFromCache charge les clusters depuis Redis, avec fallback Postgres.
func (s *ClusterService) loadClustersFromCache(ctx context.Context) {
	listKey := "k8s:clusters:list"
	raw, err := s.cache.Get(ctx, listKey)
	if err != nil || raw == "" {
		// Fallback : charger depuis Postgres (configstore)
		log.Printf("⚠️  Redis vide, tentative de restauration des clusters depuis Postgres...")
		s.loadClustersFromConfigstore(ctx)
		return
	}

	var entries []string
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		log.Printf("⚠️  Erreur lors du parsing de la liste des clusters: %v", err)
		return
	}

	log.Printf("📦 Chargement de %d cluster(s) depuis Redis...", len(entries))

	for _, entry := range entries {
		var cached string
		if idx := strings.Index(entry, ":"); idx >= 0 {
			projectID, clusterID := entry[:idx], entry[idx+1:]
			cacheKey := fmt.Sprintf("k8s:cluster:%s:%s", projectID, clusterID)
			cached, _ = s.cache.Get(ctx, cacheKey)
		} else {
			// Ancien format : liste d'IDs ; la clé peut être k8s:cluster:id ou k8s:cluster:projectID:id
			keysToTry := []string{
				fmt.Sprintf("k8s:cluster:%s", entry),
				fmt.Sprintf("k8s:cluster::%s", entry),
			}
			if keys, _ := s.cache.Keys(ctx, "k8s:cluster:*:"+entry); len(keys) > 0 {
				keysToTry = append(keysToTry, keys[0])
			}
			for _, k := range keysToTry {
				cached, err = s.cache.Get(ctx, k)
				if err == nil && cached != "" {
					break
				}
			}
		}
		if cached == "" {
			log.Printf("⚠️  Cluster non trouvé pour l'entrée %q", entry)
			continue
		}

		var cluster models.Cluster
		if err := json.Unmarshal([]byte(cached), &cluster); err != nil {
			log.Printf("⚠️  Erreur lors du parsing du cluster: %v", err)
			continue
		}

		s.clusters[cluster.ID] = &cluster
		log.Printf("✅ Cluster chargé: %s (%s) - Projet: %s", cluster.ID, cluster.Name, cluster.ProjectID)
	}

	activeID, err := s.cache.Get(ctx, "k8s:clusters:active")
	if err == nil && activeID != "" {
		s.activeClusterID = activeID
		if cluster, exists := s.clusters[activeID]; exists {
			cluster.IsActive = true
			log.Printf("✅ Cluster actif restauré: %s (%s)", cluster.ID, cluster.Name)
		} else {
			log.Printf("⚠️  Cluster actif %s non trouvé dans la mémoire", activeID)
		}
	}
}

// saveClustersList sauvegarde la liste des clusters dans Redis et Postgres.
func (s *ClusterService) saveClustersList(ctx context.Context) {
	entries := make([]string, 0, len(s.clusters))
	for _, c := range s.clusters {
		entries = append(entries, c.ProjectID+":"+c.ID)
	}
	entriesJSON, _ := json.Marshal(entries)
	_ = s.cache.Set(ctx, "k8s:clusters:list", string(entriesJSON), 0)
	_ = s.cache.Set(ctx, "k8s:clusters:active", s.activeClusterID, 0)

	// Persistance Postgres pour survie aux redéploiements
	kv := map[string]string{
		"clusters_list":  string(entriesJSON),
		"clusters_active": s.activeClusterID,
	}
	for _, c := range s.clusters {
		if b, err := json.Marshal(c); err == nil {
			kv["cluster:"+c.ID] = string(b)
		}
	}
	if err := s.cfgStore.SetMany(ctx, kv); err != nil {
		log.Printf("⚠️  Persistance clusters Postgres: %v", err)
	}
}

// loadClustersFromConfigstore restaure les clusters depuis Postgres (après docker compose down/up).
func (s *ClusterService) loadClustersFromConfigstore(ctx context.Context) {
	all, err := s.cfgStore.GetAll(ctx)
	if err != nil {
		log.Printf("⚠️  Impossible de charger les clusters depuis Postgres: %v", err)
		return
	}
	listJSON, ok := all["clusters_list"]
	if !ok || listJSON == "" {
		log.Printf("⚠️  Aucun cluster dans Postgres")
		return
	}
	var entries []string
	if err := json.Unmarshal([]byte(listJSON), &entries); err != nil {
		log.Printf("⚠️  Erreur parsing clusters_list Postgres: %v", err)
		return
	}
	log.Printf("📦 Restauration de %d cluster(s) depuis Postgres...", len(entries))
	for _, entry := range entries {
		var clusterID string
		if idx := strings.LastIndex(entry, ":"); idx >= 0 {
			clusterID = entry[idx+1:]
		} else {
			clusterID = entry
		}
		raw, ok := all["cluster:"+clusterID]
		if !ok || raw == "" {
			continue
		}
		var cluster models.Cluster
		if err := json.Unmarshal([]byte(raw), &cluster); err != nil {
			log.Printf("⚠️  Erreur parsing cluster %s: %v", clusterID, err)
			continue
		}
		s.clusters[cluster.ID] = &cluster
		// Remettre dans Redis
		cacheKey := fmt.Sprintf("k8s:cluster:%s:%s", cluster.ProjectID, cluster.ID)
		_ = s.cache.Set(ctx, cacheKey, raw, 0)
		log.Printf("✅ Cluster restauré depuis Postgres: %s (%s)", cluster.ID, cluster.Name)
	}
	if activeID, ok := all["clusters_active"]; ok && activeID != "" {
		s.activeClusterID = activeID
		if cluster, exists := s.clusters[activeID]; exists {
			cluster.IsActive = true
		}
		_ = s.cache.Set(ctx, "k8s:clusters:active", activeID, 0)
	}
	// Remettre la liste dans Redis
	_ = s.cache.Set(ctx, "k8s:clusters:list", listJSON, 0)
}
