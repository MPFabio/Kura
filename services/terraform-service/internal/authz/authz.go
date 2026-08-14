// Package authz fournit un middleware Gin qui délègue la décision d'accès à
// l'auth-service : le token Bearer de la requête et le projet actif (header
// X-Project-ID envoyé par le frontend) sont vérifiés contre le scope requis
// sur le module du service (read pour GET/HEAD, write sinon).
//
// La décision est mise en cache en mémoire quelques secondes pour ne pas
// solliciter l'auth-service à chaque requête.
package authz

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const cacheTTL = 15 * time.Second

type cacheEntry struct {
	status  int
	expires time.Time
}

type checker struct {
	authURL string
	module  string
	client  *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry
}

// Middleware retourne un middleware Gin qui exige un token valide et, quand un
// projet est fourni, le scope suffisant sur le module auprès de l'auth-service.
func Middleware(authServiceURL, module string) gin.HandlerFunc {
	ck := &checker{
		authURL: authServiceURL,
		module:  module,
		client:  &http.Client{Timeout: 5 * time.Second},
		cache:   make(map[string]cacheEntry),
	}
	return ck.handle
}

func (ck *checker) handle(c *gin.Context) {
	if c.Request.Method == http.MethodOptions {
		c.Next()
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token d'authentification manquant"})
		return
	}

	projectID := c.GetHeader("X-Project-ID")
	if projectID == "" {
		projectID = c.Query("project_id")
	}

	scope := "write"
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
		scope = "read"
	}

	status := ck.authorize(authHeader, projectID, scope)
	switch status {
	case http.StatusOK:
		c.Next()
	case http.StatusUnauthorized:
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token invalide ou expiré"})
	case http.StatusForbidden:
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "droits insuffisants sur le module " + ck.module + " (scope " + scope + " requis)"})
	default:
		// Fail-closed : si l'auth-service est injoignable, on refuse.
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "service d'authentification indisponible"})
	}
}

// authorize interroge l'auth-service (avec cache) et retourne le statut HTTP de la décision.
func (ck *checker) authorize(authHeader, projectID, scope string) int {
	sum := sha256.Sum256([]byte(authHeader))
	key := hex.EncodeToString(sum[:]) + "|" + projectID + "|" + scope

	ck.mu.Lock()
	if e, ok := ck.cache[key]; ok && time.Now().Before(e.expires) {
		ck.mu.Unlock()
		return e.status
	}
	ck.mu.Unlock()

	q := url.Values{}
	if projectID != "" {
		q.Set("project_id", projectID)
		q.Set("module", ck.module)
		q.Set("scope", scope)
	}
	req, err := http.NewRequest(http.MethodGet, ck.authURL+"/api/v1/auth/authorize?"+q.Encode(), nil)
	if err != nil {
		return http.StatusServiceUnavailable
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := ck.client.Do(req)
	if err != nil {
		return http.StatusServiceUnavailable
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	if status != http.StatusOK && status != http.StatusUnauthorized && status != http.StatusForbidden {
		return http.StatusServiceUnavailable
	}

	ck.mu.Lock()
	if len(ck.cache) > 1000 {
		ck.cache = make(map[string]cacheEntry)
	}
	ck.cache[key] = cacheEntry{status: status, expires: time.Now().Add(cacheTTL)}
	ck.mu.Unlock()

	return status
}
