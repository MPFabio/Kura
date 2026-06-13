package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/modulops/pipeline-service/internal/adapter"
	"github.com/modulops/pipeline-service/internal/cache"
	"github.com/modulops/pipeline-service/internal/client"
	"github.com/modulops/pipeline-service/internal/config"
	"github.com/modulops/pipeline-service/internal/configstore"
	"github.com/modulops/pipeline-service/internal/models"
)

const (
	keyRunPrefix   = "pipeline:run:"
	keyRunsList    = "pipeline:runs:"
	keyAggregated  = "pipeline:agg:"
	keyConfigToken = "pipeline:config:github_token"
	keyConfigRepos = "pipeline:config:github_repos"
	maxRunsPerRepo = 100
)

// PipelineService gère la logique métier des pipelines
type PipelineService struct {
	cache    *cache.RedisClient
	cfg      *config.Config
	cfgStore *configstore.Client
	adapters map[models.Provider]adapter.PipelineAdapter
}

// NewPipelineService crée un nouveau service de pipelines.
func NewPipelineService(c *cache.RedisClient, cfg *config.Config) *PipelineService {
	adapters := map[models.Provider]adapter.PipelineAdapter{
		models.ProviderGitHub:  adapter.NewGitHubAdapter(),
		models.ProviderGitLab:  adapter.NewGitLabAdapter(),
		models.ProviderJenkins: adapter.NewJenkinsAdapter(),
		models.ProviderForgejo: adapter.NewForgejoAdapter(),
	}

	return &PipelineService{
		cache:    c,
		cfg:      cfg,
		cfgStore: configstore.New(cfg.AuthServiceURL, "pipeline"),
		adapters: adapters,
	}
}

// ProcessWebhook traite un webhook et stocke le run
func (s *PipelineService) ProcessWebhook(ctx context.Context, provider models.Provider, body []byte) (*models.PipelineRun, error) {
	a, ok := s.adapters[provider]
	if !ok {
		return nil, fmt.Errorf("provider inconnu: %s", provider)
	}

	run, err := a.ParseWebhook(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("échec du parsing du webhook: %w", err)
	}
	if run == nil {
		return nil, nil
	}

	// ID stable pour GitHub/Forgejo (évite doublons webhook + API)
	if (run.Provider == models.ProviderGitHub || run.Provider == models.ProviderForgejo) && run.ExternalID != "" {
		run.ID = string(run.Provider) + "_" + run.ExternalID
		if err := s.UpsertRun(ctx, run); err != nil {
			return nil, err
		}
	} else {
		run.ID = fmt.Sprintf("run_%d", time.Now().UnixNano())
		if err := s.StoreRun(ctx, run); err != nil {
			return nil, err
		}
	}

	log.Printf("✅ Run enregistré: %s %s %s [%s]", run.Provider, run.Repository, run.Branch, run.Status)

	return run, nil
}

// StoreRun stocke une exécution dans Redis
func (s *PipelineService) StoreRun(ctx context.Context, run *models.PipelineRun) error {
	data, err := json.Marshal(run)
	if err != nil {
		return err
	}

	runKey := keyRunPrefix + run.ID
	listKey := keyRunsList + string(run.Provider) + ":" + run.Repository + ":" + run.Branch
	aggKey := keyAggregated + string(run.Provider) + ":" + run.Repository + ":" + run.Branch

	if err := s.cache.Set(ctx, runKey, string(data), s.cfg.CacheTTL); err != nil {
		return err
	}

	// Ajouter à la liste des runs (FIFO, limitée)
	if err := s.cache.RPush(ctx, listKey, run.ID); err != nil {
		return err
	}
	if err := s.cache.Expire(ctx, listKey, s.cfg.CacheTTL); err != nil {
		return err
	}

	// Tronquer la liste si nécessaire
	if err := s.cache.LTrim(ctx, listKey, -int64(maxRunsPerRepo), -1); err != nil {
		return err
	}

	// Mettre à jour l'agrégation
	if err := s.updateAggregation(ctx, aggKey, run, true); err != nil {
		return err
	}

	return nil
}

