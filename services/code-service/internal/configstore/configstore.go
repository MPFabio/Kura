// Package configstore fournit un client pour lire les configurations
// de service stockées dans Postgres via l'auth-service.
package configstore

import (
	"context"
	"encoding/json"
	"fmt"
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
// service        : nom du service propriétaire de la config (ex: "terraform")
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
