package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/modulops/metrics-service/internal/config"
	"github.com/modulops/metrics-service/internal/configstore"
	"github.com/modulops/metrics-service/internal/models"
	"github.com/redis/go-redis/v9"
)

// knownServices liste tous les microservices Kura avec leur URL de health check interne.
var knownServices = []struct {
	name      string
	job       string
	healthURL string
}{
	{"Auth", "auth-service", "http://auth-service:8080/health"},
	{"Kubernetes", "k8s-service", "http://k8s-service:8081/health"},
	{"Terraform", "terraform-service", "http://terraform-service:8082/health"},
	{"Ansible", "ansible-service", "http://ansible-service:8083/health"},
	{"Pipeline", "pipeline-service", "http://pipeline-service:8084/health"},
	{"Metrics", "metrics-service", "http://localhost:8086/health"},
}

type MetricsService struct {
	cfg        *config.Config
	cfgStore   *configstore.Client
	redis      *redis.Client
	httpClient *http.Client
}

func New(cfg *config.Config, rdb *redis.Client) *MetricsService {
	return &MetricsService{
		cfg:        cfg,
		cfgStore:   configstore.New(cfg.AuthServiceURL, "metrics"),
		redis:      rdb,
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

// getPrometheusURL retourne l'URL Prometheus (configstore prioritaire sur env).
func (s *MetricsService) getPrometheusURL(ctx context.Context) string {
	return s.cfgStore.GetOrFallback(ctx, "prometheus_url", s.cfg.PrometheusURL)
}

// getGrafanaURL retourne l'URL Grafana (configstore prioritaire sur env).
func (s *MetricsService) getGrafanaURL(ctx context.Context) string {
	return s.cfgStore.GetOrFallback(ctx, "grafana_url", s.cfg.GrafanaURL)
}

// getLokiURL retourne l'URL Loki (configstore prioritaire sur env).
func (s *MetricsService) getLokiURL(ctx context.Context) string {
	return s.cfgStore.GetOrFallback(ctx, "loki_url", s.cfg.LokiURL)
}

// getTempoURL retourne l'URL Tempo (configstore prioritaire sur env).
func (s *MetricsService) getTempoURL(ctx context.Context) string {
	return s.cfgStore.GetOrFallback(ctx, "tempo_url", s.cfg.TempoURL)
}

// SetConfig met à jour les URLs Prometheus et Grafana.
func (s *MetricsService) SetConfig(ctx context.Context, prometheusURL, grafanaURL string) error {
	kv := map[string]string{}
	if prometheusURL != "" {
		kv["prometheus_url"] = prometheusURL
	}
	if grafanaURL != "" {
		kv["grafana_url"] = grafanaURL
	}
	if len(kv) == 0 {
		return nil
	}
	return s.cfgStore.SetMany(ctx, kv)
}

// GetConfig retourne la config actuelle (URLs masquées si sensibles).
func (s *MetricsService) GetConfig(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		"prometheus_url": s.getPrometheusURL(ctx),
		"grafana_url":    s.getGrafanaURL(ctx),
	}, nil
}

// InternalObservabilityEnabled indique si l'observabilité interne de la
// plateforme Kura (santé/metrics de ses propres microservices) est exposée.
// Désactivée en mode SaaS, activable en self-hosted.
func (s *MetricsService) InternalObservabilityEnabled() bool {
	return s.cfg.InternalObservabilityEnabled
}

// GetHealth retourne l'état up/down de chaque service via health check direct + métriques Prometheus.
func (s *MetricsService) GetHealth(ctx context.Context) ([]models.ServiceHealth, error) {
	const cacheKey = "metrics:health"

	if cached, err := s.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var result []models.ServiceHealth
		if json.Unmarshal(cached, &result) == nil {
			return result, nil
		}
	}

	goroutinesMap, _ := s.queryPrometheus(ctx, "go_goroutines")
	memMap, _ := s.queryPrometheus(ctx, "process_resident_memory_bytes")

	var result []models.ServiceHealth
	for _, svc := range knownServices {
		up := s.checkHealth(ctx, svc.healthURL)
		result = append(result, models.ServiceHealth{
			Name:       svc.name,
			Job:        svc.job,
			Up:         up,
			Goroutines: goroutinesMap[svc.job],
			MemoryMB:   memMap[svc.job] / 1024 / 1024,
		})
	}

	s.cache(ctx, cacheKey, result)
	return result, nil
}

