// Package service porte la logique métier du service Ansible : orchestration
// du client Semaphore, cache Valkey et configuration persistée (configstore).
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/modulops/ansible-service/internal/client"
	"github.com/modulops/ansible-service/internal/config"
	"github.com/modulops/ansible-service/internal/configstore"
	"github.com/modulops/ansible-service/internal/parser"
)

type AnsibleService struct {
	semaphore  *client.Semaphore
	rdb        *redis.Client
	cfg        *config.Config
	cfgStore   *configstore.Client
	httpClient *http.Client
}

func New(cfg *config.Config, rdb *redis.Client) *AnsibleService {
	cs := configstore.New(cfg.AuthServiceURL, "ansible")
	ctx := context.Background()

	// La configuration persistée (Postgres via configstore) prime sur l'environnement.
	semaphoreURL := cs.GetOrFallback(ctx, "semaphore_url", cfg.SemaphoreURL)
	semaphoreToken := cs.GetOrFallback(ctx, "semaphore_token", cfg.SemaphoreAPIToken)
	projectID := cfg.SemaphoreProjectID
	if stored := cs.GetOrFallback(ctx, "semaphore_project_id", ""); stored != "" {
		if id, err := strconv.Atoi(stored); err == nil {
			projectID = id
		}
	}

	return &AnsibleService{
		semaphore:  client.NewSemaphore(semaphoreURL, semaphoreToken, projectID),
		rdb:        rdb,
		cfg:        cfg,
		cfgStore:   cs,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ── Configuration Semaphore ──────────────────────────────────────────────────

// GetConfig retourne la configuration Semaphore active (token jamais renvoyé).
func (s *AnsibleService) GetConfig(ctx context.Context) map[string]any {
	semaphoreURL := s.cfgStore.GetOrFallback(ctx, "semaphore_url", s.cfg.SemaphoreURL)
	token := s.cfgStore.GetOrFallback(ctx, "semaphore_token", s.cfg.SemaphoreAPIToken)
	projectID := s.cfg.SemaphoreProjectID
	if stored := s.cfgStore.GetOrFallback(ctx, "semaphore_project_id", ""); stored != "" {
		if id, err := strconv.Atoi(stored); err == nil {
			projectID = id
		}
	}
	hasToken := token != ""
	return map[string]any{
		"semaphore_url":        semaphoreURL,
		"semaphore_project_id": projectID,
		"has_token":            hasToken,
		"configured":           semaphoreURL != "" && hasToken,
	}
}

// SetConfig persiste la configuration Semaphore et réinitialise le client à chaud.
func (s *AnsibleService) SetConfig(ctx context.Context, semaphoreURL, token string, projectID int) (map[string]any, error) {
	kv := map[string]string{"semaphore_project_id": strconv.Itoa(projectID)}
	if semaphoreURL != "" {
		kv["semaphore_url"] = semaphoreURL
	}
	if token != "" {
		kv["semaphore_token"] = token
	}
	if err := s.cfgStore.SetMany(ctx, kv); err != nil {
		return nil, err
	}

	if semaphoreURL == "" {
		semaphoreURL = s.cfgStore.GetOrFallback(ctx, "semaphore_url", s.cfg.SemaphoreURL)
	}
	if token == "" {
		token = s.cfgStore.GetOrFallback(ctx, "semaphore_token", s.cfg.SemaphoreAPIToken)
	}
	s.semaphore = client.NewSemaphore(semaphoreURL, token, projectID)
	log.Printf("Client Semaphore réinitialisé → %s, projet %d", semaphoreURL, projectID)

	s.invalidate(ctx, "ansible:")
	return s.GetConfig(ctx), nil
}

// ── Jobs ─────────────────────────────────────────────────────────────────────

func (s *AnsibleService) GetJobs(ctx context.Context, pageSize int) (*client.Paginated, error) {
	key := fmt.Sprintf("ansible:jobs:%d", pageSize)
	if cached := s.getCachedPaginated(ctx, key); cached != nil {
		return cached, nil
	}
	result, err := s.semaphore.GetJobs(ctx, pageSize)
	if err != nil || result == nil {
		return result, err
	}
	s.setCached(ctx, key, result, s.cfg.CacheTTL)
	return result, nil
}

func (s *AnsibleService) GetJob(ctx context.Context, jobID int, includeStdout bool) (map[string]any, error) {
	key := fmt.Sprintf("ansible:job:%d:%t", jobID, includeStdout)
	if cached := s.getCachedMap(ctx, key); cached != nil {
		return cached, nil
	}
	job, err := s.semaphore.GetJob(ctx, jobID)
	if err != nil || job == nil {
		return job, err
	}
	if includeStdout {
		stdout, err := s.semaphore.GetJobStdout(ctx, jobID)
		if err == nil {
			job["stdout"] = stdout
		}
	}
	s.setCached(ctx, key, job, 60*time.Second)
	return job, nil
}

// GetJobHistory retourne l'historique des jobs (même source que la liste).
func (s *AnsibleService) GetJobHistory(ctx context.Context, pageSize int) (*client.Paginated, error) {
	return s.GetJobs(ctx, pageSize)
}

// ── Inventaires ──────────────────────────────────────────────────────────────

func (s *AnsibleService) GetInventories(ctx context.Context) (*client.Paginated, error) {
	key := "ansible:inventories"
	if cached := s.getCachedPaginated(ctx, key); cached != nil {
		return cached, nil
	}
	result, err := s.semaphore.GetInventories(ctx)
	if err != nil || result == nil {
		return result, err
	}
	s.setCached(ctx, key, result, s.cfg.CacheTTL)
	return result, nil
}

func (s *AnsibleService) GetInventory(ctx context.Context, inventoryID int) (map[string]any, error) {
	key := fmt.Sprintf("ansible:inventory:%d", inventoryID)
	if cached := s.getCachedMap(ctx, key); cached != nil {
		return cached, nil
	}
	inventory, err := s.semaphore.GetInventory(ctx, inventoryID)
	if err != nil || inventory == nil {
		return inventory, err
	}
	s.setCached(ctx, key, inventory, 300*time.Second)
	return inventory, nil
}

func (s *AnsibleService) GetInventoryHosts(ctx context.Context, inventoryID int) (*client.Paginated, error) {
	key := fmt.Sprintf("ansible:inventory_hosts:%d", inventoryID)
	if cached := s.getCachedPaginated(ctx, key); cached != nil {
		return cached, nil
	}
	result, err := s.semaphore.GetInventoryHosts(ctx, inventoryID)
	if err != nil || result == nil {
		return result, err
	}
	s.setCached(ctx, key, result, 180*time.Second)
	return result, nil
}

// ── Templates ────────────────────────────────────────────────────────────────

func (s *AnsibleService) GetJobTemplates(ctx context.Context) (*client.Paginated, error) {
	key := "ansible:job_templates"
	if cached := s.getCachedPaginated(ctx, key); cached != nil {
		return cached, nil
	}
	result, err := s.semaphore.GetJobTemplates(ctx)
	if err != nil || result == nil {
		return result, err
	}
	s.setCached(ctx, key, result, s.cfg.CacheTTL)
	return result, nil
}

func (s *AnsibleService) GetJobTemplate(ctx context.Context, templateID int) (map[string]any, error) {
	key := fmt.Sprintf("ansible:job_template:%d", templateID)
	if cached := s.getCachedMap(ctx, key); cached != nil {
		return cached, nil
	}
	template, err := s.semaphore.GetJobTemplate(ctx, templateID)
	if err != nil || template == nil {
		return template, err
	}
	s.setCached(ctx, key, template, 300*time.Second)
	return template, nil
}

// LaunchJobTemplate lance un job depuis un template, en injectant les
// informations d'accès au cluster Kubernetes actif si disponibles.
func (s *AnsibleService) LaunchJobTemplate(ctx context.Context, templateID int, extraVars map[string]any) (map[string]any, error) {
	if extraVars == nil {
		extraVars = map[string]any{}
	}

	if s.cfg.InternalAPISecret != "" {
		if cluster := s.fetchActiveCluster(ctx); cluster != nil {
			if clusterID, found := cluster["id"]; found && clusterID != nil {
				extraVars["cluster_id"] = clusterID
				extraVars["k8s_service_url"] = s.cfg.K8sServiceURL
				extraVars["internal_api_token"] = s.cfg.InternalAPISecret
			}
		}
	}

	task, err := s.semaphore.LaunchJobTemplate(ctx, templateID, extraVars)
	if err != nil || task == nil {
		return nil, err
	}

	s.invalidate(ctx, "ansible:jobs:")

	if jobID := intFromAny(task["id"]); jobID != 0 {
		return s.GetJob(ctx, jobID, false)
	}
	return nil, nil
}

func (s *AnsibleService) fetchActiveCluster(ctx context.Context) map[string]any {
	url := s.cfg.K8sServiceURL + "/api/v1/k8s/clusters/active"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Impossible de récupérer le cluster actif: %v", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var cluster map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&cluster); err != nil {
		return nil
	}
	return cluster
}

// ── Projets et playbooks ─────────────────────────────────────────────────────

func (s *AnsibleService) GetProjects(ctx context.Context) (*client.Paginated, error) {
	key := "ansible:projects"
	if cached := s.getCachedPaginated(ctx, key); cached != nil {
		return cached, nil
	}
	result, err := s.semaphore.GetProjects(ctx)
	if err != nil || result == nil {
		return result, err
	}
	s.setCached(ctx, key, result, s.cfg.CacheTTL)
	return result, nil
}

func (s *AnsibleService) GetProject(ctx context.Context, projectID int) (map[string]any, error) {
	return s.semaphore.GetProject(ctx, projectID)
}

// AnalyzePlaybook analyse en profondeur un playbook YAML.
func (s *AnsibleService) AnalyzePlaybook(content string) map[string]any {
	return parser.AnalyzePlaybook(content)
}

// GetTemplatePlaybookSource récupère le contenu YAML du playbook associé à un
// template, via le dépôt Git du projet Semaphore et le code-service.
func (s *AnsibleService) GetTemplatePlaybookSource(ctx context.Context, templateID int) (map[string]any, error) {
	template, err := s.semaphore.GetJobTemplate(ctx, templateID)
	if err != nil || template == nil {
		return nil, err
	}

	playbookPath, _ := template["playbook"].(string)
	repositoryID := intFromAny(template["repository_id"])
	if playbookPath == "" || repositoryID == 0 {
		return nil, nil
	}

	repository, err := s.semaphore.GetRepository(ctx, repositoryID)
	if err != nil || repository == nil {
		return nil, err
	}

	gitURL, _ := repository["git_url"].(string)
	repoFullName := ParseGitRepo(gitURL)
	if repoFullName == "" {
		return nil, nil
	}
	ref, _ := repository["git_branch"].(string)

	fileURL := fmt.Sprintf("%s/api/v1/code/file?repo=%s&path=%s&ref=%s",
		s.cfg.CodeServiceURL, repoFullName, playbookPath, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Erreur récupération playbook %s depuis %s: %v", playbookPath, repoFullName, err)
		return nil, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("code-service a retourné %d pour %s", resp.StatusCode, playbookPath)
		return nil, nil
	}
	var fileData map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&fileData); err != nil {
		return nil, nil
	}
	content, _ := fileData["content"].(string)

	return map[string]any{
		"path":    playbookPath,
		"repo":    repoFullName,
		"ref":     ref,
		"content": content,
	}, nil
}

