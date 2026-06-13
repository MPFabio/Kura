package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/models"
)

// argoSessionTTL est la durée de vie du token de session mis en cache, légèrement
// inférieure à l'expiration par défaut du JWT ArgoCD (12h).
const argoSessionTTL = 11 * time.Hour

// ArgoCache définit l'interface minimale du cache utilisée pour la session ArgoCD.
type ArgoCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// ArgoCDService orchestre l'installation d'ArgoCD et le proxy vers son API REST.
type ArgoCDService struct {
	cache          ArgoCache
	clusterService *ClusterService
	cfg            *config.Config
	httpClient     *http.Client
}

// NewArgoCDService crée un nouveau service ArgoCD.
func NewArgoCDService(cache ArgoCache, clusterService *ClusterService, cfg *config.Config) *ArgoCDService {
	return &ArgoCDService{
		cache:          cache,
		clusterService: clusterService,
		cfg:            cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			// argocd-server est joint via un port-forward local (127.0.0.1) avec son
			// certificat TLS auto-signé par défaut : la vérification du certificat
			// n'apporte aucune garantie de sécurité supplémentaire dans ce contexte.
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// restConfigForActiveCluster construit une configuration REST pour le cluster actif.
// Pour les clusters GKE, on utilise un kubeconfig "portable" (token OAuth2 statique)
// car le plugin gke-gcloud-auth-plugin n'est pas disponible dans le conteneur k8s-service.
func (s *ArgoCDService) restConfigForActiveCluster(ctx context.Context) (*rest.Config, *kubernetes.Clientset, string, error) {
	cluster, err := s.clusterService.GetActiveCluster(ctx)
	if err != nil {
		return nil, nil, "", fmt.Errorf("aucun cluster actif: %w", err)
	}

	kubeconfigContent, err := s.clusterService.GetPortableKubeconfig(ctx, cluster)
	if err != nil {
		return nil, nil, "", fmt.Errorf("préparation du kubeconfig: %w", err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
	if err != nil {
		return nil, nil, "", fmt.Errorf("chargement du kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, "", fmt.Errorf("création du client Kubernetes: %w", err)
	}

	return restConfig, clientset, cluster.ID, nil
}

// Install installe ArgoCD sur le cluster actif.
func (s *ArgoCDService) Install(ctx context.Context) error {
	restConfig, clientset, _, err := s.restConfigForActiveCluster(ctx)
	if err != nil {
		return err
	}

	return k8s.InstallArgoCD(ctx, restConfig, clientset)
}

// GetStatus retourne l'état d'installation et de disponibilité d'ArgoCD sur le cluster actif.
func (s *ArgoCDService) GetStatus(ctx context.Context) (*models.ArgoCDStatus, error) {
	_, clientset, _, err := s.restConfigForActiveCluster(ctx)
	if err != nil {
		return nil, err
	}

	status := &models.ArgoCDStatus{}

	if _, err := clientset.CoreV1().Namespaces().Get(ctx, k8s.ArgoCDNamespace, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return status, nil
		}
		return nil, fmt.Errorf("vérification du namespace argocd: %w", err)
	}
	status.Installed = true

	deployment, err := clientset.AppsV1().Deployments(k8s.ArgoCDNamespace).Get(ctx, "argocd-server", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return status, nil
		}
		return nil, fmt.Errorf("vérification du déploiement argocd-server: %w", err)
	}

	status.ServerReady = deployment.Status.ReadyReplicas > 0
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == "argocd-server" {
			status.Version = c.Image
		}
	}

	return status, nil
}

// argoSession contient le proxy port-forward actif et le token associé pour une requête.
type argoSession struct {
	proxy *k8s.ArgoCDProxy
	token string
}

// close arrête le port-forward associé à la session.
func (sess *argoSession) close() {
	if sess.proxy != nil {
		sess.proxy.Stop()
	}
}

// openSession démarre un port-forward vers argocd-server et obtient un token de session valide.
func (s *ArgoCDService) openSession(ctx context.Context) (*argoSession, error) {
	restConfig, clientset, clusterID, err := s.restConfigForActiveCluster(ctx)
	if err != nil {
		return nil, err
	}

	proxy, err := k8s.NewArgoCDPortForwarder(restConfig, clientset)
	if err != nil {
		return nil, fmt.Errorf("connexion à argocd-server: %w", err)
	}

	cacheKey := fmt.Sprintf("argocd:session:%s", clusterID)

	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		return &argoSession{proxy: proxy, token: cached}, nil
	}

	token, err := s.login(ctx, proxy.BaseURL(), clientset)
	if err != nil {
		proxy.Stop()
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, token, argoSessionTTL)

	return &argoSession{proxy: proxy, token: token}, nil
}

// login récupère le mot de passe admin initial depuis le secret Kubernetes et ouvre une session ArgoCD.
func (s *ArgoCDService) login(ctx context.Context, baseURL string, clientset *kubernetes.Clientset) (string, error) {
	secret, err := clientset.CoreV1().Secrets(k8s.ArgoCDNamespace).Get(ctx, "argocd-initial-admin-secret", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("lecture du secret argocd-initial-admin-secret: %w", err)
	}

	password, ok := secret.Data["password"]
	if !ok || len(password) == 0 {
		return "", fmt.Errorf("mot de passe administrateur introuvable dans argocd-initial-admin-secret")
	}

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": string(password),
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/session", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("appel de /api/v1/session: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentification ArgoCD échouée (statut %d): %s", resp.StatusCode, string(respBody))
	}

	var sessionResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBody, &sessionResp); err != nil {
		return "", fmt.Errorf("réponse de session ArgoCD invalide: %w", err)
	}
	if sessionResp.Token == "" {
		return "", fmt.Errorf("token de session ArgoCD vide")
	}

	return sessionResp.Token, nil
}

