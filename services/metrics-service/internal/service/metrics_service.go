package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/modulops/metrics-service/internal/config"
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
	redis      *redis.Client
	httpClient *http.Client
}

func New(cfg *config.Config, rdb *redis.Client) *MetricsService {
	return &MetricsService{
		cfg:   cfg,
		redis: rdb,
		// Timeout court pour les health checks : on ne veut pas bloquer le frontend
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
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
	endpoint := fmt.Sprintf("%s/api/v1/query", s.cfg.PrometheusURL)
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

func (s *MetricsService) cache(ctx context.Context, key string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.redis.Set(ctx, key, b, s.cfg.CacheTTL)
}
