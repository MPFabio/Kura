// Package client fournit un client pour l'API GitHub Contents/Commits
// (lecture seule), utilisé pour naviguer dans le code source d'un dépôt.
package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const githubAPIBase = "https://api.github.com"

// maxFileSize limite la taille des fichiers dont le contenu est récupéré.
const maxFileSize = 1 << 20 // 1 Mo

// TreeEntry représente une entrée (fichier ou dossier) d'une arborescence.
type TreeEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // "file" ou "dir"
	Size int    `json:"size,omitempty"`
}

// FileContent représente le contenu décodé d'un fichier.
type FileContent struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Size      int    `json:"size"`
	Truncated bool   `json:"truncated,omitempty"`
}

// Commit représente un commit dans l'historique d'un dépôt.
type Commit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	URL     string `json:"url"`
}

// CommitFile représente un fichier modifié par un commit.
type CommitFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch,omitempty"`
}

// CommitDetail représente le détail d'un commit, avec les fichiers modifiés.
type CommitDetail struct {
	Commit
	Files []CommitFile `json:"files"`
}

// contentsEntry représente une entrée de la réponse GitHub Contents API.
type contentsEntry struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// commitListEntry représente une entrée de GET /repos/{owner}/{repo}/commits.
type commitListEntry struct {
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

// commitDetailEntry représente la réponse de GET /repos/{owner}/{repo}/commits/{sha}.
type commitDetailEntry struct {
	commitListEntry
	Files []struct {
		Filename  string `json:"filename"`
		Status    string `json:"status"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
		Patch     string `json:"patch"`
	} `json:"files"`
}

// GitHubClient client pour les API GitHub Contents et Commits (lecture seule).
type GitHubClient struct {
	token      string
	apiBase    string
	httpClient *http.Client
}

// NewGitHubClient crée un client GitHub.
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		apiBase:    githubAPIBase,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTree liste le contenu d'un répertoire (un seul niveau, sans récursion).
func (c *GitHubClient) GetTree(owner, repo, path, ref string) ([]TreeEntry, error) {
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
func (c *GitHubClient) GetFileContent(owner, repo, path, ref string) (*FileContent, error) {
	entries, isFile, err := c.getContents(owner, repo, path, ref)
	if err != nil {
		return nil, err
	}
	if !isFile || len(entries) != 1 {
		return nil, fmt.Errorf("%s n'est pas un fichier", path)
	}
	entry := entries[0]

	if entry.Size > maxFileSize {
		return &FileContent{Path: entry.Path, Size: entry.Size, Truncated: true}, nil
	}

	content, err := decodeContent(entry.Content, entry.Encoding)
	if err != nil {
		return nil, fmt.Errorf("décodage de %s: %w", entry.Path, err)
	}

	return &FileContent{Path: entry.Path, Content: content, Size: entry.Size}, nil
}

// GetCommits récupère l'historique des commits d'un dépôt (ou d'un chemin précis).
func (c *GitHubClient) GetCommits(owner, repo, path, ref string, page int) ([]Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?per_page=20", c.apiBase, owner, repo)
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

	var entries []commitListEntry
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
func (c *GitHubClient) GetCommitDiff(owner, repo, sha string) (*CommitDetail, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s", c.apiBase, owner, repo, sha)

	body, err := c.do(url)
	if err != nil {
		return nil, err
	}

	var entry commitDetailEntry
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

func decodeContent(content, encoding string) (string, error) {
	if encoding != "base64" {
		return content, nil
	}
	cleaned := strings.ReplaceAll(content, "\n", "")
	raw, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// getContents appelle GET /repos/{owner}/{repo}/contents/{path}?ref={ref}.
// Retourne (entries, isFile, err) — isFile=true si la réponse est un objet unique (fichier).
func (c *GitHubClient) getContents(owner, repo, path, ref string) ([]contentsEntry, bool, error) {
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
		var entries []contentsEntry
		if err := json.Unmarshal(body, &entries); err != nil {
			return nil, false, err
		}
		return entries, false, nil
	}

	var entry contentsEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, false, err
	}
	return []contentsEntry{entry}, true, nil
}

// do exécute une requête GET authentifiée vers l'API GitHub et retourne le corps de la réponse.
func (c *GitHubClient) do(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API GitHub (%s): %s", url, resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 Mo max par réponse
}
