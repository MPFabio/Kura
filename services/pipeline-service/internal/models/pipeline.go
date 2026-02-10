package models

import "time"

// Provider représente le type de fournisseur CI/CD
type Provider string

const (
	ProviderGitHub  Provider = "github"
	ProviderGitLab  Provider = "gitlab"
	ProviderJenkins Provider = "jenkins"
)

// RunStatus représente le statut d'une exécution
type RunStatus string

const (
	StatusPending   RunStatus = "pending"
	StatusRunning   RunStatus = "running"
	StatusSuccess   RunStatus = "success"
	StatusFailure   RunStatus = "failure"
	StatusCancelled RunStatus = "cancelled"
	StatusSkipped   RunStatus = "skipped"
)

// PipelineRun représente une exécution de pipeline
type PipelineRun struct {
	ID           string     `json:"id"`
	Provider     Provider   `json:"provider"`
	Repository   string     `json:"repository"`
	Branch       string     `json:"branch"`
	CommitSHA    string     `json:"commit_sha,omitempty"`
	CommitMsg    string     `json:"commit_msg,omitempty"`
	Author       string     `json:"author,omitempty"`
	Status       RunStatus  `json:"status"`
	ExternalID   string     `json:"external_id,omitempty"`   // ID du run dans le provider
	ExternalURL  string     `json:"external_url,omitempty"`  // URL vers le run dans le provider
	WorkflowName string     `json:"workflow_name,omitempty"` // Nom du workflow/job
	Duration     int64      `json:"duration_ms,omitempty"`   // Durée en millisecondes
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// AggregatedStatus représente un statut agrégé pour un repository
type AggregatedStatus struct {
	Repository   string     `json:"repository"`
	Branch       string     `json:"branch"`
	Provider     Provider   `json:"provider"`
	LastStatus   RunStatus  `json:"last_status"`
	LastRunID    string     `json:"last_run_id"`
	LastRunAt    *time.Time `json:"last_run_at,omitempty"`
	SuccessCount int        `json:"success_count"`
	FailureCount int        `json:"failure_count"`
	TotalRuns    int        `json:"total_runs"`
}

// WebhookPayload interface pour les payloads de webhook
type WebhookPayload interface {
	ParseProvider() Provider
	ParseRun() (*PipelineRun, error)
}
