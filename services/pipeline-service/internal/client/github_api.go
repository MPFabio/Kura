package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/modulops/pipeline-service/internal/models"
)

const githubAPIBase = "https://api.github.com"

// WorkflowRunsResponse structure de la réponse API GitHub
type WorkflowRunsResponse struct {
	TotalCount   int `json:"total_count"`
	WorkflowRuns []struct {
		ID           int64     `json:"id"`
		Name         string    `json:"name"`
		Status       string    `json:"status"`
		Conclusion   string    `json:"conclusion"`
		HTMLURL      string    `json:"html_url"`
		RunStartedAt time.Time `json:"run_started_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		HeadBranch   string    `json:"head_branch"`
		HeadSHA      string    `json:"head_sha"`
		DisplayTitle string    `json:"display_title"`
		Repository   struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Actor struct {
			Login string `json:"login"`
		} `json:"actor"`
		TriggeringActor *struct {
			Login string `json:"login"`
		} `json:"triggering_actor"`
	} `json:"workflow_runs"`
}

// GitHubAPIClient client pour l'API GitHub Actions
type GitHubAPIClient struct {
	token  string
	client *http.Client
}

// NewGitHubAPIClient crée un client API GitHub
func NewGitHubAPIClient(token string) *GitHubAPIClient {
	return &GitHubAPIClient{
		token: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchWorkflowRuns récupère les workflow runs d'un repo
func (c *GitHubAPIClient) FetchWorkflowRuns(owner, repo string, perPage int) ([]*models.PipelineRun, error) {
	if perPage <= 0 {
		perPage = 30
	}
	url := fmt.Sprintf("%s/repos/%s/%s/actions/runs?per_page=%d", githubAPIBase, owner, repo, perPage)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API GitHub: %s", resp.Status)
	}

	var data WorkflowRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	runs := make([]*models.PipelineRun, 0, len(data.WorkflowRuns))
	for i := range data.WorkflowRuns {
		wr := &data.WorkflowRuns[i]
		repoName := wr.Repository.FullName
		if repoName == "" {
			repoName = owner + "/" + repo
		}

		status := mapAPIToStatus(wr.Status, wr.Conclusion)
		run := &models.PipelineRun{
			ID:           fmt.Sprintf("github_%d", wr.ID),
			Provider:     models.ProviderGitHub,
			Repository:   repoName,
			Branch:       wr.HeadBranch,
			CommitSHA:    wr.HeadSHA,
			CommitMsg:    wr.DisplayTitle,
			Status:       status,
			ExternalID:   fmt.Sprintf("%d", wr.ID),
			ExternalURL:  wr.HTMLURL,
			WorkflowName: wr.Name,
			StartedAt:    &wr.RunStartedAt,
			FinishedAt:   &wr.UpdatedAt,
			CreatedAt:    wr.RunStartedAt,
		}

		if wr.TriggeringActor != nil && wr.TriggeringActor.Login != "" {
			run.Author = wr.TriggeringActor.Login
		} else if wr.Actor.Login != "" {
			run.Author = wr.Actor.Login
		}

		if wr.Status == "completed" && !wr.RunStartedAt.IsZero() && !wr.UpdatedAt.IsZero() {
			run.Duration = wr.UpdatedAt.Sub(wr.RunStartedAt).Milliseconds()
		}

		runs = append(runs, run)
	}

	return runs, nil
}

func mapAPIToStatus(status, conclusion string) models.RunStatus {
	if status == "completed" {
		switch conclusion {
		case "success":
			return models.StatusSuccess
		case "failure", "timed_out":
			return models.StatusFailure
		case "cancelled":
			return models.StatusCancelled
		case "skipped":
			return models.StatusSkipped
		default:
			return models.StatusFailure
		}
	}
	if status == "in_progress" || status == "queued" || status == "requested" || status == "waiting" || status == "pending" {
		return models.StatusRunning
	}
	return models.StatusPending
}
