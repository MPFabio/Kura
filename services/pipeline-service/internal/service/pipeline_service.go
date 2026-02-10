package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/modulops/pipeline-service/internal/adapter"
	"github.com/modulops/pipeline-service/internal/cache"
	"github.com/modulops/pipeline-service/internal/config"
	"github.com/modulops/pipeline-service/internal/models"
)

const (
	keyRunPrefix   = "pipeline:run:"
	keyRunsList    = "pipeline:runs:"
	keyAggregated  = "pipeline:agg:"
	maxRunsPerRepo = 100
)

// PipelineService gère la logique métier des pipelines
type PipelineService struct {
	cache    *cache.RedisClient
	cfg      *config.Config
	adapters map[models.Provider]adapter.PipelineAdapter
}

// NewPipelineService crée un nouveau service de pipelines
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

	run.ID = fmt.Sprintf("run_%d", time.Now().UnixNano())

	if err := s.StoreRun(ctx, run); err != nil {
		return nil, err
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
	if err := s.updateAggregation(ctx, aggKey, run); err != nil {
		return err
	}

	return nil
}

func (s *PipelineService) updateAggregation(ctx context.Context, aggKey string, run *models.PipelineRun) error {
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

	switch run.Status {
	case models.StatusSuccess:
		successCount++
	case models.StatusFailure, models.StatusCancelled:
		failureCount++
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
