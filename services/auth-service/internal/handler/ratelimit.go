package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ipBucket suit les tentatives par IP avec un compteur à fenêtre glissante.
type ipBucket struct {
	count     int
	windowEnd time.Time
}

// RateLimiter implémente un rate limiting in-memory par IP.
// Utilisé comme défense en profondeur derrière Kong (qui rate-limite aussi).
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*ipBucket
	limit    int
	window   time.Duration
}

// NewRateLimiter crée un rate limiter : limit requêtes par window par IP.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*ipBucket),
		limit:   limit,
		window:  window,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[ip]
	if !exists || now.After(b.windowEnd) {
		rl.buckets[ip] = &ipBucket{count: 1, windowEnd: now.Add(rl.window)}
		return true
	}
	b.count++
	return b.count <= rl.limit
}

// cleanupLoop purge les entrées expirées toutes les 5 minutes pour éviter les fuites mémoire.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.buckets {
			if now.After(b.windowEnd) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware retourne un gin.HandlerFunc appliquant ce rate limiter.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Trop de tentatives. Veuillez réessayer dans quelques instants.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
