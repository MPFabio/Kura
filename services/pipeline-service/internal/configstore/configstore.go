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
// module : elles continuent de fonctionner sans ressaisie, et la valeur
// remonte dans le namespace partagé à la prochaine écriture.
func (c *Client) GetShared(ctx context.Context, key string) (string, error) {
	if value, err := c.shared().Get(ctx, key); err == nil && value != "" {
		return value, nil
	}
	return c.Get(ctx, key)
}

// SetShared écrit une clé dans le namespace partagé, afin qu'elle soit visible
// de tous les modules quel que soit celui depuis lequel elle a été saisie.
func (c *Client) SetShared(ctx context.Context, key, value string) error {
	return c.shared().Set(ctx, key, value)
}

// sharedForgejoKeys énumère les clés hébergées dans le namespace partagé.
var sharedForgejoKeys = map[string]bool{
	"forgejo_url":   true,
	"forgejo_token": true,
}

// GetAllShared retourne les clés du service, complétées par celles du
// namespace partagé, qui priment quand elles sont renseignées.
func (c *Client) GetAllShared(ctx context.Context) (map[string]string, error) {
	all, err := c.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if all == nil {
		all = map[string]string{}
	}
	shared, err := c.shared().GetAll(ctx)
	if err != nil {
		// Le namespace partagé peut ne pas exister sur une installation
		// antérieure : on se contente alors des clés du service.
		return all, nil
	}
	for k, v := range shared {
		if v != "" {
			all[k] = v
		}
	}
	return all, nil
}

// SetManyShared écrit les identifiants Forgejo dans le namespace partagé et
// les autres clés dans celui du service, pour qu'un jeton saisi depuis
// n'importe quel module soit visible de tous.
func (c *Client) SetManyShared(ctx context.Context, kv map[string]string) error {
	local := map[string]string{}
	shared := map[string]string{}
	for k, v := range kv {
		if sharedForgejoKeys[k] {
			shared[k] = v
		} else {
			local[k] = v
		}
	}
	if len(shared) > 0 {
		if err := c.shared().SetMany(ctx, shared); err != nil {
			return err
		}
	}
	if len(local) > 0 {
		return c.SetMany(ctx, local)
	}
	return nil
}
