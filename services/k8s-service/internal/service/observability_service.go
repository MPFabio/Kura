package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/models"
)

// lokiOrgID est l'identifiant de tenant envoyé via l'en-tête X-Scope-OrgID.
// Les charts Helm Loki (mode "monolithic") activent auth_enabled par défaut ;
// "fake" est le tenant conventionnel pour un déploiement mono-tenant.
const lokiOrgID = "fake"

// ObservabilityService interroge la stack d'observabilité déployée dans le
// cluster du client (kube-prometheus-stack, Loki, Tempo via le catalogue
// Helm ArgoCD), en s'y connectant via un port-forward vers le cluster actif —
// comme pour ArgoCD et Zot. Contrairement au metrics-service (qui interroge
// la stack d'observabilité interne de la plateforme Kura), ce service expose
// les données du projet du client.
type ObservabilityService struct {
	clusterService *ClusterService
	httpClient     *http.Client

	// grafanaMu protège le port-forward Grafana réutilisé entre les requêtes
	// du reverse proxy (un iframe Grafana effectue de nombreuses requêtes
	// d'assets : ouvrir un port-forward par requête serait beaucoup trop lent).
	grafanaMu    sync.Mutex
	grafanaProxy *k8s.ObservabilityProxy
}

// NewObservabilityService crée un nouveau service d'observabilité projet.
func NewObservabilityService(clusterService *ClusterService) *ObservabilityService {
	return &ObservabilityService{
		clusterService: clusterService,
		httpClient:     &http.Client{},
	}
}

// restConfigForActiveCluster construit une configuration REST pour le cluster actif.
func (s *ObservabilityService) restConfigForActiveCluster(ctx context.Context) (*rest.Config, *kubernetes.Clientset, error) {
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

// openSession ouvre un port-forward vers le pod correspondant à la cible donnée.
func (s *ObservabilityService) openSession(ctx context.Context, target k8s.ObservabilityTarget) (*k8s.ObservabilityProxy, error) {
	restConfig, clientset, err := s.restConfigForActiveCluster(ctx)
	if err != nil {
		return nil, err
	}

	proxy, err := k8s.NewObservabilityPortForwarder(restConfig, clientset, target)
	if err != nil {
		return nil, err
	}

	return proxy, nil
}

// GetServiceMetrics retourne, pour chaque "job" Prometheus connu du cluster,
// les goroutines/CPU/mémoire courants (instant query), pour les onglets
// "Métriques" de l'observabilité projet.
func (s *ObservabilityService) GetServiceMetrics(ctx context.Context) (map[string]map[string]float64, error) {
	proxy, err := s.openSession(ctx, k8s.PrometheusTarget)
	if err != nil {
		return nil, err
	}
	defer proxy.Stop()
	baseURL := proxy.BaseURL()

	queries := map[string]string{
		"goroutines": "go_goroutines",
		"cpu_rate":   "rate(process_cpu_seconds_total[5m])",
		"memory_mb":  "process_resident_memory_bytes",
	}

	result := make(map[string]map[string]float64)
	for metric, query := range queries {
		values, err := s.queryPrometheus(ctx, baseURL, query)
		if err != nil {
			continue
		}
		for job, val := range values {
			if result[job] == nil {
				result[job] = make(map[string]float64)
			}
			if metric == "memory_mb" {
				val = val / 1024 / 1024
			}
			result[job][metric] = val
		}
	}

	return result, nil
}

// queryPrometheus exécute une instant query et retourne une map job → valeur float64.
func (s *ObservabilityService) queryPrometheus(ctx context.Context, baseURL, query string) (map[string]float64, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query", baseURL)
	params := url.Values{"query": {query}}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pr models.PrometheusResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, err
	}

	result := make(map[string]float64)
	for _, r := range pr.Data.Result {
		job := r.Metric["job"]
		if job == "" {
			job = r.Metric["app_kubernetes_io_name"]
		}
		if len(r.Value) < 2 {
			continue
		}
		valStr, ok := r.Value[1].(string)
		if !ok {
			continue
		}
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		result[job] = val
	}
	return result, nil
}

