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

	"github.com/modulops/metrics-service/internal/config"
	"github.com/modulops/metrics-service/internal/handler"
	"github.com/modulops/metrics-service/internal/service"
	"github.com/modulops/metrics-service/internal/tracing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur configuration: %v", err)
	}

	// Initialiser le tracing OpenTelemetry (export vers Tempo)
	shutdownTracing, err := tracing.Init(context.Background(), "metrics-service", cfg.OTLPEndpoint)
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
	}

	svc := service.New(cfg, rdb)
	h := handler.New(svc)

	router := setupRouter(h, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Metrics service démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt du metrics-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur arrêt serveur: %v", err)
	}
	log.Println("Metrics-service arrêté")
}

func setupRouter(h *handler.MetricsHandler, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware de tracing OpenTelemetry
	router.Use(otelgin.Middleware("metrics-service", otelgin.WithFilter(tracing.SkipHealthAndMetrics)))

	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "metrics-service"})
	})

	v1 := router.Group("/api/v1")
	metrics := v1.Group("/metrics")
	{
		// platform-config est toujours accessible : il indique au frontend si
		// le reste de l'observabilité interne (ci-dessous) doit être affiché.
		metrics.GET("/platform-config", h.GetPlatformConfig)

		internal := metrics.Group("")
		internal.Use(h.RequireInternalObservability())
		{
			internal.GET("/health", h.GetHealth)
			internal.GET("/services", h.GetServices)
			internal.GET("/overview", h.GetOverview)
			internal.GET("/logs", h.GetLogs)
			internal.GET("/logs/services", h.GetLogServices)
			internal.GET("/traces", h.SearchTraces)
			internal.GET("/traces/:traceID", h.GetTrace)
			internal.GET("/config", h.GetConfig)
			internal.POST("/config", h.SetConfig)
		}
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