// doRequest effectue une requête HTTP authentifiée contre l'API argocd-server,
// avec ré-authentification automatique en cas de réponse 401.
func (s *ArgoCDService) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	sess, err := s.openSession(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.close()

	respBody, status, err := s.callArgoCD(ctx, sess, method, path, body)
	if err != nil {
		return nil, err
	}

	if status == http.StatusUnauthorized {
		_, clientset, clusterID, cfgErr := s.restConfigForActiveCluster(ctx)
		if cfgErr != nil {
			return nil, cfgErr
		}
		_ = s.cache.Delete(ctx, fmt.Sprintf("argocd:session:%s", clusterID))

		token, loginErr := s.login(ctx, sess.proxy.BaseURL(), clientset)
		if loginErr != nil {
			return nil, loginErr
		}
		_ = s.cache.Set(ctx, fmt.Sprintf("argocd:session:%s", clusterID), token, argoSessionTTL)
		sess.token = token

		respBody, status, err = s.callArgoCD(ctx, sess, method, path, body)
		if err != nil {
			return nil, err
		}
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("%s", argoErrorMessage(status, respBody))
	}

	return respBody, nil
}

// argoErrorMessage extrait un message d'erreur lisible d'une réponse de l'API ArgoCD.
// L'API ArgoCD renvoie typiquement {"error": "...", "message": "...", "code": ...} :
// on remonte ce message directement plutôt que le JSON brut, pour que l'utilisateur
// comprenne pourquoi une action (sync, refresh, suppression...) a échoué.
func argoErrorMessage(status int, body []byte) string {
	var parsed struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if parsed.Message != "" {
			return parsed.Message
		}
		if parsed.Error != "" {
			return parsed.Error
		}
	}
	return fmt.Sprintf("l'API ArgoCD a répondu avec le statut %d: %s", status, string(body))
}

// callArgoCD effectue une unique requête HTTP vers l'API argocd-server proxifiée.
func (s *ArgoCDService) callArgoCD(ctx context.Context, sess *argoSession, method, path string, body interface{}) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("sérialisation de la requête: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, sess.proxy.BaseURL()+path, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+sess.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("appel de l'API ArgoCD %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, 0, fmt.Errorf("lecture de la réponse ArgoCD: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// argoApplicationItem représente un sous-ensemble des champs d'une Application ArgoCD
// retournés par l'API REST, suffisant pour construire les vues résumées et détaillées.
type argoApplicationItem struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Project string `json:"project"`
		Source  struct {
			RepoURL        string `json:"repoURL"`
			Path           string `json:"path"`
			TargetRevision string `json:"targetRevision"`
		} `json:"source"`
		Destination struct {
			Server    string `json:"server"`
			Namespace string `json:"namespace"`
		} `json:"destination"`
	} `json:"spec"`
	Status struct {
		Sync struct {
			Status string `json:"status"`
		} `json:"sync"`
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
		History []struct {
			ID         int64  `json:"id"`
			Revision   string `json:"revision"`
			DeployedAt string `json:"deployedAt"`
			Source     struct {
				RepoURL string `json:"repoURL"`
				Path    string `json:"path"`
			} `json:"source"`
		} `json:"history"`
	} `json:"status"`
}

