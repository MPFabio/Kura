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

	"github.com/modulops/terraform-service/internal/cache"
	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/handler"
	"github.com/modulops/terraform-service/internal/service"
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

	// Initialiser le service métier
	terraformService := service.NewTerraformService(redisClient, cfg)

	// Initialiser les handlers HTTP
	terraformHandler := handler.NewTerraformHandler(terraformService, cfg)

	// Configurer le routeur HTTP
	router := setupRouter(terraformHandler, cfg)

	// Créer le serveur HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Démarrer le serveur dans une goroutine
	go func() {
		log.Printf("Service Terraform démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
		}
	}()

	// Attendre un signal d'interruption
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt du service Terraform...")

	// Arrêt gracieux avec timeout de 5 secondes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur: %v", err)
	}

	log.Println("Service Terraform arrêté")
}

func setupRouter(terraformHandler *handler.TerraformHandler, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware CORS simple
	router.Use(corsMiddleware())

	// Route de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "terraform-service"})
	})

	v1 := router.Group("/api/v1")
	{
		terraformGroup := v1.Group("/terraform")
		{
			// Gestion des états Terraform
			terraformGroup.POST("/states/upload", terraformHandler.UploadStateFile)
			terraformGroup.POST("/states", terraformHandler.UploadStateFileJSON)
			terraformGroup.GET("/states", terraformHandler.ListStateFiles)
			terraformGroup.GET("/states/:id", terraformHandler.GetStateFile)
			terraformGroup.GET("/states/:id/summary", terraformHandler.GetStateSummary)
			terraformGroup.DELETE("/states/:id", terraformHandler.DeleteStateFile)

			// Ressources
			terraformGroup.GET("/states/:id/resources", terraformHandler.GetResources)
			terraformGroup.GET("/states/:id/resources/:address", terraformHandler.GetResourceByAddress)

			// Sorties
			terraformGroup.GET("/states/:id/outputs", terraformHandler.GetOutputs)

			// Détection de drift
			terraformGroup.POST("/states/:id/drift", terraformHandler.DetectDrift)
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