// UpsertRun met à jour un run existant ou le stocke s'il est nouveau (évite doublons)
func (s *PipelineService) UpsertRun(ctx context.Context, run *models.PipelineRun) error {
	runKey := keyRunPrefix + run.ID
	existsVal, err := s.cache.Get(ctx, runKey)
	if err != nil {
		return err
	}

	isNew := existsVal == ""

	// Si mise à jour : incrémenter seulement si le statut passe à terminal (ex: running → success)
	incrementCounts := isNew
	if !isNew && (run.Status == models.StatusSuccess || run.Status == models.StatusFailure || run.Status == models.StatusCancelled) {
		var oldRun models.PipelineRun
		if json.Unmarshal([]byte(existsVal), &oldRun) == nil {
			wasTerminal := oldRun.Status == models.StatusSuccess || oldRun.Status == models.StatusFailure || oldRun.Status == models.StatusCancelled
			incrementCounts = !wasTerminal
		}
	}

	data, err := json.Marshal(run)
	if err != nil {
		return err
	}

	if err := s.cache.Set(ctx, runKey, string(data), s.cfg.CacheTTL); err != nil {
		return err
	}

	listKey := keyRunsList + string(run.Provider) + ":" + run.Repository + ":" + run.Branch
	aggKey := keyAggregated + string(run.Provider) + ":" + run.Repository + ":" + run.Branch

	if isNew {
		if err := s.cache.RPush(ctx, listKey, run.ID); err != nil {
			return err
		}
		if err := s.cache.Expire(ctx, listKey, s.cfg.CacheTTL); err != nil {
			return err
		}
		if err := s.cache.LTrim(ctx, listKey, -int64(maxRunsPerRepo), -1); err != nil {
			return err
		}
	}

	return s.updateAggregation(ctx, aggKey, run, incrementCounts)
}

// PipelineConfig config GitHub/Forgejo (tokens jamais exposés au GET)
type PipelineConfig struct {
	GitHubRepos []string `json:"github_repos"`
	Linked      bool     `json:"linked"` // true si token GitHub + au moins 1 repo

	ForgejoURL    string   `json:"forgejo_url"`
	ForgejoRepos  []string `json:"forgejo_repos"`
	ForgejoLinked bool     `json:"forgejo_linked"` // true si token Forgejo + URL + au moins 1 repo
}

// GetConfig retourne la config (sans tokens)
func (s *PipelineService) GetConfig(ctx context.Context) (*PipelineConfig, error) {
	all, err := s.cfgStore.GetAll(ctx)
	if err != nil {
		all = map[string]string{}
	}

	var repos []string
	if reposStr := all["github_repos"]; reposStr != "" {
		_ = json.Unmarshal([]byte(reposStr), &repos)
	}
	if len(repos) == 0 {
		repos = s.cfg.GitHubRepos
	}
	tokenSet := all["github_token"] != "" || s.cfg.GitHubToken != ""

	forgejoURL := all["forgejo_url"]
	if forgejoURL == "" {
		forgejoURL = s.cfg.ForgejoURL
	}
	var forgejoRepos []string
	if reposStr := all["forgejo_repos"]; reposStr != "" {
		_ = json.Unmarshal([]byte(reposStr), &forgejoRepos)
	}
	if len(forgejoRepos) == 0 {
		forgejoRepos = s.cfg.ForgejoRepos
	}
	forgejoTokenSet := all["forgejo_token"] != "" || s.cfg.ForgejoToken != ""

	return &PipelineConfig{
		GitHubRepos:   repos,
		Linked:        tokenSet && len(repos) > 0,
		ForgejoURL:    forgejoURL,
		ForgejoRepos:  forgejoRepos,
		ForgejoLinked: forgejoTokenSet && forgejoURL != "" && len(forgejoRepos) > 0,
	}, nil
}

// SetConfig enregistre token et/ou repos (depuis l'UI) dans Postgres via configstore.
func (s *PipelineService) SetConfig(ctx context.Context, token string, repos []string) error {
	kv := map[string]string{}
	if token != "" {
		kv["github_token"] = token
	}
	if repos != nil {
		data, err := json.Marshal(repos)
		if err != nil {
			return err
		}
		kv["github_repos"] = string(data)
	}
	if len(kv) == 0 {
		return nil
	}
	return s.cfgStore.SetMany(ctx, kv)
}