// toArgoApplication convertit la réponse brute de l'API ArgoCD en modèle ModulOps.
func (item *argoApplicationItem) toArgoApplication() models.ArgoApplication {
	return models.ArgoApplication{
		Name:           item.Metadata.Name,
		Namespace:      item.Spec.Destination.Namespace,
		Project:        item.Spec.Project,
		SyncStatus:     item.Status.Sync.Status,
		HealthStatus:   item.Status.Health.Status,
		RepoURL:        item.Spec.Source.RepoURL,
		Path:           item.Spec.Source.Path,
		TargetRevision: item.Spec.Source.TargetRevision,
		DestNamespace:  item.Spec.Destination.Namespace,
		DestServer:     item.Spec.Destination.Server,
	}
}

// toArgoApplicationDetail convertit la réponse brute de l'API ArgoCD, historique inclus.
func (item *argoApplicationItem) toArgoApplicationDetail() models.ArgoApplicationDetail {
	detail := models.ArgoApplicationDetail{
		ArgoApplication: item.toArgoApplication(),
		History:         make([]models.ArgoHistoryEntry, 0, len(item.Status.History)),
	}
	for _, h := range item.Status.History {
		deployedAt, _ := time.Parse(time.RFC3339, h.DeployedAt)
		detail.History = append(detail.History, models.ArgoHistoryEntry{
			ID:         h.ID,
			RevisionID: h.Revision,
			DeployedAt: deployedAt,
			Source:     fmt.Sprintf("%s (%s)", h.Source.RepoURL, h.Source.Path),
		})
	}
	return detail
}

// ListApplications liste les Applications ArgoCD du cluster actif.
func (s *ArgoCDService) ListApplications(ctx context.Context) ([]models.ArgoApplication, error) {
	respBody, err := s.doRequest(ctx, http.MethodGet, "/api/v1/applications", nil)
	if err != nil {
		return nil, err
	}

	var listResp struct {
		Items []argoApplicationItem `json:"items"`
	}
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		return nil, fmt.Errorf("réponse ArgoCD invalide: %w", err)
	}

	apps := make([]models.ArgoApplication, 0, len(listResp.Items))
	for i := range listResp.Items {
		apps = append(apps, listResp.Items[i].toArgoApplication())
	}
	return apps, nil
}

// GetApplication retourne le détail d'une Application ArgoCD, historique inclus.
func (s *ArgoCDService) GetApplication(ctx context.Context, name string) (*models.ArgoApplicationDetail, error) {
	respBody, err := s.doRequest(ctx, http.MethodGet, "/api/v1/applications/"+name, nil)
	if err != nil {
		return nil, err
	}

	var item argoApplicationItem
	if err := json.Unmarshal(respBody, &item); err != nil {
		return nil, fmt.Errorf("réponse ArgoCD invalide: %w", err)
	}

	detail := item.toArgoApplicationDetail()
	return &detail, nil
}

// CreateApplication crée une nouvelle Application ArgoCD.
func (s *ArgoCDService) CreateApplication(ctx context.Context, req *models.CreateApplicationRequest) (*models.ArgoApplication, error) {
	project := req.Project
	if project == "" {
		project = "default"
	}
	destServer := req.DestServer
	if destServer == "" {
		destServer = "https://kubernetes.default.svc"
	}
	targetRevision := req.TargetRevision
	if targetRevision == "" {
		targetRevision = "HEAD"
	}

	// CreateNamespace=true : le namespace cible est créé automatiquement s'il n'existe
	// pas encore, pour éviter les Applications bloquées en OutOfSync faute de namespace.
	// ServerSideApply=true : certains charts (ex: kube-prometheus-stack) embarquent des
	// CRDs volumineuses dont l'annotation kubectl.kubernetes.io/last-applied-configuration
	// dépasse la limite de 262144 octets avec un apply classique, faisant échouer la
	// synchronisation ("metadata.annotations: Too long"). Le server-side apply n'utilise
	// pas cette annotation et n'a pas cette limite.
	syncPolicy := map[string]interface{}{
		"syncOptions": []string{"CreateNamespace=true", "ServerSideApply=true"},
	}
	if req.SyncPolicyAutomated {
		automated := map[string]interface{}{}
		if req.Prune {
			automated["prune"] = true
		}
		if req.SelfHeal {
			automated["selfHeal"] = true
		}
		syncPolicy["automated"] = automated
	}

	source := map[string]interface{}{
		"repoURL":        req.RepoURL,
		"targetRevision": targetRevision,
	}
	if req.SourceType == "helm" {
		source["chart"] = req.Chart
		if req.HelmValues != "" {
			source["helm"] = map[string]interface{}{"values": req.HelmValues}
		}
	} else {
		source["path"] = req.Path
	}

	body := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": req.Name,
		},
		"spec": map[string]interface{}{
			"project": project,
			"source":  source,
			"destination": map[string]interface{}{
				"server":    destServer,
				"namespace": req.DestNamespace,
			},
			"syncPolicy": syncPolicy,
		},
	}

	respBody, err := s.doRequest(ctx, http.MethodPost, "/api/v1/applications", body)
	if err != nil {
		return nil, err
	}

	var item argoApplicationItem
	if err := json.Unmarshal(respBody, &item); err != nil {
		return nil, fmt.Errorf("réponse ArgoCD invalide: %w", err)
	}

	app := item.toArgoApplication()
	return &app, nil
}

