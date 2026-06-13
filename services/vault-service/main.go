package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/modulops/vault-service/internal/config"
	"github.com/modulops/vault-service/internal/handler"
	"github.com/modulops/vault-service/internal/service"
	"github.com/modulops/vault-service/internal/tracing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur configuration: %v", err)
	}

	// Initialiser le tracing OpenTelemetry (export vers Tempo)
	shutdownTracing, err := tracing.Init(context.Background(), "vault-service", cfg.OTLPEndpoint)
	if err != nil {
		log.Printf("⚠️  Tracing OpenTelemetry désactivé (%v)", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTracing(ctx)
		}()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("⚠️  Redis non disponible (%v) — cache désactivé", err)
		rdb = nil
	}

	svc, err := service.New(cfg, rdb)
	if err != nil {
		log.Fatalf("Erreur initialisation vault-service: %v", err)
	}

	h := handler.New(svc)
	router := setupRouter(h, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Vault service démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt du vault-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur arrêt serveur: %v", err)
	}
	log.Println("Vault-service arrêté")
}

func setupRouter(h *handler.VaultHandler, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware de tracing OpenTelemetry
	router.Use(otelgin.Middleware("vault-service", otelgin.WithFilter(tracing.SkipHealthAndMetrics)))

	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "vault-service"})
	})

	v1 := router.Group("/api/v1/vault")
	{
		v1.GET("/status", h.GetStatus)
		v1.GET("/config", h.GetConfig)
		v1.POST("/config", h.SetConfig)
		v1.GET("/secrets", h.ListSecrets)
		v1.GET("/secrets/*path", h.GetSecret)
		v1.POST("/secrets/*path", h.WriteSecret)
		v1.DELETE("/secrets/*path", h.DeleteSecret)
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
