package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/modulops/pipeline-service/internal/adapter"
	"github.com/modulops/pipeline-service/internal/cache"
	"github.com/modulops/pipeline-service/internal/client"
	"github.com/modulops/pipeline-service/internal/config"
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
	adapters map[models.Provider]adapter.PipelineAdapter
}

// NewPipelineService crée un nouveau service de pipelines.
// La synchronisation périodique est gérée au niveau de main.go via un contexte
// annulable, ce qui garantit l'arrêt gracieux de la goroutine de sync.
func NewPipelineService(c *cache.RedisClient, cfg *config.Config) *PipelineService {
	adapters := map[models.Provider]adapter.PipelineAdapter{
		models.ProviderGitHub:  adapter.NewGitHubAdapter(),
		models.ProviderGitLab:  adapter.NewGitLabAdapter(),
		models.ProviderJenkins: adapter.NewJenkinsAdapter(),
	}

	return &PipelineService{
		cache:    c,
		cfg:      cfg,
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

	// ID stable pour GitHub (évite doublons webhook + API)
	if run.Provider == models.ProviderGitHub && run.ExternalID != "" {
		run.ID = "github_" + run.ExternalID
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

// PipelineConfig config GitHub (token jamais exposé au GET)
type PipelineConfig struct {
	GitHubRepos []string `json:"github_repos"`
	Linked      bool     `json:"linked"` // true si token + au moins 1 repo
}

// GetConfig retourne la config (sans token)
func (s *PipelineService) GetConfig(ctx context.Context) (*PipelineConfig, error) {
	token, _ := s.cache.Get(ctx, keyConfigToken)
	reposStr, _ := s.cache.Get(ctx, keyConfigRepos)

	var repos []string
	if reposStr != "" {
		_ = json.Unmarshal([]byte(reposStr), &repos)
	}

	// Fallback env si pas de config UI
	if len(repos) == 0 && len(s.cfg.GitHubRepos) > 0 {
		repos = s.cfg.GitHubRepos
	}
	tokenSet := token != "" || s.cfg.GitHubToken != ""

	return &PipelineConfig{
		GitHubRepos: repos,
		Linked:      tokenSet && len(repos) > 0,
	}, nil
}

// SetConfig enregistre token et/ou repos (depuis l'UI)
func (s *PipelineService) SetConfig(ctx context.Context, token string, repos []string) error {
	ttl := 365 * 24 * time.Hour
	if token != "" {
		if err := s.cache.Set(ctx, keyConfigToken, token, ttl); err != nil {
			return err
		}
	}
	// repos == nil signifie "ne pas modifier" ; repos != nil (y compris []) met à jour
	if repos != nil {
		data, err := json.Marshal(repos)
		if err != nil {
			return err
		}
		if err := s.cache.Set(ctx, keyConfigRepos, string(data), ttl); err != nil {
			return err
		}
	}
	return nil
}

// getGitHubConfig retourne token et repos (Redis prioritaire sur env)
func (s *PipelineService) getGitHubConfig(ctx context.Context) (token string, repos []string) {
	token, _ = s.cache.Get(ctx, keyConfigToken)
	if token == "" {
		token = s.cfg.GitHubToken
	}

	reposStr, _ := s.cache.Get(ctx, keyConfigRepos)
	if reposStr != "" {
		_ = json.Unmarshal([]byte(reposStr), &repos)
	}
	if len(repos) == 0 {
		repos = s.cfg.GitHubRepos
	}
	return token, repos
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

// DetectProvider tente de détecter le provider à partir des headers
func (s *PipelineService) DetectProvider(headerXGitHub, headerXGitLab string) models.Provider {
	if headerXGitHub != "" {
		return models.ProviderGitHub
	}
	if headerXGitLab != "" {
		return models.ProviderGitLab
	}
	// Jenkins : utiliser l'endpoint dédié /webhooks/jenkins pour les webhooks Jenkins
	return ""
}
