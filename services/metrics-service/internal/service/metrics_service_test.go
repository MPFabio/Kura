package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/modulops/metrics-service/internal/config"
	"github.com/redis/go-redis/v9"
)

// newTestService construit un MetricsService dont le configstore est injoignable
// (fallback env systématique) et dont le cache Valkey pointe vers une adresse
// morte : chaque appel passe par le chemin "cache miss".
func newTestService(cfg *config.Config) *MetricsService {
	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		DialTimeout: 50 * time.Millisecond,
	})
	if cfg.AuthServiceURL == "" {
		cfg.AuthServiceURL = "http://127.0.0.1:1"
	}
	return New(cfg, rdb)
}

func TestQueryPrometheusParsesJobValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			t.Errorf("chemin inattendu: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("query"); got != "go_goroutines" {
			t.Errorf("query inattendue: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{"metric": {"job": "auth-service"}, "value": [1700000000, "42"]},
					{"metric": {"job": "k8s-service"}, "value": [1700000000, "17.5"]},
					{"metric": {"job": "casse"}, "value": [1700000000]},
					{"metric": {"job": "pas-un-nombre"}, "value": [1700000000, "abc"]}
				]
			}
		}`))
	}))
	defer srv.Close()

	s := newTestService(&config.Config{VictoriaMetricsURL: srv.URL})
	got, err := s.queryPrometheus(context.Background(), "go_goroutines")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if got["auth-service"] != 42 {
		t.Errorf("auth-service = %v, attendu 42", got["auth-service"])
	}
	if got["k8s-service"] != 17.5 {
		t.Errorf("k8s-service = %v, attendu 17.5", got["k8s-service"])
	}
	if _, ok := got["casse"]; ok {
		t.Error("une valeur incomplète ne doit pas être retenue")
	}
	if _, ok := got["pas-un-nombre"]; ok {
		t.Error("une valeur non numérique ne doit pas être retenue")
	}
}

func TestQueryPrometheusInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pas du json"))
	}))
	defer srv.Close()

	s := newTestService(&config.Config{VictoriaMetricsURL: srv.URL})
	if _, err := s.queryPrometheus(context.Background(), "up"); err == nil {
		t.Fatal("une réponse non-JSON doit produire une erreur")
	}
}

func TestCheckHealth(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer down.Close()

	s := newTestService(&config.Config{})
	if !s.checkHealth(context.Background(), up.URL) {
		t.Error("un 200 doit être considéré comme sain")
	}
	if s.checkHealth(context.Background(), down.URL) {
		t.Error("un 503 ne doit pas être considéré comme sain")
	}
	if s.checkHealth(context.Background(), "http://127.0.0.1:1/health") {
		t.Error("une cible injoignable ne doit pas être considérée comme saine")
	}
}

func TestGetLogsQueryConstructionAndOrder(t *testing.T) {
	var seen url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/loki/api/v1/query_range" {
			t.Errorf("chemin inattendu: %s", r.URL.Path)
		}
		seen = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"resultType": "streams",
				"result": []map[string]any{
					{
						"stream": map[string]string{"service": "auth-service"},
						"values": [][]string{
							{"1700000001000000000", "ancien"},
							{"1700000009000000000", "recent"},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	s := newTestService(&config.Config{LokiURL: srv.URL})
	entries, err := s.GetLogs(context.Background(), "auth-service", "erreur", 0)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}

	if got := seen.Get("query"); got != `{service="auth-service"} |= "erreur"` {
		t.Errorf("requête LogQL inattendue: %s", got)
	}
	if got := seen.Get("limit"); got != "200" {
		t.Errorf("un limit hors bornes doit être ramené à 200, obtenu %s", got)
	}
	if len(entries) != 2 || entries[0].Line != "recent" {
		t.Errorf("les entrées doivent être triées de la plus récente à la plus ancienne: %+v", entries)
	}
}

func TestGetLogsLokiUnavailable(t *testing.T) {
	s := newTestService(&config.Config{LokiURL: "http://127.0.0.1:1"})
	if _, err := s.GetLogs(context.Background(), "", "", 10); err == nil {
		t.Fatal("un Loki injoignable doit produire une erreur")
	}
}

func TestGetLogServices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/loki/api/v1/label/service/values" {
			t.Errorf("chemin inattendu: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"success","data":["auth-service","kong"]}`))
	}))
	defer srv.Close()

	s := newTestService(&config.Config{LokiURL: srv.URL})
	got, err := s.GetLogServices(context.Background())
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(got) != 2 || got[0] != "auth-service" || got[1] != "kong" {
		t.Errorf("liste de services inattendue: %v", got)
	}
}

func TestGetConfigFallsBackToEnvValues(t *testing.T) {
	s := newTestService(&config.Config{
		VictoriaMetricsURL: "http://prom.fallback:9090",
		GrafanaURL:         "http://grafana.fallback:3000",
	})
	got, err := s.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if got["victoriametrics_url"] != "http://prom.fallback:9090" {
		t.Errorf("victoriametrics_url = %s", got["victoriametrics_url"])
	}
	if got["grafana_url"] != "http://grafana.fallback:3000" {
		t.Errorf("grafana_url = %s", got["grafana_url"])
	}
}

func TestInternalObservabilityEnabled(t *testing.T) {
	on := newTestService(&config.Config{InternalObservabilityEnabled: true})
	off := newTestService(&config.Config{InternalObservabilityEnabled: false})
	if !on.InternalObservabilityEnabled() || off.InternalObservabilityEnabled() {
		t.Error("InternalObservabilityEnabled doit refléter la configuration")
	}
}
