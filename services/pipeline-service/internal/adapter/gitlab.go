package adapter

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/modulops/pipeline-service/internal/models"
)

// GitLabPipelinePayload structure du webhook GitLab CI
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#pipeline-events
type GitLabPipelinePayload struct {
	ObjectKind       string `json:"object_kind"`
	ObjectAttributes *struct {
		ID         int    `json:"id"`
		Status     string `json:"status"`
		Ref        string `json:"ref"`
		SHA        string `json:"sha"`
		Duration   *int   `json:"duration"`
		CreatedAt  string `json:"created_at"`
		FinishedAt string `json:"finished_at"`
		Source     string `json:"source"`
	} `json:"object_attributes"`
	Project *struct {
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
	} `json:"project"`
	User *struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"user"`
	Commit *struct {
		Message string `json:"message"`
		Author  *struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commit"`
	Builds []struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
		Name   string `json:"name"`
	} `json:"builds"`
}

// GitLabJobPayload structure du webhook GitLab Job
type GitLabJobPayload struct {
	ObjectKind       string `json:"object_kind"`
	ObjectAttributes *struct {
		ID         int      `json:"id"`
		Status     string   `json:"status"`
		Ref        string   `json:"ref"`
		Stage      string   `json:"stage"`
		Name       string   `json:"name"`
		Duration   *float64 `json:"duration"`
		CreatedAt  string   `json:"created_at"`
		FinishedAt string   `json:"finished_at"`
		PipelineID int      `json:"pipeline_id"`
	} `json:"object_attributes"`
	Project *struct {
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
	} `json:"project"`
	Commit *struct {
		SHA     string `json:"sha"`
		Message string `json:"message"`
	} `json:"commit"`
}

// GitLabAdapter adapte les webhooks GitLab CI
type GitLabAdapter struct{}

// NewGitLabAdapter crée un nouvel adapter GitLab
func NewGitLabAdapter() *GitLabAdapter {
	return &GitLabAdapter{}
}

// Provider retourne le fournisseur
func (g *GitLabAdapter) Provider() models.Provider {
	return models.ProviderGitLab
}

// ParseWebhook parse le payload d'un webhook GitLab
func (g *GitLabAdapter) ParseWebhook(body io.Reader) (*models.PipelineRun, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	// Détecter le type d'événement
	var header struct {
		ObjectKind string `json:"object_kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}

	switch header.ObjectKind {
	case "pipeline":
		var payload GitLabPipelinePayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, err
		}
		return g.parsePipeline(&payload)
	case "build":
		var payload GitLabJobPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, err
		}
		return g.parseJob(&payload)
	}

	return nil, nil
}

func (g *GitLabAdapter) parsePipeline(p *GitLabPipelinePayload) (*models.PipelineRun, error) {
	oa := p.ObjectAttributes
	if oa == nil {
		return nil, nil
	}

	repo := ""
	if p.Project != nil {
		repo = p.Project.PathWithNamespace
	}

	status := g.mapStatus(oa.Status)
	run := &models.PipelineRun{
		Provider:   models.ProviderGitLab,
		Repository: repo,
		Branch:     oa.Ref,
		CommitSHA:  oa.SHA,
		Status:     status,
		ExternalID: strconv.Itoa(oa.ID),
		CreatedAt:  time.Now(),
	}

	if p.Commit != nil {
		run.CommitMsg = p.Commit.Message
		if p.Commit.Author != nil {
			run.Author = p.Commit.Author.Name
		}
	}
	if p.User != nil && run.Author == "" {
		run.Author = p.User.Username
	}
	if p.Project != nil {
		run.ExternalURL = p.Project.WebURL + "/-/pipelines/" + strconv.Itoa(oa.ID)
	}
	if oa.Duration != nil {
		run.Duration = int64(*oa.Duration) * 1000 // secondes -> ms
	}
	if oa.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, oa.CreatedAt); err == nil {
			run.StartedAt = &t
		}
	}
	if oa.FinishedAt != "" {
		if t, err := time.Parse(time.RFC3339, oa.FinishedAt); err == nil {
			run.FinishedAt = &t
		}
	}

	return run, nil
}

func (g *GitLabAdapter) parseJob(p *GitLabJobPayload) (*models.PipelineRun, error) {
	oa := p.ObjectAttributes
	if oa == nil {
		return nil, nil
	}

	repo := ""
	if p.Project != nil {
		repo = p.Project.PathWithNamespace
	}

	status := g.mapStatus(oa.Status)
	run := &models.PipelineRun{
		Provider:     models.ProviderGitLab,
		Repository:   repo,
		Branch:       oa.Ref,
		Status:       status,
		ExternalID:   strconv.Itoa(oa.ID),
		WorkflowName: oa.Name + " (" + oa.Stage + ")",
		CreatedAt:    time.Now(),
	}

	if p.Commit != nil {
		run.CommitSHA = p.Commit.SHA
		run.CommitMsg = p.Commit.Message
	}
	if p.Project != nil {
		run.ExternalURL = p.Project.WebURL + "/-/jobs/" + strconv.Itoa(oa.ID)
	}
	if oa.Duration != nil {
		run.Duration = int64(*oa.Duration * 1000)
	}

	return run, nil
}

func (g *GitLabAdapter) mapStatus(s string) models.RunStatus {
	switch s {
	case "pending", "created":
		return models.StatusPending
	case "running":
		return models.StatusRunning
	case "success":
		return models.StatusSuccess
	case "failed":
		return models.StatusFailure
	case "canceled", "cancelled":
		return models.StatusCancelled
	case "skipped":
		return models.StatusSkipped
	default:
		return models.StatusPending
	}
}
