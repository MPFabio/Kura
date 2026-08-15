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

// SharedNamespace est le namespace de configuration commun aux modules qui
// parlent au même Forgejo/Codeberg (identifiants forgejo_url et forgejo_token).
//
// Sans lui, chaque module lisait ces clés dans son propre namespace
// ("pipeline", "terraform", "code") : l'utilisateur devait saisir le même
// jeton jusqu'à trois fois, et l'oubli se manifestait par un « 401 token is
// required » sur un jeton pourtant valide, ce qui n'oriente pas vers la cause.
const SharedNamespace = "forgejo"

// shared retourne un client sur le namespace partagé, en réutilisant le même
// transport HTTP.
func (c *Client) shared() *Client {
	return &Client{
		authServiceURL: c.authServiceURL,
		service:        SharedNamespace,
		httpClient:     c.httpClient,
	}
}

// GetShared lit une clé dans le namespace partagé, puis retombe sur le
// namespace du service appelant.
//
// Le repli préserve les installations où le jeton n'a été saisi que dans un
// module : elles continuent de fonctionner sans ressaisie.
func (c *Client) GetShared(ctx context.Context, key string) (string, error) {
	if value, err := c.shared().Get(ctx, key); err == nil && value != "" {
		return value, nil
	}
	return c.Get(ctx, key)
}