// ── Stubs de compatibilité AWX ───────────────────────────────────────────────

func (s *AnsibleService) GetCredentials(ctx context.Context) *client.Paginated {
	return s.semaphore.GetCredentials(ctx)
}

func (s *AnsibleService) GetOrganizations(ctx context.Context) *client.Paginated {
	return s.semaphore.GetOrganizations(ctx)
}

// ── Cache ────────────────────────────────────────────────────────────────────

func (s *AnsibleService) getCachedPaginated(ctx context.Context, key string) *client.Paginated {
	if s.rdb == nil {
		return nil
	}
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var result client.Paginated
	if json.Unmarshal(raw, &result) != nil {
		return nil
	}
	return &result
}

func (s *AnsibleService) getCachedMap(ctx context.Context, key string) map[string]any {
	if s.rdb == nil {
		return nil
	}
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var result map[string]any
	if json.Unmarshal(raw, &result) != nil {
		return nil
	}
	return result
}

func (s *AnsibleService) setCached(ctx context.Context, key string, value any, ttl time.Duration) {
	if s.rdb == nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	s.rdb.Set(ctx, key, raw, ttl)
}

func (s *AnsibleService) invalidate(ctx context.Context, prefix string) {
	if s.rdb == nil {
		return
	}
	iter := s.rdb.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		s.rdb.Del(ctx, iter.Val())
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// ParseGitRepo extrait "owner/repo" d'une URL git http(s) ou SSH.
func ParseGitRepo(gitURL string) string {
	url := strings.TrimSpace(gitURL)
	if url == "" {
		return ""
	}
	url = strings.TrimSuffix(url, ".git")

	var path string
	if strings.HasPrefix(url, "git@") {
		// git@host:owner/repo
		_, path, _ = strings.Cut(url, ":")
	} else {
		// https://host/owner/repo
		if _, rest, found := strings.Cut(url, "://"); found {
			url = rest
		}
		if _, rest, found := strings.Cut(url, "/"); found {
			path = rest
		}
	}

	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	}
	return 0
}
