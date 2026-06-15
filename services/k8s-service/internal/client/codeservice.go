// Package client fournit des clients HTTP vers les autres services Kura.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CodeServiceClient est un client HTTP vers code-service, utilisé pour le flux GitOps
// (commit des manifests ArgoCD dans le dépôt GitOps avant qu'ArgoCD ne les "pull").
type CodeServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCodeServiceClient crée un client vers code-service.
func NewCodeServiceClient(baseURL string) *CodeServiceClient {
	return &CodeServiceClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GitOpsInfo décrit le dépôt GitOps d'un projet : son URL de clone HTTPS, son nom complet
// ("owner/repo") et ses branches existantes.
type GitOpsInfo struct {
	CloneURL   string   `json:"clone_url"`
	Repository string   `json:"repository"`
	Branches   []string `json:"branches"`
}

// GetGitOpsInfo retourne les informations (URL de clone, nom complet, branches) du dépôt
// GitOps d'un projet, le créant si nécessaire.
func (c *CodeServiceClient) GetGitOpsInfo(ctx context.Context, authToken, projectID string) (*GitOpsInfo, error) {
	url := fmt.Sprintf("%s/api/v1/code/projects/%s/gitops/branches", c.baseURL, projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel code-service: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code-service gitops/branches: %s: %s", resp.Status, string(body))
	}

	var info GitOpsInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// CommitGitOpsFiles committe un ensemble de fichiers (chemin -> contenu) dans le dépôt GitOps
// d'un projet, sur la branche donnée.
func (c *CodeServiceClient) CommitGitOpsFiles(ctx context.Context, authToken, projectID, branch, createBranchFrom string, files map[string]string, message string) error {
	url := fmt.Sprintf("%s/api/v1/code/projects/%s/gitops/commit", c.baseURL, projectID)

	payload, err := json.Marshal(map[string]interface{}{
		"branch":              branch,
		"create_branch_from":  createBranchFrom,
		"files":               files,
		"message":             message,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("appel code-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("code-service gitops/commit: %s: %s", resp.Status, string(body))
	}
	return nil
}
