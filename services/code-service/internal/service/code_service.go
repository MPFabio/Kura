package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/modulops/code-service/internal/client"
	"github.com/modulops/code-service/internal/config"
	"github.com/modulops/code-service/internal/configstore"
	"github.com/modulops/code-service/internal/models"
)

// CodeService fournit l'accès en lecture seule aux dépôts GitHub liés aux projets.
type CodeService struct {
	cfg        *config.Config
	cfgStore   *configstore.Client
	httpClient *http.Client
}

// New crée un CodeService.
func New(cfg *config.Config) *CodeService {
	return &CodeService{
		cfg: cfg,
		// Réutilise le token GitHub configuré pour le drift Terraform : un seul
		// token GitHub à gérer pour l'utilisateur, partagé entre les modules.
		cfgStore:   configstore.New(cfg.AuthServiceURL, "terraform"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// githubClient construit un client GitHub avec le token configuré.
func (s *CodeService) githubClient(ctx context.Context) (*client.GitHubClient, error) {
	token, err := s.cfgStore.Get(ctx, "github_token")
	if err != nil {
		return nil, fmt.Errorf("lecture du token GitHub: %w", err)
	}
	return client.NewGitHubClient(token), nil
}

// ListRepositories liste les dépôts GitHub liés à un projet, via les ProjectMapping de l'auth-service.
func (s *CodeService) ListRepositories(ctx context.Context, authToken, projectID string) ([]models.Repository, error) {
	url := fmt.Sprintf("%s/api/v1/projects/%s/mappings", s.cfg.AuthServiceURL, projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel auth-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service mappings: %s", resp.Status)
	}

	var body struct {
		Items []models.ProjectMapping `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	repos := make([]models.Repository, 0, len(body.Items))
	for _, m := range body.Items {
		if m.GitHubRepository == "" {
			continue
		}
		repos = append(repos, models.Repository{MappingID: m.ID, FullName: m.GitHubRepository})
	}
	return repos, nil
}

// GetTree liste le contenu d'un répertoire d'un dépôt.
func (s *CodeService) GetTree(ctx context.Context, repo, path, ref string) ([]client.TreeEntry, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	gh, err := s.githubClient(ctx)
	if err != nil {
		return nil, err
	}
	return gh.GetTree(owner, name, path, ref)
}

// GetFile récupère le contenu d'un fichier d'un dépôt.
func (s *CodeService) GetFile(ctx context.Context, repo, path, ref string) (*client.FileContent, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	gh, err := s.githubClient(ctx)
	if err != nil {
		return nil, err
	}
	return gh.GetFileContent(owner, name, path, ref)
}

// GetCommits récupère l'historique des commits d'un dépôt.
func (s *CodeService) GetCommits(ctx context.Context, repo, path, ref string, page int) ([]client.Commit, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	gh, err := s.githubClient(ctx)
	if err != nil {
		return nil, err
	}
	return gh.GetCommits(owner, name, path, ref, page)
}

// GetCommitDiff récupère le détail d'un commit.
func (s *CodeService) GetCommitDiff(ctx context.Context, repo, sha string) (*client.CommitDetail, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	gh, err := s.githubClient(ctx)
	if err != nil {
		return nil, err
	}
	return gh.GetCommitDiff(owner, name, sha)
}

// splitRepo découpe "owner/repo" en (owner, repo).
func splitRepo(repo string) (string, string, error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("paramètre repo invalide, attendu \"owner/repo\": %q", repo)
	}
	return parts[0], parts[1], nil
}