// SetForgejoConfig enregistre URL, token et/ou repos Forgejo (depuis l'UI) dans Postgres via configstore.
func (s *PipelineService) SetForgejoConfig(ctx context.Context, url, token string, repos []string) error {
	kv := map[string]string{}
	if url != "" {
		kv["forgejo_url"] = url
	}
	if token != "" {
		kv["forgejo_token"] = token
	}
	if repos != nil {
		data, err := json.Marshal(repos)
		if err != nil {
			return err
		}
		kv["forgejo_repos"] = string(data)
	}
	if len(kv) == 0 {
		return nil
	}
	return s.cfgStore.SetMany(ctx, kv)
}

// getGitHubConfig retourne token et repos (Postgres prioritaire sur env vars)
func (s *PipelineService) getGitHubConfig(ctx context.Context) (token string, repos []string) {
	all, _ := s.cfgStore.GetAll(ctx)

	token = all["github_token"]
	if token == "" {
		token = s.cfg.GitHubToken
	}

	if reposStr := all["github_repos"]; reposStr != "" {
		_ = json.Unmarshal([]byte(reposStr), &repos)
	}
	if len(repos) == 0 {
		repos = s.cfg.GitHubRepos
	}
	return token, repos
}

// getForgejoConfig retourne URL, token et repos Forgejo (Postgres prioritaire sur env vars)
func (s *PipelineService) getForgejoConfig(ctx context.Context) (baseURL, token string, repos []string) {
	all, _ := s.cfgStore.GetAll(ctx)

	baseURL = all["forgejo_url"]
	if baseURL == "" {
		baseURL = s.cfg.ForgejoURL
	}

	token = all["forgejo_token"]
	if token == "" {
		token = s.cfg.ForgejoToken
	}

	if reposStr := all["forgejo_repos"]; reposStr != "" {
		_ = json.Unmarshal([]byte(reposStr), &repos)
	}
	if len(repos) == 0 {
		repos = s.cfg.ForgejoRepos
	}
	return baseURL, token, repos
}

// SyncFromGitHub récupère les workflow runs via l'API GitHub et les stocke
func (s *PipelineService) SyncFromGitHub(ctx context.Context) (int, error) {
	token, repos := s.getGitHubConfig(ctx)
	if token == "" || len(repos) == 0 {
		return 0, nil
	}

	apiClient := client.NewGitHubAPIClient(token)
	count := 0

	for _, repo := range repos {
		parts := strings.SplitN(repo, "/", 2)
		if len(parts) != 2 {
			log.Printf("⚠️ GITHUB_REPOS ignoré (format invalide): %s", repo)
			continue
		}
		owner, repoName := parts[0], parts[1]

		runs, err := apiClient.FetchWorkflowRuns(owner, repoName, 30)
		if err != nil {
			log.Printf("⚠️ Sync GitHub %s: %v", repo, err)
			continue
		}

		for _, run := range runs {
			if err := s.UpsertRun(ctx, run); err != nil {
				log.Printf("⚠️ Upsert run %s: %v", run.ID, err)
				continue
			}
			count++
		}
	}

	if count > 0 {
		log.Printf("✅ Sync GitHub: %d runs synchronisés", count)
	}

	return count, nil
}

// SyncFromForgejo récupère les runs Forgejo Actions via l'API et les stocke
func (s *PipelineService) SyncFromForgejo(ctx context.Context) (int, error) {
	baseURL, token, repos := s.getForgejoConfig(ctx)
	if baseURL == "" || token == "" || len(repos) == 0 {
		return 0, nil
	}

	apiClient := client.NewForgejoAPIClient(baseURL, token)
	count := 0

	for _, repo := range repos {
		parts := strings.SplitN(repo, "/", 2)
		if len(parts) != 2 {
			log.Printf("⚠️ FORGEJO_REPOS ignoré (format invalide): %s", repo)
			continue
		}
		owner, repoName := parts[0], parts[1]

		runs, err := apiClient.FetchActionRuns(owner, repoName, 30)
		if err != nil {
			log.Printf("⚠️ Sync Forgejo %s: %v", repo, err)
			continue
		}

		for _, run := range runs {
			if err := s.UpsertRun(ctx, run); err != nil {
				log.Printf("⚠️ Upsert run %s: %v", run.ID, err)
				continue
			}
			count++
		}
	}

	if count > 0 {
		log.Printf("✅ Sync Forgejo: %d runs synchronisés", count)
	}

	return count, nil
}

