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
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/modulops/code-service/internal/config"
	"github.com/modulops/code-service/internal/handler"
	"github.com/modulops/code-service/internal/service"
	"github.com/modulops/code-service/internal/tracing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur configuration: %v", err)
	}

	// Initialiser le tracing OpenTelemetry (export vers Tempo)
	shutdownTracing, err := tracing.Init(context.Background(), "code-service", cfg.OTLPEndpoint)
	if err != nil {
		log.Printf("⚠️  Tracing OpenTelemetry désactivé (%v)", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTracing(ctx)
		}()
	}

	svc := service.New(cfg)
	h := handler.New(svc)

	router := setupRouter(h, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Code service démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt du code-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur arrêt serveur: %v", err)
	}
	log.Println("Code-service arrêté")
}

func setupRouter(h *handler.CodeHandler, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware de tracing OpenTelemetry
	router.Use(otelgin.Middleware("code-service", otelgin.WithFilter(tracing.SkipHealthAndMetrics)))

	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "code-service"})
	})

	v1 := router.Group("/api/v1")
	code := v1.Group("/code")
	{
		code.GET("/repos", h.ListRepositories)
		code.GET("/tree", h.GetTree)
		code.GET("/file", h.GetFile)
		code.GET("/commits", h.GetCommits)
		code.GET("/commits/:sha", h.GetCommitDiff)
		code.GET("/projects/:projectID/gitops/branches", h.GetGitOpsBranches)
		code.POST("/projects/:projectID/gitops/commit", h.CommitGitOpsFiles)
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
