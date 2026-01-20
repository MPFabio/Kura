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

	"github.com/modulops/k8s-service/internal/cache"
	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/handler"
	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/service"
)

func main() {
	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur lors du chargement de la configuration: %v", err)
	}

	// Initialiser le cache Redis
	redisClient, err := cache.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialiser le client Kubernetes
	k8sClient, err := k8s.NewClient(cfg)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation du client Kubernetes: %v", err)
	}

	// Initialiser le service métier
	k8sService := service.NewK8sService(k8sClient, redisClient, cfg)

	// Initialiser les handlers HTTP
	k8sHandler := handler.NewK8sHandler(k8sService, cfg)

	// Configurer le routeur HTTP
	router := setupRouter(k8sHandler, k8sService, cfg)

	// Créer le serveur HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Démarrer le serveur dans une goroutine
	go func() {
		log.Printf("Service Kubernetes démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
		}
	}()

	// Attendre un signal d'interruption
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt du service Kubernetes...")

	// Arrêt gracieux avec timeout de 5 secondes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur: %v", err)
	}

	log.Println("Service Kubernetes arrêté")
}

func setupRouter(k8sHandler *handler.K8sHandler, k8sService *service.K8sService, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware CORS simple
	router.Use(corsMiddleware())

	// Route de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "k8s-service"})
	})

	v1 := router.Group("/api/v1")
	{
		k8sGroup := v1.Group("/k8s")
		{
			// Namespaces
			k8sGroup.GET("/namespaces", k8sHandler.GetNamespaces)
			
			// Ressources par namespace
			k8sGroup.GET("/namespaces/:namespace/pods", k8sHandler.GetPods)
			k8sGroup.GET("/namespaces/:namespace/deployments", k8sHandler.GetDeployments)
			k8sGroup.GET("/namespaces/:namespace/services", k8sHandler.GetServices)
			k8sGroup.GET("/namespaces/:namespace/configmaps", k8sHandler.GetConfigMaps)
			k8sGroup.GET("/namespaces/:namespace/secrets", k8sHandler.GetSecrets)
			
			// Détails et logs
			k8sGroup.GET("/namespaces/:namespace/pods/:name/logs", k8sHandler.GetPodLogs)
			k8sGroup.GET("/namespaces/:namespace/pods/:name/yaml", k8sHandler.GetPodYAML)
			k8sGroup.GET("/namespaces/:namespace/deployments/:name/yaml", k8sHandler.GetDeploymentYAML)
			k8sGroup.GET("/namespaces/:namespace/services/:name/yaml", k8sHandler.GetServiceYAML)
			
			// Actions
			k8sGroup.PUT("/namespaces/:namespace/deployments/:name/scale", k8sHandler.ScaleDeployment)
			k8sGroup.DELETE("/namespaces/:namespace/pods/:name", k8sHandler.DeletePod)
			k8sGroup.DELETE("/namespaces/:namespace/deployments/:name", k8sHandler.DeleteDeployment)
			k8sGroup.DELETE("/namespaces/:namespace/services/:name", k8sHandler.DeleteService)
			
			// Actions en masse (Bulk Actions)
			k8sGroup.POST("/namespaces/:namespace/pods/bulk/delete", k8sHandler.BulkDeletePods)
			k8sGroup.POST("/namespaces/:namespace/pods/bulk/restart", k8sHandler.BulkRestartPods)
			k8sGroup.POST("/namespaces/:namespace/deployments/bulk/delete", k8sHandler.BulkDeleteDeployments)
			k8sGroup.POST("/namespaces/:namespace/deployments/bulk/scale", k8sHandler.BulkScaleDeployments)
			k8sGroup.POST("/namespaces/:namespace/services/bulk/delete", k8sHandler.BulkDeleteServices)
			
			// Événements
			k8sGroup.GET("/namespaces/:namespace/events", k8sHandler.GetEvents)
			
			// Terminal (WebSocket)
			terminalHandler := handler.NewTerminalHandler(k8sService)
			k8sGroup.GET("/namespaces/:namespace/pods/:name/terminal", terminalHandler.HandleTerminal)
			
			// Nodes (cluster-wide)
			k8sGroup.GET("/nodes", k8sHandler.GetNodes)

			// Webhook pour recevoir des événements Kubernetes (squelette)
			k8sGroup.POST("/webhooks/events", k8sHandler.ReceiveEventWebhook)
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

		// Support WebSocket upgrade
		if c.GetHeader("Upgrade") == "websocket" {
			c.Writer.Header().Set("Connection", "Upgrade")
			c.Writer.Header().Set("Upgrade", "websocket")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