func (s *PipelineService) updateAggregation(ctx context.Context, aggKey string, run *models.PipelineRun, incrementCounts bool) error {
	agg, err := s.cache.HGetAll(ctx, aggKey)
	if err != nil {
		return err
	}

	var successCount, failureCount int
	if v, ok := agg["success_count"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &successCount)
	}
	if v, ok := agg["failure_count"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &failureCount)
	}

	if incrementCounts {
		switch run.Status {
		case models.StatusSuccess:
			successCount++
		case models.StatusFailure, models.StatusCancelled:
			failureCount++
		}
	}

	lastRunAt := ""
	if run.FinishedAt != nil {
		lastRunAt = run.FinishedAt.Format(time.RFC3339)
	} else if run.StartedAt != nil {
		lastRunAt = run.StartedAt.Format(time.RFC3339)
	} else {
		lastRunAt = run.CreatedAt.Format(time.RFC3339)
	}

	fields := map[string]string{
		"repository":    run.Repository,
		"branch":        run.Branch,
		"provider":      string(run.Provider),
		"last_status":   string(run.Status),
		"last_run_id":   run.ID,
		"last_run_at":   lastRunAt,
		"success_count": fmt.Sprintf("%d", successCount),
		"failure_count": fmt.Sprintf("%d", failureCount),
		"total_runs":    fmt.Sprintf("%d", successCount+failureCount),
	}

	for k, v := range fields {
		if err := s.cache.HSet(ctx, aggKey, k, v); err != nil {
			return err
		}
	}

	return s.cache.Expire(ctx, aggKey, s.cfg.CacheTTL)
}

// GetRun récupère une exécution par ID
func (s *PipelineService) GetRun(ctx context.Context, id string) (*models.PipelineRun, error) {
	val, err := s.cache.Get(ctx, keyRunPrefix+id)
	if err != nil || val == "" {
		return nil, nil
	}

	var run models.PipelineRun
	if err := json.Unmarshal([]byte(val), &run); err != nil {
		return nil, err
	}

	return &run, nil
}

// ListRuns récupère les exécutions (optionnel: filtrer par provider, repo, branch)
func (s *PipelineService) ListRuns(ctx context.Context, provider, repository, branch string, limit int) ([]*models.PipelineRun, error) {
	if limit <= 0 {
		limit = 50
	}

	var runIDs []string
	if provider != "" && repository != "" {
		if branch != "" {
			key := keyRunsList + provider + ":" + repository + ":" + branch
			var err error
			runIDs, err = s.cache.LRange(ctx, key, -int64(limit), -1)
			if err != nil {
				return nil, err
			}
		} else {
			// Sans branch, on liste toutes les clés correspondantes
			pattern := keyRunsList + provider + ":" + repository + ":*"
			keys, err := s.cache.Keys(ctx, pattern)
			if err != nil {
				return nil, err
			}
			for _, k := range keys {
				ids, _ := s.cache.LRange(ctx, k, -int64(limit), -1)
				runIDs = append(runIDs, ids...)
			}
		}
	} else {
		// Liste globale : toutes les clés pipeline:runs:*
		keys, err := s.cache.Keys(ctx, keyRunsList+"*")
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			ids, _ := s.cache.LRange(ctx, k, -int64(limit), -1)
			runIDs = append(runIDs, ids...)
		}
	}

	// Dédupliquer et limiter
	seen := make(map[string]bool)
	var result []*models.PipelineRun
	for i := len(runIDs) - 1; i >= 0 && len(result) < limit; i-- {
		id := runIDs[i]
		if seen[id] {
			continue
		}
		seen[id] = true
		run, err := s.GetRun(ctx, id)
		if err != nil || run == nil {
			continue
		}
		result = append(result, run)
	}

	return result, nil
}

