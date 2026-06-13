package adapter

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/modulops/pipeline-service/internal/models"
)

// ForgejoWorkflowPayload structure du webhook Forgejo Actions (workflow_run)
// Compatible avec le format Gitea/Forgejo : pas de champ "conclusion" séparé,
// le statut terminal est directement dans "status".
type ForgejoWorkflowPayload struct {
	Action      string `json:"action"`
	WorkflowRun *struct {
		ID           int64     `json:"id"`
		Name         string    `json:"name"`
		Status       string    `json:"status"`
		HTMLURL      string    `json:"html_url"`
		HeadBranch   string    `json:"head_branch"`
		HeadSHA      string    `json:"head_sha"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		RunStartedAt time.Time `json:"run_started_at"`
		Repository   *struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	} `json:"workflow_run"`
	Repository *struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender *struct {
		Login string `json:"login"`
	} `json:"sender"`
}

// ForgejoAdapter adapte les webhooks Forgejo Actions
type ForgejoAdapter struct{}

// NewForgejoAdapter crée un nouvel adapter Forgejo
func NewForgejoAdapter() *ForgejoAdapter {
	return &ForgejoAdapter{}
}

// Provider retourne le fournisseur
func (f *ForgejoAdapter) Provider() models.Provider {
	return models.ProviderForgejo
}

// ParseWebhook parse le payload d'un webhook Forgejo Actions
func (f *ForgejoAdapter) ParseWebhook(body io.Reader) (*models.PipelineRun, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var payload ForgejoWorkflowPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if payload.WorkflowRun == nil {
		return nil, nil
	}

	return f.parseWorkflowRun(&payload)
}

func (f *ForgejoAdapter) parseWorkflowRun(p *ForgejoWorkflowPayload) (*models.PipelineRun, error) {
	wr := p.WorkflowRun
	repo := ""
	if wr.Repository != nil {
		repo = wr.Repository.FullName
	}
	if repo == "" && p.Repository != nil {
		repo = p.Repository.FullName
	}

	startedAt := wr.RunStartedAt
	if startedAt.IsZero() {
		startedAt = wr.CreatedAt
	}

	run := &models.PipelineRun{
		Provider:     models.ProviderForgejo,
		Repository:   repo,
		Branch:       wr.HeadBranch,
		CommitSHA:    wr.HeadSHA,
		Status:       f.mapStatus(wr.Status),
		ExternalID:   strconv.FormatInt(wr.ID, 10),
		ExternalURL:  wr.HTMLURL,
		WorkflowName: wr.Name,
		StartedAt:    &startedAt,
		FinishedAt:   &wr.UpdatedAt,
		CreatedAt:    wr.CreatedAt,
	}

	if p.Sender != nil {
		run.Author = p.Sender.Login
	}

	if f.isTerminal(wr.Status) && !startedAt.IsZero() && !wr.UpdatedAt.IsZero() {
		run.Duration = wr.UpdatedAt.Sub(startedAt).Milliseconds()
	}

	return run, nil
}

// mapStatus convertit le statut Forgejo Actions (status seul, sans "conclusion")
func (f *ForgejoAdapter) mapStatus(status string) models.RunStatus {
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

func (f *ForgejoAdapter) isTerminal(status string) bool {
	switch status {
	case "success", "failure", "timed_out", "cancelled", "skipped":
		return true
	default:
		return false
	}
}
