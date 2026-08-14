// Package client fournit des clients pour les API externes utilisées par terraform-service.
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

// Limites de sécurité pour la récupération récursive de fichiers .tf depuis un repo.
const (
	maxTFFiles  = 100
	maxTFDepth  = 5
	maxFileSize = 1 << 20 // 1 Mo par fichier
)

// TFFile représente un fichier .tf récupéré depuis un dépôt Forgejo/Codeberg (ou GitHub).
type TFFile struct {
	Path    string
	Content string
}

// decodeContent décode le contenu d'un fichier selon l'encodage indiqué par l'API
// Contents (Forgejo/Codeberg ou GitHub) — "base64" ou vide (texte brut).
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

// forgejoContentsEntry représente une entrée de la réponse Forgejo/Codeberg Contents API.
type forgejoContentsEntry struct {
	Type     string `json:"type"` // "file" ou "dir"
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// ForgejoClient client pour l'API Forgejo/Codeberg Contents (lecture seule).
type ForgejoClient struct {
	token      string
	apiBase    string
	httpClient *http.Client
}

// NewForgejoClient crée un client Forgejo/Codeberg Contents API. baseURL est
// l'URL de l'instance (ex: "https://codeberg.org" ou une instance self-hébergée).
func NewForgejoClient(baseURL, token string) *ForgejoClient {
	if baseURL == "" {
		baseURL = "https://codeberg.org"
	}
	return &ForgejoClient{
		token:      token,
		apiBase:    strings.TrimSuffix(baseURL, "/") + "/api/v1",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchTFFiles récupère récursivement tous les fichiers .tf sous le chemin donné
// d'un dépôt Forgejo/Codeberg, à la référence (branche/tag/sha) indiquée.
// Les chemins retournés sont relatifs à la racine du dépôt (ex: "terraform/main.tf").
func (c *ForgejoClient) FetchTFFiles(owner, repo, path, ref string) ([]TFFile, error) {
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner et repo requis")
	}
	var files []TFFile
	if err := c.walk(owner, repo, path, ref, 0, &files); err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("aucun fichier .tf trouvé sous %s/%s/%s@%s", owner, repo, path, ref)
	}

	return files, nil
}

// FetchFile récupère un fichier unique (quel que soit son type) à un chemin
// relatif à la racine du dépôt. Utilisé pour récupérer les fichiers référencés
// par `file("${path.module}/...")` dans les .tf (ex: scripts cloud-init) qui
// se trouvent hors du répertoire de configuration principal.
func (c *ForgejoClient) FetchFile(owner, repo, path, ref string) (TFFile, error) {
	if owner == "" || repo == "" {
		return TFFile{}, fmt.Errorf("owner et repo requis")
	}
	entries, isFile, err := c.getContents(owner, repo, path, ref)
	if err != nil {
		return TFFile{}, err
	}
	if !isFile || len(entries) != 1 {
		return TFFile{}, fmt.Errorf("chemin %s n'est pas un fichier", path)
	}
	entry := entries[0]
	if entry.Size > maxFileSize {
		return TFFile{}, fmt.Errorf("fichier %s trop volumineux (%d octets, max %d)", entry.Path, entry.Size, maxFileSize)
	}
	content := entry.Content
	encoding := entry.Encoding
	if content == "" {
		fileEntries, isFile, err := c.getContents(owner, repo, entry.Path, ref)
		if err != nil {
			return TFFile{}, err
		}
		if !isFile || len(fileEntries) != 1 {
			return TFFile{}, fmt.Errorf("impossible de récupérer le contenu de %s", entry.Path)
		}
		content = fileEntries[0].Content
		encoding = fileEntries[0].Encoding
	}
	decoded, err := decodeContent(content, encoding)
	if err != nil {
		return TFFile{}, fmt.Errorf("décodage de %s: %w", entry.Path, err)
	}
	return TFFile{Path: entry.Path, Content: decoded}, nil
}

func (c *ForgejoClient) walk(owner, repo, path, ref string, depth int, files *[]TFFile) error {
	if depth > maxTFDepth {
		return nil
	}
	entries, isFile, err := c.getContents(owner, repo, path, ref)
	if err != nil {
		return err
	}

	if isFile {
		if len(entries) != 1 {
			return fmt.Errorf("réponse Forgejo Contents inattendue pour %s", path)
		}
		return c.maybeAddFile(owner, repo, ref, entries[0], files)
	}

	for _, entry := range entries {
		if len(*files) >= maxTFFiles {
			return nil
		}
		switch entry.Type {
		case "dir":
			if err := c.walk(owner, repo, entry.Path, ref, depth+1, files); err != nil {
				return err
			}
		case "file":
			if err := c.maybeAddFile(owner, repo, ref, entry, files); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *ForgejoClient) maybeAddFile(owner, repo, ref string, entry forgejoContentsEntry, files *[]TFFile) error {
	if !strings.HasSuffix(entry.Name, ".tf") {
		return nil
	}
	if len(*files) >= maxTFFiles {
		return nil
	}
	if entry.Size > maxFileSize {
		return fmt.Errorf("fichier %s trop volumineux (%d octets, max %d)", entry.Path, entry.Size, maxFileSize)
	}

	content := entry.Content
	encoding := entry.Encoding
	if content == "" {
		// La liste de répertoire ne contient pas le contenu : refetch du fichier.
		fileEntries, isFile, err := c.getContents(owner, repo, entry.Path, ref)
		if err != nil {
			return err
		}
		if !isFile || len(fileEntries) != 1 {
			return fmt.Errorf("impossible de récupérer le contenu de %s", entry.Path)
		}
		content = fileEntries[0].Content
		encoding = fileEntries[0].Encoding
	}

	decoded, err := decodeContent(content, encoding)
	if err != nil {
		return fmt.Errorf("décodage de %s: %w", entry.Path, err)
	}

	*files = append(*files, TFFile{Path: entry.Path, Content: decoded})
	return nil
}

// getContents appelle GET /repos/{owner}/{repo}/contents/{path}?ref={ref}.
// Retourne (entries, isFile, err) — isFile=true si la réponse est un objet unique (fichier).
func (c *ForgejoClient) getContents(owner, repo, path, ref string) ([]forgejoContentsEntry, bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiBase, owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("API Forgejo Contents (%s): %s", path, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 Mo max par réponse
	if err != nil {
		return nil, false, err
	}

	// La réponse est soit un tableau (répertoire), soit un objet (fichier).
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
