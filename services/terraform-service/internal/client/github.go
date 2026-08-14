// Package client fournit des clients pour les API externes utilisées par terraform-service.
package client

// Le client GitHub Contents API pour le drift "fine" est désactivé : le drift
// fine s'appuie désormais sur Forgejo/Codeberg (voir forgejo.go), seul hôte Git
// utilisé par Kura. Conservé ci-dessous pour référence en cas de besoin futur
// de réintroduire un support GitHub (ex: sélecteur de provider).
//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"strings"
// 	"time"
// )
//
// const githubAPIBase = "https://api.github.com"
//
// // contentsEntry représente une entrée de la réponse GitHub Contents API.
// type contentsEntry struct {
// 	Type     string `json:"type"` // "file" ou "dir"
// 	Name     string `json:"name"`
// 	Path     string `json:"path"`
// 	Size     int    `json:"size"`
// 	Content  string `json:"content"`
// 	Encoding string `json:"encoding"`
// }
//
// // GitHubClient client pour l'API GitHub Contents (lecture seule).
// type GitHubClient struct {
// 	token      string
// 	apiBase    string
// 	httpClient *http.Client
// }
//
// // NewGitHubClient crée un client GitHub Contents API.
// func NewGitHubClient(token string) *GitHubClient {
// 	return &GitHubClient{
// 		token:      token,
// 		apiBase:    githubAPIBase,
// 		httpClient: &http.Client{Timeout: 30 * time.Second},
// 	}
// }
//
// // FetchTFFiles récupère récursivement tous les fichiers .tf sous le chemin donné
// // d'un dépôt GitHub, à la référence (branche/tag/sha) indiquée.
// func (c *GitHubClient) FetchTFFiles(owner, repo, path, ref string) ([]TFFile, error) {
// 	if owner == "" || repo == "" {
// 		return nil, fmt.Errorf("owner et repo requis")
// 	}
// 	var files []TFFile
// 	if err := c.walk(owner, repo, path, ref, 0, &files); err != nil {
// 		return nil, err
// 	}
// 	if len(files) == 0 {
// 		return nil, fmt.Errorf("aucun fichier .tf trouvé sous %s/%s/%s@%s", owner, repo, path, ref)
// 	}
//
// 	prefix := strings.Trim(path, "/")
// 	if prefix != "" {
// 		prefix += "/"
// 		for i := range files {
// 			files[i].Path = strings.TrimPrefix(files[i].Path, prefix)
// 		}
// 	}
//
// 	return files, nil
// }
//
// func (c *GitHubClient) walk(owner, repo, path, ref string, depth int, files *[]TFFile) error {
// 	if depth > maxTFDepth {
// 		return nil
// 	}
// 	entries, isFile, err := c.getContents(owner, repo, path, ref)
// 	if err != nil {
// 		return err
// 	}
//
// 	if isFile {
// 		if len(entries) != 1 {
// 			return fmt.Errorf("réponse GitHub Contents inattendue pour %s", path)
// 		}
// 		return c.maybeAddFile(owner, repo, ref, entries[0], files)
// 	}
//
// 	for _, entry := range entries {
// 		if len(*files) >= maxTFFiles {
// 			return nil
// 		}
// 		switch entry.Type {
// 		case "dir":
// 			if err := c.walk(owner, repo, entry.Path, ref, depth+1, files); err != nil {
// 				return err
// 			}
// 		case "file":
// 			if err := c.maybeAddFile(owner, repo, ref, entry, files); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
//
// func (c *GitHubClient) maybeAddFile(owner, repo, ref string, entry contentsEntry, files *[]TFFile) error {
// 	if !strings.HasSuffix(entry.Name, ".tf") {
// 		return nil
// 	}
// 	if len(*files) >= maxTFFiles {
// 		return nil
// 	}
// 	if entry.Size > maxFileSize {
// 		return fmt.Errorf("fichier %s trop volumineux (%d octets, max %d)", entry.Path, entry.Size, maxFileSize)
// 	}
//
// 	content := entry.Content
// 	encoding := entry.Encoding
// 	if content == "" {
// 		fileEntries, isFile, err := c.getContents(owner, repo, entry.Path, ref)
// 		if err != nil {
// 			return err
// 		}
// 		if !isFile || len(fileEntries) != 1 {
// 			return fmt.Errorf("impossible de récupérer le contenu de %s", entry.Path)
// 		}
// 		content = fileEntries[0].Content
// 		encoding = fileEntries[0].Encoding
// 	}
//
// 	decoded, err := decodeContent(content, encoding)
// 	if err != nil {
// 		return fmt.Errorf("décodage de %s: %w", entry.Path, err)
// 	}
//
// 	*files = append(*files, TFFile{Path: entry.Path, Content: decoded})
// 	return nil
// }
//
// // getContents appelle GET /repos/{owner}/{repo}/contents/{path}?ref={ref}.
// func (c *GitHubClient) getContents(owner, repo, path, ref string) ([]contentsEntry, bool, error) {
// 	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiBase, owner, repo, path)
// 	if ref != "" {
// 		url += "?ref=" + ref
// 	}
//
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	req.Header.Set("Accept", "application/vnd.github+json")
// 	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
// 	if c.token != "" {
// 		req.Header.Set("Authorization", "Bearer "+c.token)
// 	}
//
// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, false, fmt.Errorf("API GitHub Contents (%s): %s", path, resp.Status)
// 	}
//
// 	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
// 	if err != nil {
// 		return nil, false, err
// 	}
//
// 	trimmed := strings.TrimSpace(string(body))
// 	if strings.HasPrefix(trimmed, "[") {
// 		var entries []contentsEntry
// 		if err := json.Unmarshal(body, &entries); err != nil {
// 			return nil, false, err
// 		}
// 		return entries, false, nil
// 	}
//
// 	var entry contentsEntry
// 	if err := json.Unmarshal(body, &entry); err != nil {
// 		return nil, false, err
// 	}
// 	return []contentsEntry{entry}, true, nil
// }
