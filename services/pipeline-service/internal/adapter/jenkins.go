package adapter

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/modulops/pipeline-service/internal/models"
)

// JenkinsGenericPayload structure générique pour les webhooks Jenkins
// Jenkins peut envoyer différents formats selon les plugins
type JenkinsGenericPayload struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`
	Build       *struct {
		FullURL   string `json:"full_url"`
		Number    int    `json:"number"`
		Phase     string `json:"phase"`  // STARTED, COMPLETED, FINALIZED
		Status    string `json:"status"` // SUCCESS, FAILURE, UNSTABLE, etc.
		Duration  int64  `json:"duration"`
		Timestamp int64  `json:"timestamp"`
		Scm       *struct {
			Branch string `json:"branch"`
			Commit string `json:"commit"`
			URL    string `json:"url"`
		} `json:"scm"`
	} `json:"build"`
}

// JenkinsNotificationPayload format des notifications Jenkins (Jenkins Notification Plugin)
type JenkinsNotificationPayload struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Build *struct {
		FullURL    string `json:"fullUrl"`
		Number     int    `json:"number"`
		Phase      string `json:"phase"`
		Status     string `json:"status"`
		Duration   int64  `json:"duration"`
		Timestamp  int64  `json:"timestamp"`
		Parameters *struct {
			Branch []string `json:"BRANCH"`
			Commit []string `json:"COMMIT"`
		} `json:"parameters"`
	} `json:"build"`
}

// JenkinsAdapter adapte les webhooks Jenkins
type JenkinsAdapter struct{}

// NewJenkinsAdapter crée un nouvel adapter Jenkins
func NewJenkinsAdapter() *JenkinsAdapter {
	return &JenkinsAdapter{}
}

// Provider retourne le fournisseur
func (j *JenkinsAdapter) Provider() models.Provider {
	return models.ProviderJenkins
}

// ParseWebhook parse le payload d'un webhook Jenkins
func (j *JenkinsAdapter) ParseWebhook(body io.Reader) (*models.PipelineRun, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	// Essayer le format avec full_url (snake_case)
	var genPayload JenkinsGenericPayload
	if err := json.Unmarshal(data, &genPayload); err == nil && genPayload.Build != nil {
		return j.parseGeneric(&genPayload)
	}

	// Essayer le format avec fullUrl (camelCase)
	var notifPayload JenkinsNotificationPayload
	if err := json.Unmarshal(data, &notifPayload); err == nil && notifPayload.Build != nil {
		return j.parseNotification(&notifPayload)
	}

	return nil, nil
}

func (j *JenkinsAdapter) parseGeneric(p *JenkinsGenericPayload) (*models.PipelineRun, error) {
	b := p.Build
	status := j.mapStatus(b.Phase, b.Status)

	run := &models.PipelineRun{
		Provider:     models.ProviderJenkins,
		Repository:   p.Name,
		WorkflowName: p.DisplayName,
		Status:       status,
		ExternalID:   "",
		ExternalURL:  b.FullURL,
		Duration:     b.Duration,
		CreatedAt:    time.UnixMilli(b.Timestamp),
	}

	run.ExternalID = strconv.Itoa(b.Number)

	if b.Scm != nil {
		run.Branch = b.Scm.Branch
		run.CommitSHA = b.Scm.Commit
	}

	if b.Phase == "COMPLETED" || b.Phase == "FINALIZED" {
		finished := time.UnixMilli(b.Timestamp + b.Duration)
		run.FinishedAt = &finished
	}

	return run, nil
}

func (j *JenkinsAdapter) parseNotification(p *JenkinsNotificationPayload) (*models.PipelineRun, error) {
	b := p.Build
	status := j.mapStatus(b.Phase, b.Status)

	run := &models.PipelineRun{
		Provider:    models.ProviderJenkins,
		Repository:  p.Name,
		Status:      status,
		ExternalID:  strconv.Itoa(b.Number),
		ExternalURL: b.FullURL,
		Duration:    b.Duration,
		CreatedAt:   time.UnixMilli(b.Timestamp),
	}

	if b.Parameters != nil {
		if len(b.Parameters.Branch) > 0 {
			run.Branch = b.Parameters.Branch[0]
		}
		if len(b.Parameters.Commit) > 0 {
			run.CommitSHA = b.Parameters.Commit[0]
		}
	}

	return run, nil
}

func (j *JenkinsAdapter) mapStatus(phase, status string) models.RunStatus {
	if phase == "STARTED" {
		return models.StatusRunning
	}
	if phase == "COMPLETED" || phase == "FINALIZED" {
		switch status {
		case "SUCCESS":
			return models.StatusSuccess
		case "FAILURE", "ABORTED":
			return models.StatusFailure
		case "UNSTABLE":
			return models.StatusFailure
		default:
			return models.StatusFailure
		}
	}
	return models.StatusPending
}
