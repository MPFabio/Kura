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
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/modulops/k8s-service/internal/authz"
	"github.com/modulops/k8s-service/internal/cache"
	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/handler"
	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/service"
	"github.com/modulops/k8s-service/internal/tracing"
)

func main() {

	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur lors du chargement de la configuration: %v", err)
	}

	// Initialiser le tracing OpenTelemetry (export vers Tempo)
	shutdownTracing, err := tracing.Init(context.Background(), "k8s-service", cfg.OTLPEndpoint)
	if err != nil {
		log.Printf("⚠️  Tracing OpenTelemetry désactivé (%v)", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTracing(ctx)
		}()
	}

	// Initialiser le cache Valkey
	valkeyClient, err := cache.NewValkeyClient(cfg)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de Valkey: %v", err)
	}
	defer valkeyClient.Close()

	// Note: La migration des clusters existants vers un projet par défaut
	// doit être effectuée manuellement ou via l'API, pas automatiquement au démarrage

	// Initialiser le service de gestion de clusters
	clusterService := service.NewClusterService(valkeyClient, cfg)

	// Initialiser le client Kubernetes (optionnel au démarrage : les clusters
	// sont ajoutés via l'API et le client est alors créé dynamiquement).
	var k8sClient *k8s.Client
	if cfg.KubeconfigPath != "" {
		if k8sClient, err = k8s.NewClient(cfg); err != nil {
			k8sClient = nil
		}
	}
	if k8sClient == nil {
		log.Printf("⚠️  Aucun cluster Kubernetes configuré au démarrage")
		log.Printf("💡 Vous pouvez ajouter des clusters via l'API /api/v1/k8s/clusters ou via l'interface frontend")
	}

	// Service métier : nil tant qu'aucun cluster n'est configuré.
	var k8sService *service.K8sService
	if k8sClient != nil {
		var k8sClientInterface service.K8sClient = k8sClient
		k8sService = service.NewK8sService(k8sClientInterface, valkeyClient, cfg)
	}

	// Initialiser les handlers HTTP
	k8sHandler := handler.NewK8sHandler(k8sService, clusterService, valkeyClient, cfg)
	terminalHandler := handler.NewTerminalHandler(k8sService, clusterService, valkeyClient, cfg)
	clusterHandler := handler.NewClusterHandler(clusterService, cfg)
	clusterHandler.SetInvalidators(k8sHandler, terminalHandler)
	argocdService := service.NewArgoCDService(valkeyClient, clusterService, cfg)
	helmCatalogService := service.NewHelmCatalogService(valkeyClient)
	argocdHandler := handler.NewArgoCDHandler(argocdService, helmCatalogService)
	registryService := service.NewRegistryService(valkeyClient, clusterService, cfg)
	registryHandler := handler.NewRegistryHandler(registryService)
	observabilityService := service.NewObservabilityService(clusterService)
	observabilityHandler := handler.NewObservabilityHandler(observabilityService)
	discoveryService := service.NewDiscoveryService(clusterService, argocdService)
	discoveryHandler := handler.NewDiscoveryHandler(discoveryService)

	// Configurer le routeur HTTP
	router := setupRouter(k8sHandler, terminalHandler, clusterHandler, argocdHandler, registryHandler, observabilityHandler, discoveryHandler, k8sService, clusterService, valkeyClient, cfg)

	// Créer le serveur HTTP
	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
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

