package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modulops/vault-service/internal/config"
)

// newTestService construit un VaultService pointant vers un faux OpenBao HTTP.
// Le configstore est injoignable : les valeurs proviennent du fallback env.
func newTestService(t *testing.T, handler http.Handler) (*VaultService, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	svc, err := New(&config.Config{
		AuthServiceURL: "http://127.0.0.1:1",
		OpenBaoAddr:    srv.URL,
		OpenBaoToken:   "token-de-test",
		MountPath:      "secret",
	}, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc, srv
}

func TestKvPaths(t *testing.T) {
	svc, _ := newTestService(t, http.NotFoundHandler())

	cases := []struct{ in, data, meta string }{
		{"app/db", "secret/data/app/db", "secret/metadata/app/db"},
		{"/app/db", "secret/data/app/db", "secret/metadata/app/db"},
		{"", "secret/data/", "secret/metadata"},
	}
	for _, c := range cases {
		if got := svc.kvDataPath(c.in); got != c.data {
			t.Errorf("kvDataPath(%q) = %q, attendu %q", c.in, got, c.data)
		}
		if got := svc.kvMetaPath(c.in); got != c.meta {
			t.Errorf("kvMetaPath(%q) = %q, attendu %q", c.in, got, c.meta)
		}
	}
}

func TestIntFromMeta(t *testing.T) {
	if got := intFromMeta(map[string]interface{}{"version": float64(3)}, "version"); got != 3 {
		t.Errorf("float64: %d", got)
	}
	if got := intFromMeta(map[string]interface{}{"version": 7}, "version"); got != 7 {
		t.Errorf("int: %d", got)
	}
	if got := intFromMeta(nil, "version"); got != 0 {
		t.Errorf("nil meta: %d", got)
	}
	if got := intFromMeta(map[string]interface{}{"version": "abc"}, "version"); got != 0 {
		t.Errorf("type inattendu: %d", got)
	}
}

func TestStatus(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/health" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"initialized":  true,
			"sealed":       false,
			"version":      "2.5.0",
			"cluster_name": "kura-openbao",
		})
	}))

	st, err := svc.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !st.Initialized || st.Sealed || st.Version != "2.5.0" || st.ClusterName != "kura-openbao" {
		t.Errorf("statut inattendu: %+v", st)
	}
}

func TestGetSecretReadsKVv2(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/data/app/db" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("X-Vault-Token"); got != "token-de-test" {
			t.Errorf("token manquant ou inattendu: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"data":     map[string]any{"password": "hunter2"},
				"metadata": map[string]any{"version": float64(4)},
			},
		})
	}))

	sec, err := svc.GetSecret(context.Background(), "app/db")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if sec.Data["password"] != "hunter2" {
		t.Errorf("données inattendues: %+v", sec.Data)
	}
	if sec.Metadata.Version != 4 {
		t.Errorf("version = %d, attendu 4", sec.Metadata.Version)
	}
}

func TestGetSecretNotFound(t *testing.T) {
	svc, _ := newTestService(t, http.NotFoundHandler())
	if _, err := svc.GetSecret(context.Background(), "inexistant"); err == nil {
		t.Fatal("un secret absent doit produire une erreur")
	}
}

func TestListSecrets(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/metadata/app" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"keys": []any{"db", "api/"},
			},
		})
	}))

	keys, err := svc.ListSecrets(context.Background(), "app")
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(keys) != 2 || keys[0] != "db" || keys[1] != "api/" {
		t.Errorf("clés inattendues: %v", keys)
	}
}

func TestListSecretsEmpty(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[]}`))
	}))
	keys, err := svc.ListSecrets(context.Background(), "vide")
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("un path vide doit retourner une liste vide, obtenu %v", keys)
	}
}

func TestWriteSecretWrapsPayload(t *testing.T) {
	var body map[string]any
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/data/app/db" || r.Method != http.MethodPut {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := svc.WriteSecret(context.Background(), "app/db", map[string]interface{}{"user": "kura"})
	if err != nil {
		t.Fatalf("WriteSecret: %v", err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok || data["user"] != "kura" {
		t.Errorf("le payload doit être enveloppé sous \"data\" (KV v2), obtenu %+v", body)
	}
}

func TestGetConfigMasksToken(t *testing.T) {
	svc, _ := newTestService(t, http.NotFoundHandler())
	cfg, err := svc.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg["openbao_token"] != "***" {
		t.Errorf("le token doit être masqué, obtenu %q", cfg["openbao_token"])
	}
	if cfg["linked"] != "true" {
		t.Errorf("linked = %q, attendu true (token configuré)", cfg["linked"])
	}
	if cfg["openbao_mount_path"] != "secret" {
		t.Errorf("mount path = %q", cfg["openbao_mount_path"])
	}
}
