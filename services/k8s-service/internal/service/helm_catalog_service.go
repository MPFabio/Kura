package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/modulops/k8s-service/internal/models"
)

// helmCatalogCacheTTL est la durée de mise en cache des résultats de recherche
// ArtifactHub : ces données évoluent peu, un TTL d'une heure évite de solliciter
// l'API publique à chaque ouverture du catalogue.
const helmCatalogCacheTTL = 1 * time.Hour

const artifactHubSearchURL = "https://artifacthub.io/api/v1/packages/search"

// HelmCatalogService interroge le catalogue public ArtifactHub pour proposer
// des charts Helm prêts à déployer via ArgoCD.
type HelmCatalogService struct {
	cache      ArgoCache
	httpClient *http.Client
}

// NewHelmCatalogService crée un nouveau service de catalogue Helm.
func NewHelmCatalogService(cache ArgoCache) *HelmCatalogService {
	return &HelmCatalogService{
		cache: cache,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// artifactHubPackage représente un élément de la réponse de recherche ArtifactHub.
type artifactHubPackage struct {
	PackageID   string `json:"package_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	LogoImageID string `json:"logo_image_id"`
	Stars       int    `json:"stars"`
	HomeURL     string `json:"home_url"`
	Repository  struct {
		URL               string `json:"url"`
		Name              string `json:"name"`
		DisplayName       string `json:"display_name"`
		VerifiedPublisher bool   `json:"verified_publisher"`
	} `json:"repository"`
	CNCF     bool `json:"cncf"`
	Official bool `json:"official"`
}

type artifactHubSearchResponse struct {
	Packages []artifactHubPackage `json:"packages"`
}

// SearchCharts recherche des charts Helm via l'API ArtifactHub (kind=0), avec
// mise en cache des résultats. Si query est vide, retourne le catalogue par
// défaut filtré sur les charts officiels/CNCF.
func (s *HelmCatalogService) SearchCharts(ctx context.Context, query string, page int) ([]models.HelmChartSummary, error) {
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	cacheKey := fmt.Sprintf("helmcatalog:search:%s:%d", query, page)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []models.HelmChartSummary
		if jsonErr := json.Unmarshal([]byte(cached), &items); jsonErr == nil {
			return items, nil
		}
	}

	params := url.Values{}
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("facets", "false")
	params.Set("kind", "0")
	if query != "" {
		params.Set("ts_query_web", query)
		params.Set("sort", "relevance")
	} else {
		params.Set("sort", "stars")
		params.Set("cncf", "true")
	}

	reqURL := fmt.Sprintf("%s?%s", artifactHubSearchURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("création de la requête ArtifactHub: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel à ArtifactHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ArtifactHub a répondu avec le statut %d", resp.StatusCode)
	}

	var parsed artifactHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("réponse ArtifactHub invalide: %w", err)
	}

	items := make([]models.HelmChartSummary, 0, len(parsed.Packages))
	for _, pkg := range parsed.Packages {
		logoURL := ""
		if pkg.LogoImageID != "" {
			logoURL = fmt.Sprintf("https://artifacthub.io/image/%s?kind=0", pkg.LogoImageID)
		}
		items = append(items, models.HelmChartSummary{
			PackageID:   pkg.PackageID,
			Name:        pkg.Name,
			DisplayName: pkg.DisplayName,
			Description: pkg.Description,
			Version:     pkg.Version,
			LogoURL:     logoURL,
			RepoURL:     pkg.Repository.URL,
			RepoName:    pkg.Repository.DisplayName,
			Official:    pkg.Official,
			CNCF:        pkg.CNCF,
			Stars:       pkg.Stars,
			HomeURL:     pkg.HomeURL,
		})
	}

	if encoded, err := json.Marshal(items); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(encoded), helmCatalogCacheTTL)
	}

	return items, nil
}
