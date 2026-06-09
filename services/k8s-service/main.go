package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	// #region agent log
	func() {
		f, _ := os.OpenFile("/tmp/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"main.go:21","message":"main() entry","data":{},"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
		}
	}()
	// #endregion

	// Charger la configuration
	cfg, err := config.Load()
	// #region agent log
	func() {
		f, _ := os.OpenFile("/tmp/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			cfgData := map[string]interface{}{"KubeconfigPath": cfg.KubeconfigPath, "InCluster": cfg.InCluster, "ServerPort": cfg.ServerPort}
			if err != nil {
				cfgData["error"] = err.Error()
			}
			f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"B","location":"main.go:24","message":"config.Load() result","data":` + func() string { b, _ := json.Marshal(cfgData); return string(b) }() + `,"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
		}
	}()
	// #endregion
	if err != nil {
		log.Fatalf("Erreur lors du chargement de la configuration: %v", err)
	}

	// Initialiser le cache Redis
	redisClient, err := cache.NewRedisClient(cfg)
	// #region agent log
	func() {
		f, _ := os.OpenFile("/tmp/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			redisData := map[string]interface{}{}
			if err != nil {
				redisData["error"] = err.Error()
			} else {
				redisData["status"] = "success"
			}
			f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"C","location":"main.go:30","message":"Redis client init result","data":` + func() string { b, _ := json.Marshal(redisData); return string(b) }() + `,"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
		}
	}()
	// #endregion
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de Redis: %v", err)
	}
	defer redisClient.Close()

	// Note: La migration des clusters existants vers un projet par défaut
	// doit être effectuée manuellement ou via l'API, pas automatiquement au démarrage

	// Initialiser le service de gestion de clusters
	clusterService := service.NewClusterService(redisClient, cfg)

	// Initialiser le client Kubernetes (optionnel au démarrage)
	// Le service peut démarrer sans cluster, les clusters seront ajoutés via l'API
	var k8sClient *k8s.Client
	// Ne pas initialiser le client Kubernetes au démarrage si KUBECONFIG_PATH est vide
	// Les clusters seront ajoutés via l'API et le client sera initialisé dynamiquement
	// #region agent log
	func() {
		f, _ := os.OpenFile("/tmp/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"main.go:43","message":"Before k8s client init check","data":{"KubeconfigPath":"` + cfg.KubeconfigPath + `","InCluster":` + fmt.Sprintf("%v", cfg.InCluster) + `},"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
		}
	}()
	// #endregion
	if cfg.KubeconfigPath != "" {
		// #region agent log
		func() {
			f, _ := os.OpenFile("c:\\Users\\fabio\\Documents\\Kura\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				defer f.Close()
				f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"main.go:44","message":"Calling k8s.NewClient (KubeconfigPath not empty)","data":{"KubeconfigPath":"` + cfg.KubeconfigPath + `"},"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
			}
		}()
		// #endregion
		k8sClient, err = k8s.NewClient(cfg)
		// #region agent log
		func() {
			f, _ := os.OpenFile("c:\\Users\\fabio\\Documents\\Kura\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				defer f.Close()
				k8sData := map[string]interface{}{}
				if err != nil {
					k8sData["error"] = err.Error()
				} else {
					k8sData["status"] = "success"
				}
				f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"main.go:45","message":"k8s.NewClient result","data":` + func() string { b, _ := json.Marshal(k8sData); return string(b) }() + `,"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
			}
		}()
		// #endregion
		if err != nil {
			log.Printf("⚠️  Aucun cluster Kubernetes configuré au démarrage")
			log.Printf("💡 Vous pouvez ajouter des clusters via l'API /api/v1/k8s/clusters")
			log.Printf("   ou via l'interface frontend")
			k8sClient = nil
		}
	} else {
		// #region agent log
		func() {
			f, _ := os.OpenFile("c:\\Users\\fabio\\Documents\\Kura\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				defer f.Close()
				f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"main.go:52","message":"Skipping k8s.NewClient (KubeconfigPath empty)","data":{},"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
			}
		}()
		// #endregion
		log.Printf("⚠️  Aucun cluster Kubernetes configuré au démarrage")
		log.Printf("💡 Vous pouvez ajouter des clusters via l'API /api/v1/k8s/clusters")
		log.Printf("   ou via l'interface frontend")
		k8sClient = nil
	}

	// Initialiser le service métier (peut être nil si pas de cluster)
	var k8sService *service.K8sService
	if k8sClient != nil {
		// Convertir *k8s.Client en service.K8sClient (interface)
		var k8sClientInterface service.K8sClient = k8sClient
		k8sService = service.NewK8sService(k8sClientInterface, redisClient, cfg)
	} else {
		// Créer un service vide qui sera initialisé quand un cluster sera ajouté
		k8sService = nil
	}

	// Initialiser les handlers HTTP
	k8sHandler := handler.NewK8sHandler(k8sService, clusterService, redisClient, cfg)
	clusterHandler := handler.NewClusterHandler(clusterService, cfg)

	// Configurer le routeur HTTP
	router := setupRouter(k8sHandler, clusterHandler, k8sService, clusterService, redisClient, cfg)

	// Créer le serveur HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Démarrer le serveur dans une goroutine
	go func() {
		// #region agent log
		func() {
			f, _ := os.OpenFile("c:\\Users\\fabio\\Documents\\Kura\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				defer f.Close()
				f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"main.go:82","message":"Starting HTTP server","data":{"port":"` + cfg.ServerPort + `"},"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
			}
		}()
		// #endregion
		log.Printf("Service Kubernetes démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// #region agent log
			func() {
				f, _ := os.OpenFile("c:\\Users\\fabio\\Documents\\Kura\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if f != nil {
					defer f.Close()
					f.WriteString(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"main.go:84","message":"HTTP server error","data":{"error":"` + err.Error() + `"},"timestamp":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + "}\n")
				}
			}()
			// #endregion
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

func setupRouter(k8sHandler *handler.K8sHandler, clusterHandler *handler.ClusterHandler, k8sService *service.K8sService, clusterService *service.ClusterService, redisClient service.Cache, cfg *config.Config) *gin.Engine {
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

			// Terminal (WebSocket) - toujours disponible, créé dynamiquement si nécessaire
			terminalHandler := handler.NewTerminalHandler(k8sService, clusterService, redisClient, cfg)
			k8sGroup.GET("/namespaces/:namespace/pods/:name/terminal", terminalHandler.HandleTerminal)

			// Nodes (cluster-wide)
			k8sGroup.GET("/nodes", k8sHandler.GetNodes)
			k8sGroup.GET("/nodes/:name/yaml", k8sHandler.GetNodeYAML)

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