// GetServices retourne les métriques détaillées par service.
func (s *MetricsService) GetServices(ctx context.Context) ([]models.ServiceMetric, error) {
	const cacheKey = "metrics:services"

	if cached, err := s.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var result []models.ServiceMetric
		if json.Unmarshal(cached, &result) == nil {
			return result, nil
		}
	}

	goroutinesMap, _ := s.queryPrometheus(ctx, "go_goroutines")
	cpuMap, _ := s.queryPrometheus(ctx, "rate(process_cpu_seconds_total[5m])")
	memMap, _ := s.queryPrometheus(ctx, "process_resident_memory_bytes")

	var result []models.ServiceMetric
	for _, svc := range knownServices {
		up := s.checkHealth(ctx, svc.healthURL)
		result = append(result, models.ServiceMetric{
			Name:       svc.name,
			Job:        svc.job,
			Up:         up,
			Goroutines: goroutinesMap[svc.job],
			CPURate:    cpuMap[svc.job],
			MemoryMB:   memMap[svc.job] / 1024 / 1024,
		})
	}

	s.cache(ctx, cacheKey, result)
	return result, nil
}

// GetOverview retourne les KPIs globaux de la plateforme.
func (s *MetricsService) GetOverview(ctx context.Context) (*models.Overview, error) {
	const cacheKey = "metrics:overview"

	if cached, err := s.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var result models.Overview
		if json.Unmarshal(cached, &result) == nil {
			return &result, nil
		}
	}

	services, err := s.GetServices(ctx)
	if err != nil {
		return nil, err
	}

	ov := &models.Overview{TotalServices: len(services)}
	for _, svc := range services {
		if svc.Up {
			ov.ServicesUp++
		} else {
			ov.ServicesDown++
		}
		ov.TotalGoroutines += svc.Goroutines
		ov.TotalMemoryMB += svc.MemoryMB
	}

	s.cache(ctx, cacheKey, ov)
	return ov, nil
}

// checkHealth effectue un GET sur l'URL de health check et retourne true si HTTP 200.
func (s *MetricsService) checkHealth(ctx context.Context, healthURL string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return false
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// queryPrometheus exécute une instant query et retourne une map job → valeur float64.
func (s *MetricsService) queryPrometheus(ctx context.Context, query string) (map[string]float64, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query", s.getPrometheusURL(ctx))
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

// GetLogs interroge Loki et retourne les logs correspondant au service et à la recherche fournis.
// service: nom du service Docker Compose (label "service" injecté par Promtail), vide = tous.
// search: filtre texte appliqué via |= "..." (LogQL).
// limit: nombre maximum d'entrées retournées (les plus récentes en premier).
func (s *MetricsService) GetLogs(ctx context.Context, service, search string, limit int) ([]models.LogEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}

	selector := `{service=~".+"}`
	if service != "" {
		selector = fmt.Sprintf(`{service=%q}`, service)
	}
	query := selector
	if search != "" {
		query = fmt.Sprintf(`%s |= %q`, selector, search)
	}

	endpoint := fmt.Sprintf("%s/loki/api/v1/query_range", s.getLokiURL(ctx))
	params := url.Values{
		"query":     {query},
		"limit":     {strconv.Itoa(limit)},
		"direction": {"backward"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Loki injoignable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Loki a retourné %d: %s", resp.StatusCode, string(body))
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

// GetLogServices retourne la liste des valeurs possibles du label "service" connues de Loki.
func (s *MetricsService) GetLogServices(ctx context.Context) ([]string, error) {
	endpoint := fmt.Sprintf("%s/loki/api/v1/label/service/values", s.getLokiURL(ctx))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Loki injoignable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Loki a retourné %d: %s", resp.StatusCode, string(body))
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

// SearchTraces interroge Tempo et retourne les traces correspondant au service et à la durée minimale fournis.
// service: filtre sur l'attribut "service.name" (vide = tous).
// minDurationMs: durée minimale en millisecondes (0 = pas de filtre).
// limit: nombre maximum de traces retournées.
func (s *MetricsService) SearchTraces(ctx context.Context, service string, minDurationMs, limit int) ([]models.TraceSummary, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	endpoint := fmt.Sprintf("%s/api/search", s.getTempoURL(ctx))
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
		return nil, fmt.Errorf("Tempo injoignable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tempo a retourné %d: %s", resp.StatusCode, string(body))
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

// GetTrace récupère le détail complet d'une trace (format OTLP/JSON brut de Tempo).
func (s *MetricsService) GetTrace(ctx context.Context, traceID string) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("%s/api/traces/%s", s.getTempoURL(ctx), traceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Tempo injoignable: %w", err)
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
		return nil, fmt.Errorf("Tempo a retourné %d: %s", resp.StatusCode, string(body))
	}

	return json.RawMessage(body), nil
}

func (s *MetricsService) cache(ctx context.Context, key string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.redis.Set(ctx, key, b, s.cfg.CacheTTL)
}
