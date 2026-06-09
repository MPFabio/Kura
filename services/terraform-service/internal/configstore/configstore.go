// Package configstore fournit un client pour lire et écrire les configurations
// de service dans Postgres via l'auth-service.
//
// Usage:
//
//	cs := configstore.New("http://auth-service:8080", "pipeline")
//	val, _ := cs.Get(ctx, "github_token")
//	_ = cs.SetMany(ctx, map[string]string{"github_token": "ghp_..."})
package configstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client est un client vers l'API interne de config de l'auth-service.
type Client struct {
	authServiceURL string
	service        string
	httpClient     *http.Client
}

// New crée un nouveau client configstore.
// authServiceURL : URL de base de l'auth-service (ex: "http://auth-service:8080")
// service        : nom du service appelant (ex: "pipeline", "metrics", "vault")
func New(authServiceURL, service string) *Client {
	return &Client{
		authServiceURL: authServiceURL,
		service:        service,
		httpClient:     &http.Client{Timeout: 5 * time.Second},
	}
}

// Get retourne la valeur d'une clé. Retourne "" si la clé n'existe pas.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	url := fmt.Sprintf("%s/internal/config/%s/%s", c.authServiceURL, c.service, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("configstore: GET %s → %d", url, resp.StatusCode)
	}
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Value, nil
}

// GetAll retourne toutes les clés du service.
func (c *Client) GetAll(ctx context.Context) (map[string]string, error) {
	url := fmt.Sprintf("%s/internal/config/%s", c.authServiceURL, c.service)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("configstore: GET all %s → %d: %s", c.service, resp.StatusCode, body)
	}
	var result struct {
		Config map[string]string `json:"config"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Config == nil {
		result.Config = make(map[string]string)
	}
	return result.Config, nil
}

// SetMany insère ou met à jour plusieurs clés en une seule requête.
func (c *Client) SetMany(ctx context.Context, kv map[string]string) error {
	url := fmt.Sprintf("%s/internal/config/%s", c.authServiceURL, c.service)
	body, err := json.Marshal(kv)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("configstore: POST %s → %d: %s", c.service, resp.StatusCode, b)
	}
	return nil
}

// Set insère ou met à jour une seule clé.
func (c *Client) Set(ctx context.Context, key, value string) error {
	return c.SetMany(ctx, map[string]string{key: value})
}

// GetOrFallback retourne la valeur Postgres si elle existe et est non vide,
// sinon retourne la valeur de fallback (typiquement depuis une env var).
func (c *Client) GetOrFallback(ctx context.Context, key, fallback string) string {
	val, err := c.Get(ctx, key)
	if err != nil || val == "" {
		return fallback
	}
	return val
}
