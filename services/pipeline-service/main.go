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
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/modulops/pipeline-service/internal/cache"
	"github.com/modulops/pipeline-service/internal/config"
	"github.com/modulops/pipeline-service/internal/handler"
	"github.com/modulops/pipeline-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur lors du chargement de la configuration: %v", err)
	}

	redisClient, err := cache.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de Redis: %v", err)
	}
	defer redisClient.Close()

	// Contexte racine annulé à l'arrêt du service : propagé à toutes les goroutines.
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	pipelineService := service.NewPipelineService(redisClient, cfg)
	pipelineHandler := handler.NewPipelineHandler(pipelineService, cfg)
	webhookHandler := handler.NewWebhookHandler(pipelineService, cfg)

	router := setupRouter(pipelineHandler, webhookHandler, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Service Pipeline démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
		}
	}()

	// Sync périodique GitHub respectant le contexte d'arrêt.
	syncDone := make(chan struct{})
	go func() {
		defer close(syncDone)
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		if _, err := pipelineService.SyncFromGitHub(rootCtx); err != nil {
			log.Printf("Sync GitHub initial: %v", err)
		}
		for {
			select {
			case <-rootCtx.Done():
				return
			case <-ticker.C:
				if _, err := pipelineService.SyncFromGitHub(rootCtx); err != nil {
					log.Printf("Sync GitHub: %v", err)
				}
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt du service Pipeline...")

	// Annuler le contexte racine pour stopper les goroutines de sync.
	rootCancel()
	<-syncDone

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur: %v", err)
	}

	log.Println("Service Pipeline arrêté")
}

func setupRouter(pipelineHandler *handler.PipelineHandler, webhookHandler *handler.WebhookHandler, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "pipeline-service"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")
	{
		pipelineGroup := v1.Group("/pipeline")
		{
			// Config (lier repo depuis l'UI)
			pipelineGroup.GET("/config", pipelineHandler.GetConfig)
			pipelineGroup.POST("/config", pipelineHandler.SetConfig)

			// API REST
			pipelineGroup.GET("/runs", pipelineHandler.ListRuns)
			pipelineGroup.GET("/runs/:id", pipelineHandler.GetRun)
			pipelineGroup.GET("/aggregated", pipelineHandler.GetAggregatedStatus)
			pipelineGroup.GET("/providers", pipelineHandler.ListProviders)

			// Sync manuel (API GitHub - token + GITHUB_REPOS)
			pipelineGroup.POST("/sync", pipelineHandler.SyncGitHub)

			// Webhooks
			pipelineGroup.POST("/webhooks", webhookHandler.HandleGeneric)
			pipelineGroup.POST("/webhooks/github", webhookHandler.HandleGitHub)
			pipelineGroup.POST("/webhooks/gitlab", webhookHandler.HandleGitLab)
			pipelineGroup.POST("/webhooks/jenkins", webhookHandler.HandleJenkins)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