// GetAggregatedStatus récupère le statut agrégé pour un repo/branch
func (s *PipelineService) GetAggregatedStatus(ctx context.Context, provider, repository, branch string) (*models.AggregatedStatus, error) {
	key := keyAggregated + provider + ":" + repository + ":" + branch
	data, err := s.cache.HGetAll(ctx, key)
	if err != nil || len(data) == 0 {
		return nil, nil
	}

	var successCount, failureCount int
	fmt.Sscanf(data["success_count"], "%d", &successCount)
	fmt.Sscanf(data["failure_count"], "%d", &failureCount)

	agg := &models.AggregatedStatus{
		Repository:   data["repository"],
		Branch:       data["branch"],
		Provider:     models.Provider(data["provider"]),
		LastStatus:   models.RunStatus(data["last_status"]),
		LastRunID:    data["last_run_id"],
		SuccessCount: successCount,
		FailureCount: failureCount,
		TotalRuns:    successCount + failureCount,
	}

	if t, err := time.Parse(time.RFC3339, data["last_run_at"]); err == nil {
		agg.LastRunAt = &t
	}

	return agg, nil
}

// RerunRun relance un workflow run GitHub Actions ou Forgejo Actions via l'API du provider.
// Le run doit exister dans le cache.
// Sécurité : requiert le scope `workflow` (GitHub) ou un token avec droit d'écriture Actions (Forgejo).
func (s *PipelineService) RerunRun(ctx context.Context, runID string) error {
	run, err := s.GetRun(ctx, runID)
	if err != nil || run == nil {
		return fmt.Errorf("run introuvable: %s", runID)
	}
	if run.Provider != models.ProviderGitHub && run.Provider != models.ProviderForgejo {
		return fmt.Errorf("relance uniquement supportée pour GitHub Actions et Forgejo Actions (provider: %s)", run.Provider)
	}
	if run.ExternalID == "" {
		return fmt.Errorf("ID externe manquant pour ce run")
	}

	// Extraire owner/repo depuis run.Repository ("owner/repo")
	parts := strings.SplitN(run.Repository, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("format repository invalide: %s", run.Repository)
	}
	owner, repo := parts[0], parts[1]

	if run.Provider == models.ProviderForgejo {
		baseURL, token, _ := s.getForgejoConfig(ctx)
		if baseURL == "" || token == "" {
			return fmt.Errorf("connexion Forgejo non configurée — configurez-la dans la page Pipelines")
		}

		forgejoRunID, err := strconv.ParseInt(run.ExternalID, 10, 64)
		if err != nil {
			return fmt.Errorf("ID Forgejo invalide: %s", run.ExternalID)
		}

		fjClient := client.NewForgejoAPIClient(baseURL, token)
		if err := fjClient.RerunActionTask(owner, repo, forgejoRunID); err != nil {
			return fmt.Errorf("Forgejo API: %w", err)
		}

		log.Printf("▶ Pipeline relancé: %s/%s run #%s (Forgejo) par Kura", owner, repo, run.ExternalID)
		return nil
	}

	// Récupérer le token GitHub depuis le cache (configuré via l'UI)
	token, _ := s.getGitHubConfig(ctx)
	if token == "" {
		return fmt.Errorf("token GitHub non configuré — configurez-le dans la page Pipelines")
	}

	// Convertir l'ExternalID en int64
	githubRunID, err := strconv.ParseInt(run.ExternalID, 10, 64)
	if err != nil {
		return fmt.Errorf("ID GitHub invalide: %s", run.ExternalID)
	}

	ghClient := client.NewGitHubAPIClient(token)
	if err := ghClient.RerunWorkflowRun(owner, repo, githubRunID); err != nil {
		return fmt.Errorf("GitHub API: %w", err)
	}

	log.Printf("▶ Pipeline relancé: %s/%s run #%s par Kura", owner, repo, run.ExternalID)
	return nil
}

// DetectProvider tente de détecter le provider à partir des headers
func (s *PipelineService) DetectProvider(headerXGitHub, headerXGitLab, headerXForgejo string) models.Provider {
	if headerXGitHub != "" {
		return models.ProviderGitHub
	}
	if headerXGitLab != "" {
		return models.ProviderGitLab
	}
	if headerXForgejo != "" {
		return models.ProviderForgejo
	}
	// Jenkins : utiliser l'endpoint dédié /webhooks/jenkins pour les webhooks Jenkins
	return ""
}
