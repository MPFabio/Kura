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

	"github.com/modulops/k8s-service/internal/client"
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
	codeClient     *client.CodeServiceClient
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
		codeClient: client.NewCodeServiceClient(cfg.CodeServiceURL),
	}
}

// restConfigForActiveCluster construit une configuration REST pour le cluster actif.
// Pour les clusters GKE, on utilise un kubeconfig "portable" (token OAuth2 statique)
// car le plugin gke-gcloud-auth-plugin n'est pas disponible dans le conteneur k8s-service.
func (s *ArgoCDService) restConfigForActiveCluster(ctx context.Context) (*rest.Config, *kubernetes.Clientset, string, error) {
	cluster, _, _, err := s.activeClusterContext(ctx)
	if err != nil {
		return nil, nil, "", err
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

// activeClusterContext retourne le cluster actif ainsi que son ID et l'ID du projet
// auquel il appartient (utilisé pour le flux GitOps : dépôt GitOps = celui du projet).
func (s *ArgoCDService) activeClusterContext(ctx context.Context) (*models.Cluster, string, string, error) {
	cluster, err := s.clusterService.GetActiveCluster(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("aucun cluster actif: %w", err)
	}
	return cluster, cluster.ID, cluster.ProjectID, nil
}

// Install installe ArgoCD sur le cluster actif, puis amorce son auto-gestion GitOps
// ("app of apps") : un commit décrivant l'installation d'ArgoCD lui-même (chart Helm
// officiel argo-cd + values) est poussé sur le dépôt GitOps du projet, puis ArgoCD
// crée l'Application correspondante afin de se gérer lui-même via Git.
func (s *ArgoCDService) Install(ctx context.Context, authToken, branch, createBranchFrom string) error {
	restConfig, clientset, clusterID, err := s.restConfigForActiveCluster(ctx)
	if err != nil {
		return err
	}

	if err := k8s.InstallArgoCD(ctx, restConfig, clientset); err != nil {
		return err
	}

	return s.bootstrapSelfManagement(ctx, authToken, clusterID, branch, createBranchFrom)
}

// bootstrapSelfManagement committe les manifests d'auto-gestion d'ArgoCD (chart Helm
// officiel argo-cd + values reprenant les patches de résilience de argocd-repo-server)
// dans le dépôt GitOps du projet, puis crée l'Application "argocd" correspondante.
func (s *ArgoCDService) bootstrapSelfManagement(ctx context.Context, authToken, clusterID, branch, createBranchFrom string) error {
	_, _, projectID, err := s.activeClusterContext(ctx)
	if err != nil {
		return err
	}

	basePath := fmt.Sprintf("clusters/%s/argocd", clusterID)
	values := `# Values du chart Helm officiel argo-cd (https://argoproj.github.io/argo-helm),
# géré par ArgoCD lui-même (auto-gestion / "app of apps").
# Reprend les patches de résilience appliqués au bootstrap (cf. patchRepoServerResilience).
redis:
  # Sans ce ServiceAccount dédié, le conteneur d'initialisation du secret Redis
  # (secret-init) tourne avec le ServiceAccount "default" et échoue
  # (secrets is forbidden). Avec serviceAccount.create=true, le chart bascule
  # vers un Job dédié (RBAC correct) et retire cet init-container du Deployment.
  serviceAccount:
    create: true
repoServer:
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      memory: 512Mi
  livenessProbe:
    initialDelaySeconds: 90
    periodSeconds: 30
    timeoutSeconds: 10
    failureThreshold: 5
  readinessProbe:
    initialDelaySeconds: 60
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 5
`
	application := `apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: argocd
  namespace: argocd
spec:
  project: default
  sources:
    - repoURL: https://argoproj.github.io/argo-helm
      chart: argo-cd
      targetRevision: "*"
      helm:
        valueFiles:
          - $values/clusters/` + clusterID + `/argocd/values.yaml
    - repoURL: <dépôt GitOps du projet>
      targetRevision: ` + branch + `
      path: "."
      ref: values
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
`

	files := map[string]string{
		basePath + "/application.yaml": application,
		basePath + "/values.yaml":      values,
	}

	if err := s.codeClient.CommitGitOpsFiles(ctx, authToken, projectID, branch, createBranchFrom, files, "argocd: bootstrap auto-gestion (app of apps)"); err != nil {
		return fmt.Errorf("commit GitOps du bootstrap ArgoCD: %w", err)
	}

	gitopsRepo, err := s.gitopsRepoURL(ctx, authToken, projectID)
	if err != nil {
		return err
	}

	_, err = s.createApplicationDirect(ctx, &models.CreateApplicationRequest{
		Name:                "argocd",
		Project:             "default",
		SourceType:          "helm",
		RepoURL:             "https://argoproj.github.io/argo-helm",
		Chart:               "argo-cd",
		TargetRevision:      "*",
		DestNamespace:       k8s.ArgoCDNamespace,
		SyncPolicyAutomated: true,
		Prune:               true,
		SelfHeal:            true,
	}, &gitopsSource{
		RepoURL:        gitopsRepo,
		TargetRevision: branch,
		ValuesPath:     basePath + "/values.yaml",
	})
	if err != nil {
		return fmt.Errorf("création de l'Application argocd (auto-gestion): %w", err)
	}

	return nil
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

	// La présence du namespace seule ne suffit pas : une installation interrompue
	// (ex. context canceled en plein apply du manifest) peut laisser le namespace
	// créé sans que argocd-server ou le StatefulSet du controller existent. On ne
	// considère ArgoCD "installé" que si ces deux ressources sont présentes, afin
	// que le bouton d'installation reste disponible pour relancer une install partielle.
	deployment, err := clientset.AppsV1().Deployments(k8s.ArgoCDNamespace).Get(ctx, "argocd-server", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return status, nil
		}
		return nil, fmt.Errorf("vérification du déploiement argocd-server: %w", err)
	}

	if _, err := clientset.AppsV1().StatefulSets(k8s.ArgoCDNamespace).Get(ctx, "argocd-application-controller", metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return status, nil
		}
		return nil, fmt.Errorf("vérification du statefulset argocd-application-controller: %w", err)
	}

	status.Installed = true
	status.ServerReady = deployment.Status.ReadyReplicas > 0
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == "argocd-server" {
			status.Version = c.Image
		}
	}

	// L'auto-gestion (commit GitOps du bootstrap + Application "argocd") peut avoir
	// échoué après une installation par ailleurs réussie (ex. droits insuffisants sur
	// le dépôt GitOps) : on vérifie ici la présence de l'Application "argocd" pour que
	// l'UI puisse proposer de relancer ce seul bootstrap sans tout réinstaller.
	if status.ServerReady {
		if _, err := s.GetApplication(ctx, "argocd"); err == nil {
			status.SelfManaged = true
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

// gitopsSource décrit la référence vers le dépôt GitOps à inclure dans une Application
// ArgoCD, soit comme source unique (manifests Git purs), soit comme second "source"
// (multi-source, ≥ ArgoCD 2.6) fournissant le fichier de values Helm d'un chart catalogué.
type gitopsSource struct {
	RepoURL        string
	TargetRevision string
	// ValuesPath est le chemin du fichier values.yaml dans le dépôt GitOps, utilisé en
	// multi-source via la référence "$values" (source Helm uniquement).
	ValuesPath string
}

// createApplicationDirect crée une Application ArgoCD via l'API REST d'ArgoCD.
//
// Si gitops est nil, l'Application est créée telle que décrite par req (comportement
// historique, source = req.RepoURL/req.Path ou req.RepoURL/req.Chart).
//
// Si gitops est non-nil :
//   - pour une source "git", repoURL/path pointent directement sur le dépôt GitOps
//     (pull pur depuis le repo poussé par Kura) ;
//   - pour une source "helm", une Application multi-source est créée : le chart
//     catalogué (req.RepoURL/req.Chart) reste la première source, et le dépôt GitOps
//     est ajouté comme seconde source référencée "$values", dont le fichier
//     gitops.ValuesPath fournit les values Helm.
func (s *ArgoCDService) createApplicationDirect(ctx context.Context, req *models.CreateApplicationRequest, gitops *gitopsSource) (*models.ArgoApplication, error) {
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

	spec := map[string]interface{}{
		"project": project,
		"destination": map[string]interface{}{
			"server":    destServer,
			"namespace": req.DestNamespace,
		},
		"syncPolicy": syncPolicy,
	}

	if req.SourceType == "helm" {
		helm := map[string]interface{}{}
		if req.HelmValues != "" {
			helm["values"] = req.HelmValues
		}

		chartSource := map[string]interface{}{
			"repoURL":        req.RepoURL,
			"chart":          req.Chart,
			"targetRevision": targetRevision,
		}

		if gitops == nil {
			if len(helm) > 0 {
				chartSource["helm"] = helm
			}
			spec["source"] = chartSource
		} else {
			helm["valueFiles"] = []string{"$values/" + gitops.ValuesPath}
			chartSource["helm"] = helm
			spec["sources"] = []map[string]interface{}{
				chartSource,
				{
					"repoURL":        gitops.RepoURL,
					"targetRevision": gitops.TargetRevision,
					"path":           ".",
					"ref":            "values",
				},
			}
		}
	} else {
		if gitops == nil {
			spec["source"] = map[string]interface{}{
				"repoURL":        req.RepoURL,
				"path":           req.Path,
				"targetRevision": targetRevision,
			}
		} else {
			spec["source"] = map[string]interface{}{
				"repoURL":        gitops.RepoURL,
				"path":           req.Path,
				"targetRevision": gitops.TargetRevision,
			}
		}
	}

	body := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": req.Name,
		},
		"spec": spec,
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

// CreateApplicationViaGitOps committe le manifest et les values (le cas échéant) de
// l'Application demandée dans le dépôt GitOps du projet (sur la branche choisie), puis
// crée l'Application ArgoCD correspondante en référençant ce dépôt :
//   - source "git" : repoURL/path pointent directement sur apps/<name> du dépôt GitOps
//     (pull pur) ;
//   - source "helm" : Application multi-source, le chart catalogué reste la source
//     applicative, et apps/<name>/values.yaml du dépôt GitOps fournit les values Helm.
func (s *ArgoCDService) CreateApplicationViaGitOps(ctx context.Context, authToken string, req *models.CreateApplicationRequest) (*models.ArgoApplication, error) {
	_, _, projectID, err := s.activeClusterContext(ctx)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("apps/%s", req.Name)
	files := map[string]string{
		basePath + "/application.yaml": renderApplicationManifest(req),
	}
	if req.SourceType == "helm" {
		values := req.HelmValues
		if values == "" {
			values = "# values Helm (vide)\n"
		}
		files[basePath+"/values.yaml"] = values
	}

	message := fmt.Sprintf("argocd: ajout de l'Application %s", req.Name)
	if err := s.codeClient.CommitGitOpsFiles(ctx, authToken, projectID, req.Branch, req.CreateBranchFrom, files, message); err != nil {
		return nil, fmt.Errorf("commit GitOps de l'Application %s: %w", req.Name, err)
	}

	gitopsRepoURL, err := s.gitopsRepoURL(ctx, authToken, projectID)
	if err != nil {
		return nil, err
	}

	gitops := &gitopsSource{
		RepoURL:        gitopsRepoURL,
		TargetRevision: req.Branch,
		ValuesPath:     basePath + "/values.yaml",
	}
	if req.SourceType != "helm" {
		gitops = &gitopsSource{RepoURL: gitopsRepoURL, TargetRevision: req.Branch}
		req = &models.CreateApplicationRequest{
			Name: req.Name, Project: req.Project, SourceType: req.SourceType,
			Path: basePath, DestNamespace: req.DestNamespace, DestServer: req.DestServer,
			SyncPolicyAutomated: req.SyncPolicyAutomated, Prune: req.Prune, SelfHeal: req.SelfHeal,
		}
	}

	return s.createApplicationDirect(ctx, req, gitops)
}

// renderApplicationManifest produit le manifest YAML "argoproj.io/v1alpha1/Application"
// correspondant à req, committé dans le dépôt GitOps comme trace/audit de la demande.
func renderApplicationManifest(req *models.CreateApplicationRequest) string {
	targetRevision := req.TargetRevision
	if targetRevision == "" {
		targetRevision = "HEAD"
	}
	destServer := req.DestServer
	if destServer == "" {
		destServer = "https://kubernetes.default.svc"
	}
	project := req.Project
	if project == "" {
		project = "default"
	}

	var source string
	if req.SourceType == "helm" {
		source = fmt.Sprintf(`    repoURL: %s
    chart: %s
    targetRevision: %q
    helm:
      valueFiles:
        - values.yaml`, req.RepoURL, req.Chart, targetRevision)
	} else {
		source = fmt.Sprintf(`    repoURL: %s
    path: %s
    targetRevision: %q`, req.RepoURL, req.Path, targetRevision)
	}

	return fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: argocd
spec:
  project: %s
  source:
%s
  destination:
    server: %s
    namespace: %s
`, req.Name, project, source, destServer, req.DestNamespace)
}

// GetGitOpsInfo retourne les informations (URL de clone, nom complet, branches) du dépôt
// GitOps du projet associé au cluster actif, pour peupler le sélecteur de branche du
// frontend.
func (s *ArgoCDService) GetGitOpsInfo(ctx context.Context, authToken string) (*client.GitOpsInfo, error) {
	_, _, projectID, err := s.activeClusterContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.codeClient.GetGitOpsInfo(ctx, authToken, projectID)
}

// gitopsRepoURL retourne l'URL HTTPS de clone du dépôt GitOps du projet, en s'assurant
// qu'il existe (création éventuelle déléguée à code-service via EnsureGitOpsRepo).
func (s *ArgoCDService) gitopsRepoURL(ctx context.Context, authToken, projectID string) (string, error) {
	info, err := s.codeClient.GetGitOpsInfo(ctx, authToken, projectID)
	if err != nil {
		return "", fmt.Errorf("préparation du dépôt GitOps: %w", err)
	}
	return info.CloneURL, nil
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
