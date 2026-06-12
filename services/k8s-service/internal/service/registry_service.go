package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/models"
)

// registryCacheTTL est la durée de mise en cache du catalogue et des tags : courte
// car ces données évoluent à chaque push CI.
const registryCacheTTL = 30 * time.Second

const (
	helmChartConfigMediaType = "application/vnd.cncf.helm.config.v1+json"
)

// RegistryService interroge le registre OCI interne (Zot), déployé dans le cluster
// du client (namespace "zot"), pour lister les dépôts/tags disponibles et leur
// statut de signature Cosign. La connexion se fait via un port-forward vers le
// pod Zot, comme pour ArgoCD.
type RegistryService struct {
	cache          ArgoCache
	clusterService *ClusterService
	httpClient     *http.Client
}

// NewRegistryService crée un nouveau service de registre OCI.
func NewRegistryService(cache ArgoCache, clusterService *ClusterService, _ *config.Config) *RegistryService {
	return &RegistryService{
		cache:          cache,
		clusterService: clusterService,
		httpClient:     &http.Client{Timeout: 15 * time.Second},
	}
}

// restConfigForActiveCluster construit une configuration REST pour le cluster actif.
func (s *RegistryService) restConfigForActiveCluster(ctx context.Context) (*rest.Config, *kubernetes.Clientset, error) {
	cluster, err := s.clusterService.GetActiveCluster(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("aucun cluster actif: %w", err)
	}

	kubeconfigContent, err := s.clusterService.GetPortableKubeconfig(ctx, cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("préparation du kubeconfig: %w", err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
	if err != nil {
		return nil, nil, fmt.Errorf("chargement du kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("création du client Kubernetes: %w", err)
	}

	return restConfig, clientset, nil
}

// openSession ouvre un port-forward vers le pod Zot du cluster actif.
func (s *RegistryService) openSession(ctx context.Context) (*k8s.ZotProxy, error) {
	restConfig, clientset, err := s.restConfigForActiveCluster(ctx)
	if err != nil {
		return nil, err
	}

	proxy, err := k8s.NewZotPortForwarder(restConfig, clientset)
	if err != nil {
		return nil, fmt.Errorf("connexion au registre Zot: %w", err)
	}

	return proxy, nil
}

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

type tagsListResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type ociManifest struct {
	MediaType string `json:"mediaType"`
	Config    struct {
		MediaType string `json:"mediaType"`
	} `json:"config"`
	Layers []struct {
		Size int64 `json:"size"`
	} `json:"layers"`
}

// ListRepositories liste les dépôts disponibles dans le registre, avec leur nombre de tags.
func (s *RegistryService) ListRepositories(ctx context.Context) ([]models.RegistryRepository, error) {
	cacheKey := "registry:catalog"
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []models.RegistryRepository
		if jsonErr := json.Unmarshal([]byte(cached), &items); jsonErr == nil {
			return items, nil
		}
	}

	proxy, err := s.openSession(ctx)
	if err != nil {
		return nil, err
	}
	defer proxy.Stop()
	baseURL := proxy.BaseURL()

	var catalog catalogResponse
	if err := s.getJSON(ctx, baseURL, "/v2/_catalog", &catalog); err != nil {
		return nil, fmt.Errorf("appel au catalogue du registre: %w", err)
	}

	items := make([]models.RegistryRepository, 0, len(catalog.Repositories))
	for _, name := range catalog.Repositories {
		var tagsList tagsListResponse
		if err := s.getJSON(ctx, baseURL, fmt.Sprintf("/v2/%s/tags/list", name), &tagsList); err != nil {
			continue
		}
		items = append(items, models.RegistryRepository{
			Name:     name,
			TagCount: len(tagsList.Tags),
		})
	}

	if encoded, err := json.Marshal(items); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(encoded), registryCacheTTL)
	}

	return items, nil
}

// GetRepositoryDetail retourne le détail d'un dépôt : tags, taille, type
// (image ou chart Helm) et statut de signature Cosign.
func (s *RegistryService) GetRepositoryDetail(ctx context.Context, name string) (*models.RegistryRepositoryDetail, error) {
	cacheKey := fmt.Sprintf("registry:repo:%s:tags", name)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var detail models.RegistryRepositoryDetail
		if jsonErr := json.Unmarshal([]byte(cached), &detail); jsonErr == nil {
			return &detail, nil
		}
	}

	proxy, err := s.openSession(ctx)
	if err != nil {
		return nil, err
	}
	defer proxy.Stop()
	baseURL := proxy.BaseURL()

	var tagsList tagsListResponse
	if err := s.getJSON(ctx, baseURL, fmt.Sprintf("/v2/%s/tags/list", name), &tagsList); err != nil {
		return nil, fmt.Errorf("liste des tags pour %s: %w", name, err)
	}

	signedDigests := make(map[string]bool)
	tags := make([]models.RegistryTag, 0, len(tagsList.Tags))
	for _, tag := range tagsList.Tags {
		if strings.HasSuffix(tag, ".sig") {
			// Les tags de signature Cosign ("sha256-<hex>.sig") marquent le digest signé.
			digest := strings.TrimSuffix(tag, ".sig")
			signedDigests[digest] = true
			continue
		}

		manifest, digest, err := s.getManifest(ctx, baseURL, name, tag)
		if err != nil {
			continue
		}

		tagType := "image"
		if manifest.Config.MediaType == helmChartConfigMediaType {
			tagType = "helm-chart"
		}

		var size int64
		for _, layer := range manifest.Layers {
			size += layer.Size
		}

		tags = append(tags, models.RegistryTag{
			Name:      tag,
			Digest:    digest,
			MediaType: manifest.MediaType,
			SizeBytes: size,
			Type:      tagType,
		})
	}

	// Marquer les tags signés via les tags de signature Cosign collectés ci-dessus.
	for i := range tags {
		normalizedDigest := strings.ReplaceAll(tags[i].Digest, ":", "-")
		if signedDigests[normalizedDigest] {
			tags[i].Signed = true
		}
	}

	detail := &models.RegistryRepositoryDetail{Name: name, Tags: tags}

	if encoded, err := json.Marshal(detail); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(encoded), registryCacheTTL)
	}

	return detail, nil
}

// getManifest récupère le manifest OCI d'un tag et retourne le digest associé
// (depuis l'en-tête Docker-Content-Digest).
func (s *RegistryService) getManifest(ctx context.Context, baseURL, repo, tag string) (*ociManifest, string, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", baseURL, repo, tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("le registre a répondu avec le statut %d", resp.StatusCode)
	}

	var manifest ociManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, "", fmt.Errorf("manifest invalide: %w", err)
	}

	digest := resp.Header.Get("Docker-Content-Digest")
	return &manifest, digest, nil
}

func (s *RegistryService) getJSON(ctx context.Context, baseURL, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("le registre a répondu avec le statut %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
