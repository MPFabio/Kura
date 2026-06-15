package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/modulops/code-service/internal/client"
	"github.com/modulops/code-service/internal/models"
)

// GitOpsService fournit les opérations de gestion du dépôt GitOps Forgejo d'un projet :
// le dépôt dans lequel sont commités les manifests ArgoCD avant qu'ArgoCD ne les "pull".
type GitOpsService struct {
	*CodeService
}

// NewGitOpsService crée un GitOpsService réutilisant la configuration Forgejo du CodeService.
func NewGitOpsService(cs *CodeService) *GitOpsService {
	return &GitOpsService{CodeService: cs}
}

// EnsureGitOpsRepo retourne le dépôt GitOps ("owner/repo") associé au projet, le créant
// si nécessaire (nommé "<repo applicatif>-gitops", dans la même organisation/utilisateur).
// Le résultat est persisté sur le mapping projet via l'auth-service.
func (s *GitOpsService) EnsureGitOpsRepo(ctx context.Context, authToken, projectID string) (string, error) {
	mapping, err := s.getMappingWithForgejoRepo(ctx, authToken, projectID)
	if err != nil {
		return "", err
	}

	if mapping.ForgejoGitOpsRepository != "" {
		return mapping.ForgejoGitOpsRepository, nil
	}

	if mapping.ForgejoRepository == "" {
		return "", fmt.Errorf("aucun dépôt Forgejo associé au projet : impossible de déterminer l'organisation du dépôt GitOps")
	}

	owner, appRepo, err := splitRepo(mapping.ForgejoRepository)
	if err != nil {
		return "", err
	}
	gitopsRepoName := appRepo + "-gitops"
	gitopsFullName := owner + "/" + gitopsRepoName

	fj, err := s.forgejoClient(ctx)
	if err != nil {
		return "", err
	}

	exists, err := fj.RepositoryExists(owner, gitopsRepoName)
	if err != nil {
		return "", fmt.Errorf("vérification de l'existence du dépôt GitOps: %w", err)
	}
	if !exists {
		if err := fj.CreateRepository(owner, gitopsRepoName, true); err != nil {
			return "", fmt.Errorf("création du dépôt GitOps: %w", err)
		}
	}

	if err := s.setMappingGitOpsRepository(ctx, authToken, projectID, mapping.ID, gitopsFullName); err != nil {
		return "", fmt.Errorf("enregistrement du dépôt GitOps sur le mapping: %w", err)
	}

	return gitopsFullName, nil
}

// GitOpsInfo décrit le dépôt GitOps d'un projet : son URL de clone HTTPS, son nom complet
// ("owner/repo") et ses branches existantes.
type GitOpsInfo struct {
	CloneURL   string   `json:"clone_url"`
	Repository string   `json:"repository"`
	Branches   []string `json:"branches"`
}

// GetGitOpsInfo retourne les informations du dépôt GitOps d'un projet (le créant si
// nécessaire) : URL de clone, nom complet et liste des branches.
func (s *GitOpsService) GetGitOpsInfo(ctx context.Context, authToken, projectID string) (*GitOpsInfo, error) {
	gitopsRepo, err := s.EnsureGitOpsRepo(ctx, authToken, projectID)
	if err != nil {
		return nil, err
	}
	owner, repo, err := splitRepo(gitopsRepo)
	if err != nil {
		return nil, err
	}
	fj, err := s.forgejoClient(ctx)
	if err != nil {
		return nil, err
	}
	branches, err := fj.GetBranches(owner, repo)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(branches))
	for _, b := range branches {
		names = append(names, b.Name)
	}

	baseURL, err := s.cfgStore.Get(ctx, "forgejo_url")
	if err != nil {
		return nil, fmt.Errorf("lecture de l'URL Forgejo: %w", err)
	}
	if baseURL == "" {
		baseURL = "https://codeberg.org"
	}

	return &GitOpsInfo{
		CloneURL:   strings.TrimSuffix(baseURL, "/") + "/" + gitopsRepo + ".git",
		Repository: gitopsRepo,
		Branches:   names,
	}, nil
}

// CommitFiles committe un ensemble de fichiers (chemin -> contenu texte) dans le dépôt GitOps
// d'un projet, sur la branche donnée. Si createBranchFrom est non vide, la branche est créée
// au préalable à partir de cette branche source.
func (s *GitOpsService) CommitFiles(ctx context.Context, authToken, projectID, branch, createBranchFrom string, files map[string]string, message string) error {
	gitopsRepo, err := s.EnsureGitOpsRepo(ctx, authToken, projectID)
	if err != nil {
		return err
	}
	owner, repo, err := splitRepo(gitopsRepo)
	if err != nil {
		return err
	}
	fj, err := s.forgejoClient(ctx)
	if err != nil {
		return err
	}

	if createBranchFrom != "" {
		if err := fj.CreateBranch(owner, repo, branch, createBranchFrom); err != nil {
			return fmt.Errorf("création de la branche %q: %w", branch, err)
		}
	}

	for path, content := range files {
		sha, err := fj.GetFileSHA(owner, repo, path, branch)
		if err != nil {
			return fmt.Errorf("lecture du SHA de %s: %w", path, err)
		}
		req := client.CreateFileRequest{
			Content: base64.StdEncoding.EncodeToString([]byte(content)),
			Message: message,
			Branch:  branch,
			SHA:     sha,
		}
		if err := fj.CreateOrUpdateFile(owner, repo, path, req); err != nil {
			return fmt.Errorf("écriture de %s: %w", path, err)
		}
	}

	return nil
}

// getMappingWithForgejoRepo récupère le premier mapping du projet portant un dépôt Forgejo
// (et, le cas échéant, un dépôt GitOps déjà enregistré).
func (s *GitOpsService) getMappingWithForgejoRepo(ctx context.Context, authToken, projectID string) (*models.ProjectMapping, error) {
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

	for _, m := range body.Items {
		if m.ForgejoGitOpsRepository != "" {
			return &m, nil
		}
	}
	for _, m := range body.Items {
		if m.ForgejoRepository != "" {
			return &m, nil
		}
	}
	if len(body.Items) > 0 {
		return &body.Items[0], nil
	}
	return nil, fmt.Errorf("aucun mapping trouvé pour le projet %s", projectID)
}

// setMappingGitOpsRepository met à jour le champ forgejo_gitops_repository du mapping via l'auth-service.
func (s *GitOpsService) setMappingGitOpsRepository(ctx context.Context, authToken, projectID, mappingID, gitopsRepo string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/mappings/%s/gitops-repository", s.cfg.AuthServiceURL, projectID, mappingID)
	payload, err := json.Marshal(map[string]string{"forgejo_gitops_repository": gitopsRepo})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("appel auth-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth-service gitops-repository: %s", resp.Status)
	}
	return nil
}