func setupRouter(k8sHandler *handler.K8sHandler, terminalHandler *handler.TerminalHandler, clusterHandler *handler.ClusterHandler, argocdHandler *handler.ArgoCDHandler, registryHandler *handler.RegistryHandler, observabilityHandler *handler.ObservabilityHandler, discoveryHandler *handler.DiscoveryHandler, k8sService *service.K8sService, clusterService *service.ClusterService, valkeyClient service.Cache, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware de tracing OpenTelemetry
	router.Use(otelgin.Middleware("k8s-service", otelgin.WithFilter(tracing.SkipHealthAndMetrics)))

	// Middleware CORS simple
	router.Use(corsMiddleware())

	// Route de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "k8s-service"})
	})

	// Métriques Prometheus (cible de scraping vmagent)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routes internes (réseau Docker uniquement, non exposées via Kong) :
	// utilisées par d'autres services internes (ex: Semaphore via les playbooks Ansible).
	internal := router.Group("/internal")
	{
		internal.GET("/k8s/clusters/:id/kubeconfig", clusterHandler.GetClusterKubeconfig)
	}

	v1 := router.Group("/api/v1")
	{
		// Routes hors session utilisateur : webhook (émis par des systèmes externes)
		// et terminal WebSocket (le navigateur ne peut pas poser de header Authorization).
		open := v1.Group("/k8s")
		{
			open.POST("/webhooks/events", k8sHandler.ReceiveEventWebhook)
			open.GET("/namespaces/:namespace/pods/:name/terminal", terminalHandler.HandleTerminal)
		}

		k8sGroup := v1.Group("/k8s")
		k8sGroup.Use(authz.Middleware(cfg.AuthServiceURL, "k8s"))
		{
			// Gestion des clusters
			k8sGroup.POST("/clusters", clusterHandler.CreateCluster)
			k8sGroup.GET("/clusters", clusterHandler.ListClusters)
			k8sGroup.GET("/clusters/active", clusterHandler.GetActiveCluster)
			k8sGroup.GET("/clusters/:id", clusterHandler.GetCluster)
			k8sGroup.PUT("/clusters/:id", clusterHandler.UpdateCluster)
			k8sGroup.DELETE("/clusters/:id", clusterHandler.DeleteCluster)
			k8sGroup.POST("/clusters/:id/activate", clusterHandler.SetActiveCluster)
			k8sGroup.GET("/clusters/:id/test", clusterHandler.TestClusterConnection)

			// Namespaces
			k8sGroup.GET("/namespaces", k8sHandler.GetNamespaces)

			// Ressources par namespace
			k8sGroup.GET("/namespaces/:namespace/pods", k8sHandler.GetPods)
			k8sGroup.GET("/namespaces/:namespace/deployments", k8sHandler.GetDeployments)
			k8sGroup.GET("/namespaces/:namespace/services", k8sHandler.GetServices)
			k8sGroup.GET("/namespaces/:namespace/configmaps", k8sHandler.GetConfigMaps)
			k8sGroup.GET("/namespaces/:namespace/secrets", k8sHandler.GetSecrets)

			// Détails et logs
			k8sGroup.GET("/namespaces/:namespace/pods/:name", k8sHandler.GetPodDetail)
			k8sGroup.GET("/namespaces/:namespace/pods/:name/logs", k8sHandler.GetPodLogs)
			k8sGroup.GET("/namespaces/:namespace/pods/:name/yaml", k8sHandler.GetPodYAML)
			k8sGroup.GET("/namespaces/:namespace/deployments/:name", k8sHandler.GetDeploymentDetail)
			k8sGroup.GET("/namespaces/:namespace/deployments/:name/yaml", k8sHandler.GetDeploymentYAML)
			k8sGroup.GET("/namespaces/:namespace/services/:name/yaml", k8sHandler.GetServiceYAML)
			k8sGroup.GET("/namespaces/:namespace/configmaps/:name/yaml", k8sHandler.GetConfigMapYAML)
			k8sGroup.GET("/namespaces/:namespace/secrets/:name/yaml", k8sHandler.GetSecretYAML)

			// Actions
			k8sGroup.PUT("/namespaces/:namespace/deployments/:name/scale", k8sHandler.ScaleDeployment)
			k8sGroup.POST("/namespaces/:namespace/deployments/:name/restart", k8sHandler.RestartDeployment)
			k8sGroup.PATCH("/namespaces/:namespace/deployments/:name/env", k8sHandler.PatchDeploymentEnv)
			k8sGroup.PATCH("/namespaces/:namespace/deployments/:name/resources", k8sHandler.PatchDeploymentResources)
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

			// Nodes (cluster-wide)
			k8sGroup.GET("/nodes", k8sHandler.GetNodes)
			k8sGroup.GET("/nodes/:name/yaml", k8sHandler.GetNodeYAML)

			// ArgoCD (GitOps CD)
			argocdGroup := k8sGroup.Group("/argocd")
			{
				argocdGroup.POST("/install", argocdHandler.InstallArgoCD)
				argocdGroup.GET("/status", argocdHandler.GetStatus)
				argocdGroup.GET("/applications", argocdHandler.ListApplications)
				argocdGroup.GET("/applications/:name", argocdHandler.GetApplication)
				argocdGroup.POST("/applications", argocdHandler.CreateApplication)
				argocdGroup.POST("/applications/:name/sync", argocdHandler.SyncApplication)
				argocdGroup.POST("/applications/:name/refresh", argocdHandler.RefreshApplication)
				argocdGroup.POST("/applications/:name/rollback", argocdHandler.RollbackApplication)
				argocdGroup.PUT("/applications/:name/values", argocdHandler.UpdateApplicationValues)
				argocdGroup.DELETE("/applications/:name", argocdHandler.DeleteApplication)
				argocdGroup.GET("/helm-catalog", argocdHandler.SearchHelmCharts)
				argocdGroup.GET("/gitops/branches", argocdHandler.GetGitOpsBranches)
			}

			// Registre OCI interne (Zot)
			registryGroup := k8sGroup.Group("/registry")
			{
				registryGroup.GET("/repositories", registryHandler.ListRepositories)
				registryGroup.GET("/repositories/*name", registryHandler.GetRepository)
			}

			// Observabilité du projet (Prometheus/Loki/Tempo déployés dans le cluster client)
			observabilityGroup := k8sGroup.Group("/observability")
			{
				observabilityGroup.GET("/overview", observabilityHandler.GetOverview)
				observabilityGroup.GET("/services", observabilityHandler.GetServices)
				observabilityGroup.GET("/logs", observabilityHandler.GetLogs)
				observabilityGroup.GET("/logs/services", observabilityHandler.GetLogServices)
				observabilityGroup.GET("/traces", observabilityHandler.SearchTraces)
				observabilityGroup.GET("/traces/:traceID", observabilityHandler.GetTrace)
				observabilityGroup.Any("/grafana/*path", observabilityHandler.ProxyGrafana)
			}

			// Auto-découverte des applications ArgoCD et composants
			// d'observabilité du cluster client
			k8sGroup.GET("/discovery", discoveryHandler.GetReport)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Project-ID")
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
