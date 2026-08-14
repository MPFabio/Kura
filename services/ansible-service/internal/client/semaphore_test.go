package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeSemaphore monte un faux serveur Semaphore avec un projet, un template,
// un inventaire et deux tâches.
func fakeSemaphore(t *testing.T) *Semaphore {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/project/1", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "kura-ops"})
	})
	mux.HandleFunc("/api/project/1/templates", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 10, "alias": "deploy-web", "playbook": "site.yml", "inventory_id": 5, "repository_id": 3},
		})
	})
	mux.HandleFunc("/api/project/1/templates/10", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 10, "alias": "deploy-web", "playbook": "site.yml", "inventory_id": 5, "repository_id": 3,
		})
	})
	mux.HandleFunc("/api/project/1/tasks/42", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 42, "template_id": 10, "status": "success",
			"created": "2026-06-01T10:00:00Z", "end": "2026-06-01T10:02:30Z",
		})
	})
	mux.HandleFunc("/api/project/1/inventory", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 5, "name": "prod", "type": "static", "inventory": "[web]\nsrv1\nsrv2\n# commentaire\n"},
		})
	})
	mux.HandleFunc("/api/project/1/inventory/5", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 5, "name": "prod", "type": "static", "inventory": "[web]\nsrv1\nsrv2\n# commentaire\n",
		})
	})
	mux.HandleFunc("/api/project/1/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var payload map[string]any
			_ = json.NewDecoder(r.Body).Decode(&payload)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 99, "template_id": payload["template_id"], "status": "waiting",
			})
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 42, "template_id": 10, "status": "success",
				"created": "2026-06-01T10:00:00Z", "end": "2026-06-01T10:02:30Z"},
			{"id": 41, "template_id": 10, "status": "error"},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return NewSemaphore(srv.URL, "jeton", 1)
}

func TestGetJobsMapsToAWXFormat(t *testing.T) {
	s := fakeSemaphore(t)
	jobs, err := s.GetJobs(context.Background(), 20)
	if err != nil {
		t.Fatalf("GetJobs: %v", err)
	}
	if jobs.Count != 2 {
		t.Fatalf("count = %d, attendu 2", jobs.Count)
	}

	first := jobs.Results[0]
	if first["status"] != "successful" {
		t.Errorf("statut success doit devenir successful, obtenu %v", first["status"])
	}
	if first["name"] != "deploy-web" {
		t.Errorf("le nom doit venir de l'alias du template, obtenu %v", first["name"])
	}
	if first["project_name"] != "kura-ops" {
		t.Errorf("project_name = %v", first["project_name"])
	}
	if elapsed := first["elapsed"].(float64); elapsed != 150 {
		t.Errorf("elapsed = %v, attendu 150 (2min30)", elapsed)
	}

	second := jobs.Results[1]
	if second["status"] != "failed" {
		t.Errorf("statut error doit devenir failed, obtenu %v", second["status"])
	}
}

func TestGetInventoryHostsParsesStaticInventory(t *testing.T) {
	s := fakeSemaphore(t)
	hosts, err := s.GetInventoryHosts(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetInventoryHosts: %v", err)
	}
	if hosts.Count != 2 {
		t.Fatalf("count = %d, attendu 2 (sections et commentaires exclus)", hosts.Count)
	}
	if hosts.Results[0]["name"] != "srv1" || hosts.Results[1]["name"] != "srv2" {
		t.Errorf("hôtes inattendus: %v", hosts.Results)
	}
}

func TestLaunchJobTemplateSerializesExtraVars(t *testing.T) {
	s := fakeSemaphore(t)
	task, err := s.LaunchJobTemplate(context.Background(), 10, map[string]any{"cluster_id": 7})
	if err != nil {
		t.Fatalf("LaunchJobTemplate: %v", err)
	}
	if task == nil || intVal(task["id"]) != 99 {
		t.Fatalf("tâche inattendue: %v", task)
	}
}

func TestGetJobSingleFetchEnrichment(t *testing.T) {
	s := fakeSemaphore(t)
	job, err := s.GetJob(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if job["name"] != "deploy-web" {
		t.Errorf("le nom doit venir du template référencé, obtenu %v", job["name"])
	}
	if job["inventory_name"] != "prod" {
		t.Errorf("inventory_name = %v, attendu prod", job["inventory_name"])
	}
	if job["status"] != "successful" {
		t.Errorf("status = %v, attendu successful", job["status"])
	}
}

func TestUnconfiguredClientReturnsNil(t *testing.T) {
	s := NewSemaphore("", "", 1)
	jobs, err := s.GetJobs(context.Background(), 20)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if jobs != nil {
		t.Error("un client non configuré doit retourner nil (traduit en liste vide par le handler)")
	}
}

func TestComputeElapsed(t *testing.T) {
	if got := computeElapsed("2026-06-01T10:00:00Z", "2026-06-01T10:01:00Z"); got != 60 {
		t.Errorf("elapsed = %v, attendu 60", got)
	}
	if got := computeElapsed("", "2026-06-01T10:01:00Z"); got != 0 {
		t.Errorf("started vide doit donner 0, obtenu %v", got)
	}
	if got := computeElapsed("2026-06-01T10:01:00Z", "2026-06-01T10:00:00Z"); got != 0 {
		t.Errorf("une durée négative doit être ramenée à 0, obtenu %v", got)
	}
}
