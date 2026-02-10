package adapter

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/modulops/pipeline-service/internal/models"
)

// GitHubWorkflowPayload structure du webhook GitHub Actions
// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#workflow_run
type GitHubWorkflowPayload struct {
	Action      string `json:"action"`
	WorkflowRun *struct {
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
		Repository   *struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		TriggeringActor *struct {
			Login string `json:"login"`
		} `json:"triggering_actor"`
	} `json:"workflow_run"`
	Repository *struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// GitHubCheckSuitePayload structure du webhook check_suite (alternative)
type GitHubCheckSuitePayload struct {
	Action     string `json:"action"`
	CheckSuite *struct {
		ID         int64  `json:"id"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		HeadSHA    string `json:"head_sha"`
		HeadBranch string `json:"head_branch"`
		HTMLURL    string `json:"html_url"`
		UpdatedAt  string `json:"updated_at"`
	} `json:"check_suite"`
	Repository *struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// GitHubAdapter adapte les webhooks GitHub Actions
type GitHubAdapter struct{}

// NewGitHubAdapter crée un nouvel adapter GitHub
func NewGitHubAdapter() *GitHubAdapter {
	return &GitHubAdapter{}
}

// Provider retourne le fournisseur
func (g *GitHubAdapter) Provider() models.Provider {
	return models.ProviderGitHub
}

// ParseWebhook parse le payload d'un webhook GitHub
func (g *GitHubAdapter) ParseWebhook(body io.Reader) (*models.PipelineRun, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	// Essayer workflow_run en premier
	var workflowPayload GitHubWorkflowPayload
	if err := json.Unmarshal(data, &workflowPayload); err == nil && workflowPayload.WorkflowRun != nil {
		return g.parseWorkflowRun(&workflowPayload)
	}

	// Essayer check_suite
	var checkPayload GitHubCheckSuitePayload
	if err := json.Unmarshal(data, &checkPayload); err == nil && checkPayload.CheckSuite != nil {
		return g.parseCheckSuite(&checkPayload)
	}

	return nil, nil
}

func (g *GitHubAdapter) parseWorkflowRun(p *GitHubWorkflowPayload) (*models.PipelineRun, error) {
	wr := p.WorkflowRun
	repo := wr.Repository.FullName
	if repo == "" && p.Repository != nil {
		repo = p.Repository.FullName
	}

	status := g.mapStatus(wr.Status, wr.Conclusion)
	run := &models.PipelineRun{
		Provider:     models.ProviderGitHub,
		Repository:   repo,
		Branch:       wr.HeadBranch,
		CommitSHA:    wr.HeadSHA,
		CommitMsg:    wr.DisplayTitle,
		Status:       status,
		ExternalID:   strconv.FormatInt(wr.ID, 10),
		ExternalURL:  wr.HTMLURL,
		WorkflowName: wr.Name,
		StartedAt:    &wr.RunStartedAt,
		FinishedAt:   &wr.UpdatedAt,
		CreatedAt:    wr.RunStartedAt,
	}

	if wr.TriggeringActor != nil {
		run.Author = wr.TriggeringActor.Login
	}

	if wr.Status == "completed" && !wr.RunStartedAt.IsZero() && !wr.UpdatedAt.IsZero() {
		run.Duration = wr.UpdatedAt.Sub(wr.RunStartedAt).Milliseconds()
	}

	return run, nil
}

func (g *GitHubAdapter) parseCheckSuite(p *GitHubCheckSuitePayload) (*models.PipelineRun, error) {
	cs := p.CheckSuite
	repo := ""
	if p.Repository != nil {
		repo = p.Repository.FullName
	}

	status := g.mapStatus(cs.Status, cs.Conclusion)
	run := &models.PipelineRun{
		Provider:    models.ProviderGitHub,
		Repository:  repo,
		Branch:      cs.HeadBranch,
		CommitSHA:   cs.HeadSHA,
		Status:      status,
		ExternalID:  strconv.FormatInt(cs.ID, 10),
		ExternalURL: cs.HTMLURL,
		CreatedAt:   time.Now(),
	}

	return run, nil
}

func (g *GitHubAdapter) mapStatus(status, conclusion string) models.RunStatus {
	if status == "completed" {
		switch conclusion {
		case "success":
			return models.StatusSuccess
		case "failure":
			return models.StatusFailure
		case "cancelled":
			return models.StatusCancelled
		case "skipped":
			return models.StatusSkipped
		default:
			return models.StatusFailure
		}
	}
	if status == "in_progress" || status == "queued" {
		return models.StatusRunning
	}
	return models.StatusPending
}