// SyncApplication déclenche une synchronisation d'une Application ArgoCD.
func (s *ArgoCDService) SyncApplication(ctx context.Context, name string, prune bool) error {
	body := map[string]interface{}{}
	if prune {
		body["prune"] = true
	}
	_, err := s.doRequest(ctx, http.MethodPost, "/api/v1/applications/"+name+"/sync", body)
	return err
}

// RefreshApplication force le rafraîchissement de l'état d'une Application ArgoCD.
func (s *ArgoCDService) RefreshApplication(ctx context.Context, name string) error {
	_, err := s.doRequest(ctx, http.MethodGet, "/api/v1/applications/"+name+"?refresh=normal", nil)
	return err
}

// mergeYAMLMaps fusionne récursivement override dans base (override est prioritaire).
func mergeYAMLMaps(base, override map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = map[string]interface{}{}
	}
	for k, v := range override {
		if overrideMap, ok := v.(map[string]interface{}); ok {
			if baseMap, ok := base[k].(map[string]interface{}); ok {
				base[k] = mergeYAMLMaps(baseMap, overrideMap)
				continue
			}
		}
		base[k] = v
	}
	return base
}

// UpdateApplicationValues fusionne les values Helm fournies (YAML) dans les values
// existantes de spec.source.helm.values d'une Application ArgoCD, puis déclenche une
// synchronisation pour appliquer le changement. Utilisé pour ajuster la configuration
// d'un chart déjà déployé (ex: désactiver des caches memcached trop gourmands pour un
// petit cluster) sans perdre les values déjà nécessaires (ex: bucketNames).
func (s *ArgoCDService) UpdateApplicationValues(ctx context.Context, name, valuesOverride string) error {
	respBody, err := s.doRequest(ctx, http.MethodGet, "/api/v1/applications/"+name, nil)
	if err != nil {
		return err
	}

	var app map[string]interface{}
	if err := json.Unmarshal(respBody, &app); err != nil {
		return fmt.Errorf("réponse ArgoCD invalide: %w", err)
	}

	spec, ok := app["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("réponse ArgoCD invalide: spec manquant")
	}
	source, ok := spec["source"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("réponse ArgoCD invalide: spec.source manquant")
	}
	helm, ok := source["helm"].(map[string]interface{})
	if !ok {
		helm = map[string]interface{}{}
	}

	var currentValues map[string]interface{}
	if existing, ok := helm["values"].(string); ok && existing != "" {
		if err := yaml.Unmarshal([]byte(existing), &currentValues); err != nil {
			return fmt.Errorf("values Helm existantes invalides: %w", err)
		}
	}

	var override map[string]interface{}
	if err := yaml.Unmarshal([]byte(valuesOverride), &override); err != nil {
		return fmt.Errorf("values Helm fournies invalides: %w", err)
	}

	merged := mergeYAMLMaps(currentValues, override)
	mergedBytes, err := yaml.Marshal(merged)
	if err != nil {
		return fmt.Errorf("sérialisation des values fusionnées: %w", err)
	}

	helm["values"] = string(mergedBytes)
	source["helm"] = helm

	if _, err := s.doRequest(ctx, http.MethodPut, "/api/v1/applications/"+name+"?validate=false", map[string]interface{}{
		"metadata": app["metadata"],
		"spec":     spec,
	}); err != nil {
		return err
	}

	return s.SyncApplication(ctx, name, false)
}

// RollbackApplication revient à une révision précédente de l'historique d'une Application ArgoCD.
func (s *ArgoCDService) RollbackApplication(ctx context.Context, name string, id int64) error {
	body := map[string]interface{}{
		"id": id,
	}
	_, err := s.doRequest(ctx, http.MethodPost, "/api/v1/applications/"+name+"/rollback", body)
	return err
}

// DeleteApplication supprime une Application ArgoCD ainsi que les ressources
// qu'elle a déployées dans le cluster (suppression en cascade).
func (s *ArgoCDService) DeleteApplication(ctx context.Context, name string) error {
	_, err := s.doRequest(ctx, http.MethodDelete, "/api/v1/applications/"+name+"?cascade=true", map[string]interface{}{})
	return err
}