// GetLogs interroge Loki (du cluster client) et retourne les logs
// correspondant au service et à la recherche fournis.
func (s *ObservabilityService) GetLogs(ctx context.Context, service, search string, limit int) ([]models.LogEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}

	proxy, err := s.openSession(ctx, k8s.LokiTarget)
	if err != nil {
		return nil, err
	}
	defer proxy.Stop()

	selector := `{namespace=~".+"}`
	if service != "" {
		selector = fmt.Sprintf(`{namespace=~".+", pod=~"%s.*"}`, service)
	}
	query := selector
	if search != "" {
		query = fmt.Sprintf(`%s |= %q`, selector, search)
	}

	endpoint := fmt.Sprintf("%s/loki/api/v1/query_range", proxy.BaseURL())
	params := url.Values{
		"query":     {query},
		"limit":     {strconv.Itoa(limit)},
		"direction": {"backward"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Scope-OrgID", lokiOrgID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Loki (projet) injoignable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Loki (projet) a retourné %d: %s", resp.StatusCode, string(body))
	}

	var lr models.LokiResponse
	if err := json.Unmarshal(body, &lr); err != nil {
		return nil, err
	}

	var entries []models.LogEntry
	for _, stream := range lr.Data.Result {
		for _, v := range stream.Values {
			entries = append(entries, models.LogEntry{
				Timestamp: v[0],
				Line:      v[1],
				Labels:    stream.Stream,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp > entries[j].Timestamp
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// GetLogServices retourne la liste des valeurs possibles du label "pod"
// connues de Loki (cluster client), utilisée pour filtrer les logs par service.
func (s *ObservabilityService) GetLogServices(ctx context.Context) ([]string, error) {
	proxy, err := s.openSession(ctx, k8s.LokiTarget)
	if err != nil {
		return nil, err
	}
	defer proxy.Stop()

	endpoint := fmt.Sprintf("%s/loki/api/v1/label/pod/values", proxy.BaseURL())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Scope-OrgID", lokiOrgID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Loki (projet) injoignable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Loki (projet) a retourné %d: %s", resp.StatusCode, string(body))
	}

	var lr struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(body, &lr); err != nil {
		return nil, err
	}

	sort.Strings(lr.Data)
	return lr.Data, nil
}

// SearchTraces interroge Tempo (cluster client) et retourne les traces
// correspondant au service et à la durée minimale fournis.
func (s *ObservabilityService) SearchTraces(ctx context.Context, service string, minDurationMs, limit int) ([]models.TraceSummary, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	proxy, err := s.openSession(ctx, k8s.TempoTarget)
	if err != nil {
		return nil, err
	}
	defer proxy.Stop()

	endpoint := fmt.Sprintf("%s/api/search", proxy.BaseURL())
	params := url.Values{"limit": {strconv.Itoa(limit)}}
	if service != "" {
		params.Set("tags", fmt.Sprintf("service.name=%s", service))
	}
	if minDurationMs > 0 {
		params.Set("minDuration", fmt.Sprintf("%dms", minDurationMs))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Tempo (projet) injoignable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tempo (projet) a retourné %d: %s", resp.StatusCode, string(body))
	}

	var tr models.TempoSearchResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}

	result := make([]models.TraceSummary, 0, len(tr.Traces))
	for _, t := range tr.Traces {
		result = append(result, models.TraceSummary{
			TraceID:           t.TraceID,
			RootServiceName:   t.RootServiceName,
			RootTraceName:     t.RootTraceName,
			StartTimeUnixNano: t.StartTimeUnixNano,
			DurationMs:        t.DurationMs,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTimeUnixNano > result[j].StartTimeUnixNano
	})

	return result, nil
}

// GetTrace récupère le détail complet d'une trace (format OTLP/JSON brut de Tempo, cluster client).
func (s *ObservabilityService) GetTrace(ctx context.Context, traceID string) (json.RawMessage, error) {
	proxy, err := s.openSession(ctx, k8s.TempoTarget)
	if err != nil {
		return nil, err
	}
	defer proxy.Stop()

	endpoint := fmt.Sprintf("%s/api/traces/%s", proxy.BaseURL(), traceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Tempo (projet) injoignable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("trace introuvable")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tempo (projet) a retourné %d: %s", resp.StatusCode, string(body))
	}

	return json.RawMessage(body), nil
}

// GrafanaProxyHandler retourne un reverse proxy HTTP vers le pod Grafana du
// cluster client, en réutilisant un port-forward existant s'il est encore
// fonctionnel, ou en en ouvrant un nouveau sinon. Le port-forward est gardé
// ouvert entre les requêtes (un chargement de dashboard Grafana déclenche de
// nombreuses requêtes d'assets).
func (s *ObservabilityService) GrafanaProxyHandler(ctx context.Context) (*httputil.ReverseProxy, error) {
	s.grafanaMu.Lock()
	defer s.grafanaMu.Unlock()

	if s.grafanaProxy != nil {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.grafanaProxy.BaseURL()+"/api/health", nil)
		if err == nil {
			if resp, err := s.httpClient.Do(req); err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return s.newGrafanaReverseProxy(s.grafanaProxy.BaseURL()), nil
				}
			}
		}
		s.grafanaProxy.Stop()
		s.grafanaProxy = nil
	}

	proxy, err := s.openSession(ctx, k8s.GrafanaTarget)
	if err != nil {
		return nil, err
	}

	s.grafanaProxy = proxy
	return s.newGrafanaReverseProxy(proxy.BaseURL()), nil
}

func (s *ObservabilityService) newGrafanaReverseProxy(baseURL string) *httputil.ReverseProxy {
	target, _ := url.Parse(baseURL)
	return httputil.NewSingleHostReverseProxy(target)
}

// GetOverview retourne des KPIs globaux pour le cluster client, dérivés des
// métriques par job exposées par Prometheus.
func (s *ObservabilityService) GetOverview(ctx context.Context) (map[string]float64, error) {
	metrics, err := s.GetServiceMetrics(ctx)
	if err != nil {
		return nil, err
	}

	overview := map[string]float64{
		"total_services":   float64(len(metrics)),
		"total_goroutines": 0,
		"total_memory_mb":  0,
	}
	for _, m := range metrics {
		overview["total_goroutines"] += m["goroutines"]
		overview["total_memory_mb"] += m["memory_mb"]
	}

	return overview, nil
}
