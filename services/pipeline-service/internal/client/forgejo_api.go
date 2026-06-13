package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/modulops/pipeline-service/internal/models"
)

// ActionTasksResponse structure de la réponse API Forgejo
// GET /api/v1/repos/{owner}/{repo}/actions/tasks
type ActionTasksResponse struct {
	TotalCount   int64 `json:"total_count"`
	WorkflowRuns []struct {
		ID           int64     `json:"id"`
		Name         string    `json:"name"`
		Status       string    `json:"status"`
		HTMLURL      string    `json:"html_url"`
		HeadBranch   string    `json:"head_branch"`
		HeadSHA      string    `json:"head_sha"`
		Event        string    `json:"event"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		RunStartedAt time.Time `json:"run_started_at"`
		Repository   *struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		TriggerActor *struct {
			Login string `json:"login"`
		} `json:"triggered_by,omitempty"`
		Actor *struct {
			Login string `json:"login"`
		} `json:"actor,omitempty"`
	} `json:"workflow_runs"`
}

// ForgejoAPIClient client pour l'API Forgejo Actions
type ForgejoAPIClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewForgejoAPIClient crée un client API Forgejo
func NewForgejoAPIClient(baseURL, token string) *ForgejoAPIClient {
	return &ForgejoAPIClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchActionRuns récupère les runs Forgejo Actions d'un repo
func (c *ForgejoAPIClient) FetchActionRuns(owner, repo string, limit int) ([]*models.PipelineRun, error) {
	if limit <= 0 {
		limit = 30
	}
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/actions/tasks?limit=%d", c.baseURL, owner, repo, limit)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Forgejo: %s", resp.Status)
	}

	var data ActionTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	runs := make([]*models.PipelineRun, 0, len(data.WorkflowRuns))
	for i := range data.WorkflowRuns {
		wr := &data.WorkflowRuns[i]
		repoName := ""
		if wr.Repository != nil {
			repoName = wr.Repository.FullName
		}
		if repoName == "" {
			repoName = owner + "/" + repo
		}

		startedAt := wr.RunStartedAt
		if startedAt.IsZero() {
			startedAt = wr.CreatedAt
		}

		run := &models.PipelineRun{
			ID:           fmt.Sprintf("forgejo_%d", wr.ID),
			Provider:     models.ProviderForgejo,
			Repository:   repoName,
			Branch:       wr.HeadBranch,
			CommitSHA:    wr.HeadSHA,
			Status:       mapForgejoStatus(wr.Status),
			ExternalID:   fmt.Sprintf("%d", wr.ID),
			ExternalURL:  wr.HTMLURL,
			WorkflowName: wr.Name,
			StartedAt:    &startedAt,
			FinishedAt:   &wr.UpdatedAt,
			CreatedAt:    wr.CreatedAt,
		}

		if wr.TriggerActor != nil && wr.TriggerActor.Login != "" {
			run.Author = wr.TriggerActor.Login
		} else if wr.Actor != nil && wr.Actor.Login != "" {
			run.Author = wr.Actor.Login
		}

		if isTerminalForgejoStatus(wr.Status) && !startedAt.IsZero() && !wr.UpdatedAt.IsZero() {
			run.Duration = wr.UpdatedAt.Sub(startedAt).Milliseconds()
		}

		runs = append(runs, run)
	}

	return runs, nil
}

// RerunActionTask déclenche le re-run d'un workflow run Forgejo Actions.
// POST /api/v1/repos/{owner}/{repo}/actions/tasks/{task_id}/rerun
func (c *ForgejoAPIClient) RerunActionTask(owner, repo string, taskID int64) error {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/actions/tasks/%d/rerun", c.baseURL, owner, repo, taskID)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("création requête rerun: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.token == "" {
		return fmt.Errorf("token Forgejo requis pour relancer un workflow")
	}
	req.Header.Set("Authorization", "token "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("appel API Forgejo rerun: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("permission refusée — le token Forgejo doit avoir le droit d'écriture sur les Actions")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Forgejo rerun: %s", resp.Status)
	}
	return nil
}

// mapForgejoStatus convertit le statut Forgejo Actions (status only, pas de "conclusion" séparé)
// Valeurs possibles : success, failure, cancelled, skipped, running, waiting, blocked
func mapForgejoStatus(status string) models.RunStatus {
	switch status {
	case "success":
		return models.StatusSuccess
	case "failure", "timed_out":
		return models.StatusFailure
	case "cancelled":
		return models.StatusCancelled
	case "skipped":
		return models.StatusSkipped
	case "running":
		return models.StatusRunning
	case "waiting", "blocked":
		return models.StatusPending
	default:
		return models.StatusPending
	}
}

func isTerminalForgejoStatus(status string) bool {
	switch status {
	case "success", "failure", "timed_out", "cancelled", "skipped":
		return true
	default:
		return false
	}
}
