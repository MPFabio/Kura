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
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/modulops/ansible-service/internal/authz"
	"github.com/modulops/ansible-service/internal/config"
	"github.com/modulops/ansible-service/internal/handler"
	"github.com/modulops/ansible-service/internal/service"
	"github.com/modulops/ansible-service/internal/tracing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur configuration: %v", err)
	}

	// Initialiser le tracing OpenTelemetry (export vers Tempo)
	shutdownTracing, err := tracing.Init(context.Background(), "ansible-service", cfg.OTLPEndpoint)
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
		Addr:     cfg.ValkeyAddr,
		Password: cfg.ValkeyPassword,
		DB:       cfg.ValkeyDB,
	})
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("⚠️  Valkey non disponible (%v) — cache désactivé", err)
	}

	svc := service.New(cfg, rdb)
	h := handler.New(svc)

	router := setupRouter(h, cfg)

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Ansible service démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt de l'ansible-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur arrêt serveur: %v", err)
	}
	log.Println("Ansible-service arrêté")
}

func setupRouter(h *handler.AnsibleHandler, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware de tracing OpenTelemetry
	router.Use(otelgin.Middleware("ansible-service", otelgin.WithFilter(tracing.SkipHealthAndMetrics)))

	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ansible-service"})
	})
	router.HEAD("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")

	ansible := v1.Group("/ansible")
	ansible.Use(authz.Middleware(cfg.AuthServiceURL, "ansible"))
	{
		// Jobs ("history" est déclaré avant ":job_id" pour ne pas être pris pour un id)
		ansible.GET("/jobs", h.GetJobs)
		ansible.GET("/jobs/history", h.GetJobHistory)
		ansible.GET("/jobs/:job_id", h.GetJob)

		// Inventaires
		ansible.GET("/inventories", h.GetInventories)
		ansible.GET("/inventories/:inventory_id", h.GetInventory)
		ansible.GET("/inventories/:inventory_id/hosts", h.GetInventoryHosts)

		// Templates de jobs
		ansible.GET("/job-templates", h.GetJobTemplates)
		ansible.GET("/job-templates/:template_id", h.GetJobTemplate)
		ansible.GET("/job-templates/:template_id/playbook", h.GetJobTemplatePlaybook)
		ansible.POST("/job-templates/:template_id/launch", h.LaunchJobTemplate)

		// Projets
		ansible.GET("/projects", h.GetProjects)
		ansible.GET("/projects/:project_id", h.GetProject)
		ansible.GET("/projects/:project_id/playbooks", h.GetProjectPlaybooks)

		// Credentials et organisations : stubs de compatibilité AWX
		ansible.GET("/credentials", h.GetCredentials)
		ansible.POST("/credentials", h.NotSupported)
		ansible.GET("/credentials/:credential_id", h.GetCredential)
		ansible.PUT("/credentials/:credential_id", h.NotSupported)
		ansible.DELETE("/credentials/:credential_id", h.NotSupported)
		ansible.GET("/organizations", h.GetOrganizations)
		ansible.POST("/organizations", h.NotSupported)
		ansible.GET("/organizations/:organization_id", h.GetOrganization)
		ansible.PUT("/organizations/:organization_id", h.NotSupported)
		ansible.DELETE("/organizations/:organization_id", h.NotSupported)

		// Analyse de playbook
		ansible.POST("/playbooks/analyze", h.AnalyzePlaybook)

		// Configuration Semaphore
		ansible.GET("/config", h.GetConfig)
		ansible.POST("/config", h.SetConfig)
	}

	// Webhooks
	v1.POST("/webhooks/ansible", h.HandleWebhook)

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Project-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
