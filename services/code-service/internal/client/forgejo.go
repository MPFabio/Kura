// Package client fournit un client pour l'API Forgejo/Codeberg Contents/Commits
// (lecture seule), utilisé pour naviguer dans le code source d'un dépôt.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// maxFileSizeForgejo limite la taille des fichiers dont le contenu est récupéré.
const maxFileSizeForgejo = 1 << 20 // 1 Mo

// forgejoContentsEntry représente une entrée de la réponse Forgejo Contents API.
type forgejoContentsEntry struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// Branch représente une branche d'un dépôt Forgejo.
type Branch struct {
	Name string `json:"name"`
}

// CreateFileRequest représente le corps d'une requête de création/mise à jour
// de fichier via l'API Contents de Forgejo.
type CreateFileRequest struct {
	Content string `json:"content"` // base64
	Message string `json:"message"`
	Branch  string `json:"branch"`
	SHA     string `json:"sha,omitempty"` // requis pour une mise à jour
}

// forgejoCommitListEntry représente une entrée de GET /repos/{owner}/{repo}/commits.
type forgejoCommitListEntry struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	HTMLURL string `json:"html_url"`
}

// forgejoCommitDetailEntry représente la réponse de GET /repos/{owner}/{repo}/git/commits/{sha}.
type forgejoCommitDetailEntry struct {
	forgejoCommitListEntry
	Files []struct {
		Filename  string `json:"filename"`
		Status    string `json:"status"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
		Patch     string `json:"patch"`
	} `json:"files"`
}

// ForgejoClient client pour les API Forgejo/Codeberg Contents et Commits (lecture seule).
type ForgejoClient struct {
	token      string
	apiBase    string
	httpClient *http.Client
}

// NewForgejoClient crée un client Forgejo/Codeberg. baseURL est l'URL de l'instance
// (ex: "https://codeberg.org" ou une instance Forgejo self-hosted).
func NewForgejoClient(baseURL, token string) *ForgejoClient {
	return &ForgejoClient{
		token:      token,
		apiBase:    strings.TrimSuffix(baseURL, "/") + "/api/v1",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTree liste le contenu d'un répertoire (un seul niveau, sans récursion).
func (c *ForgejoClient) GetTree(owner, repo, path, ref string) ([]TreeEntry, error) {
	entries, isFile, err := c.getContents(owner, repo, path, ref)
	if err != nil {
		return nil, err
	}
	if isFile {
		return nil, fmt.Errorf("%s est un fichier, pas un répertoire", path)
	}

	result := make([]TreeEntry, 0, len(entries))
	for _, e := range entries {
		if e.Type != "file" && e.Type != "dir" {
			continue
		}
		result = append(result, TreeEntry{Name: e.Name, Path: e.Path, Type: e.Type, Size: e.Size})
	}
	return result, nil
}

// GetFileContent récupère le contenu décodé d'un fichier.
func (c *ForgejoClient) GetFileContent(owner, repo, path, ref string) (*FileContent, error) {
	entries, isFile, err := c.getContents(owner, repo, path, ref)
	if err != nil {
		return nil, err
	}
	if !isFile || len(entries) != 1 {
		return nil, fmt.Errorf("%s n'est pas un fichier", path)
	}
	entry := entries[0]

	if entry.Size > maxFileSizeForgejo {
		return &FileContent{Path: entry.Path, Size: entry.Size, Truncated: true}, nil
	}

	content, err := decodeContent(entry.Content, entry.Encoding)
	if err != nil {
		return nil, fmt.Errorf("décodage de %s: %w", entry.Path, err)
	}

	return &FileContent{Path: entry.Path, Content: content, Size: entry.Size}, nil
}

// GetCommits récupère l'historique des commits d'un dépôt (ou d'un chemin précis).
func (c *ForgejoClient) GetCommits(owner, repo, path, ref string, page int) ([]Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?limit=20", c.apiBase, owner, repo)
	if path != "" {
		url += "&path=" + path
	}
	if ref != "" {
		url += "&sha=" + ref
	}
	if page > 0 {
		url += fmt.Sprintf("&page=%d", page)
	}

	body, err := c.do(url)
	if err != nil {
		return nil, err
	}

	var entries []forgejoCommitListEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	commits := make([]Commit, 0, len(entries))
	for _, e := range entries {
		commits = append(commits, Commit{
			SHA:     e.SHA,
			Message: e.Commit.Message,
			Author:  e.Commit.Author.Name,
			Date:    e.Commit.Author.Date,
			URL:     e.HTMLURL,
		})
	}
	return commits, nil
}

// GetCommitDiff récupère le détail d'un commit, avec la liste des fichiers modifiés et leurs diffs.
func (c *ForgejoClient) GetCommitDiff(owner, repo, sha string) (*CommitDetail, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/commits/%s", c.apiBase, owner, repo, sha)

	body, err := c.do(url)
	if err != nil {
		return nil, err
	}

	var entry forgejoCommitDetailEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, err
	}

	detail := &CommitDetail{
		Commit: Commit{
			SHA:     entry.SHA,
			Message: entry.Commit.Message,
			Author:  entry.Commit.Author.Name,
			Date:    entry.Commit.Author.Date,
			URL:     entry.HTMLURL,
		},
	}
	for _, f := range entry.Files {
		detail.Files = append(detail.Files, CommitFile{
			Filename:  f.Filename,
			Status:    f.Status,
			Additions: f.Additions,
			Deletions: f.Deletions,
			Patch:     f.Patch,
		})
	}
	return detail, nil
}

// getContents appelle GET /repos/{owner}/{repo}/contents/{path}?ref={ref}.
// Retourne (entries, isFile, err) — isFile=true si la réponse est un objet unique (fichier).
func (c *ForgejoClient) getContents(owner, repo, path, ref string) ([]forgejoContentsEntry, bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiBase, owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}

	body, err := c.do(url)
	if err != nil {
		return nil, false, err
	}

	trimmed := strings.TrimSpace(string(body))
	if strings.HasPrefix(trimmed, "[") {
		var entries []forgejoContentsEntry
		if err := json.Unmarshal(body, &entries); err != nil {
			return nil, false, err
		}
		return entries, false, nil
	}

	var entry forgejoContentsEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, false, err
	}
	return []forgejoContentsEntry{entry}, true, nil
}

// do exécute une requête GET authentifiée vers l'API Forgejo et retourne le corps de la réponse.
func (c *ForgejoClient) do(url string) ([]byte, error) {
	return c.doMethod(http.MethodGet, url, nil)
}

// doMethod exécute une requête authentifiée vers l'API Forgejo avec la méthode et
// le corps JSON donnés, et retourne le corps de la réponse.
func (c *ForgejoClient) doMethod(method, url string, body []byte) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("API Forgejo (%s %s): %s: %s", method, url, resp.Status, string(respBody))
	}

	return io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 Mo max par réponse
}

// GetBranches liste les branches d'un dépôt.
func (c *ForgejoClient) GetBranches(owner, repo string) ([]Branch, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/branches", c.apiBase, owner, repo)
	body, err := c.do(url)
	if err != nil {
		return nil, err
	}
	var branches []Branch
	if err := json.Unmarshal(body, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}

// CreateBranch crée une nouvelle branche à partir d'une branche existante.
func (c *ForgejoClient) CreateBranch(owner, repo, newBranch, fromBranch string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/branches", c.apiBase, owner, repo)
	payload, err := json.Marshal(map[string]string{
		"new_branch_name": newBranch,
		"old_branch_name": fromBranch,
	})
	if err != nil {
		return err
	}
	_, err = c.doMethod(http.MethodPost, url, payload)
	return err
}

// GetFileSHA retourne le SHA du blob d'un fichier existant, ou "" s'il n'existe pas.
func (c *ForgejoClient) GetFileSHA(owner, repo, path, ref string) (string, error) {
	entries, isFile, err := c.getContents(owner, repo, path, ref)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return "", nil
		}
		return "", err
	}
	if !isFile || len(entries) != 1 {
		return "", fmt.Errorf("%s n'est pas un fichier", path)
	}
	return entries[0].SHA, nil
}

// CreateOrUpdateFile crée ou met à jour un fichier via l'API Contents de Forgejo.
// Si req.SHA est vide, le fichier est créé ; sinon il est mis à jour.
func (c *ForgejoClient) CreateOrUpdateFile(owner, repo, path string, req CreateFileRequest) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiBase, owner, repo, path)
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	method := http.MethodPost
	if req.SHA != "" {
		method = http.MethodPut
	}
	_, err = c.doMethod(method, url, payload)
	return err
}

// CreateRepository crée un dépôt pour l'utilisateur authentifié ou une organisation.
func (c *ForgejoClient) CreateRepository(owner, name string, private bool) error {
	payload, err := json.Marshal(map[string]interface{}{
		"name":    name,
		"private": private,
	})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/orgs/%s/repos", c.apiBase, owner)
	if _, err := c.doMethod(http.MethodPost, url, payload); err == nil {
		return nil
	}
	// Repli sur la création d'un dépôt utilisateur si l'owner n'est pas une organisation.
	url = fmt.Sprintf("%s/user/repos", c.apiBase)
	_, err = c.doMethod(http.MethodPost, url, payload)
	return err
}

// RepositoryExists vérifie si un dépôt existe.
func (c *ForgejoClient) RepositoryExists(owner, repo string) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.apiBase, owner, repo)
	_, err := c.do(url)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
