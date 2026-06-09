package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modulops/auth-service/internal/config"
	"github.com/modulops/auth-service/internal/handler"
	"github.com/modulops/auth-service/internal/repository"
	"github.com/modulops/auth-service/internal/service"

	"github.com/gin-gonic/gin"
)

// authRateLimiter protège Login et Register contre le brute-force :
// 10 tentatives par IP par minute (défense en profondeur derrière Kong).
var authRateLimiter = handler.NewRateLimiter(10, time.Minute)

func main() {
	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur lors du chargement de la configuration: %v", err)
	}

	// Initialiser le repository
	repo, err := repository.New(cfg)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation du repository: %v", err)
	}
	defer repo.Close()

	// Initialiser les services
	authService := service.NewAuthService(repo, cfg)
	projectService := service.NewProjectService(repo)

	// Initialiser les handlers
	authHandler := handler.NewAuthHandler(authService, projectService, cfg)
	projectHandler := handler.NewProjectHandler(projectService, cfg)
	configHandler := handler.NewConfigHandler(repo)

	// Configurer le routeur
	router := setupRouter(authHandler, projectHandler, configHandler, cfg)

	// Créer le serveur HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Démarrer le serveur dans une goroutine
	go func() {
		log.Printf("Service d'authentification démarré sur le port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
		}
	}()

	// Attendre un signal d'interruption
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Arrêt du service d'authentification...")

	// Arrêt gracieux avec timeout de 5 secondes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur: %v", err)
	}

	log.Println("Service d'authentification arrêté")
}

func setupRouter(authHandler *handler.AuthHandler, projectHandler *handler.ProjectHandler, configHandler *handler.ConfigHandler, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware CORS
	router.Use(corsMiddleware())

	// Routes de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "auth-service"})
	})

	// Routes d'authentification
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authRateLimiter.Middleware(), authHandler.Register)
			auth.POST("/login", authRateLimiter.Middleware(), authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", authHandler.RequireAuth(), authHandler.GetCurrentUser)
			auth.GET("/permissions", authHandler.RequireAuth(), authHandler.GetPermissions)
			auth.PUT("/me", authHandler.RequireAuth(), authHandler.UpdateCurrentUser)
			auth.PUT("/password", authHandler.RequireAuth(), authHandler.ChangePassword)
		}

		// Routes de projets
		projects := v1.Group("/projects")
		projects.Use(authHandler.RequireAuth())
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.ListProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)
			projects.POST("/:id/members", projectHandler.AddProjectMember)
			projects.GET("/:id/members", projectHandler.ListProjectMembers)
			projects.PUT("/:id/members/:user_id", projectHandler.UpdateProjectMember)
			projects.DELETE("/:id/members/:user_id", projectHandler.RemoveProjectMember)
			projects.GET("/:id/mappings", projectHandler.ListProjectMappings)
			projects.POST("/:id/mappings", projectHandler.CreateProjectMapping)
			projects.DELETE("/:id/mappings/:mapping_id", projectHandler.DeleteProjectMapping)
			projects.GET("/:id/permissions", projectHandler.ListProjectPermissions)
			projects.POST("/:id/permissions", projectHandler.CreateProjectPermission)
		}

		// Routes d'administration (nécessitent le rôle admin)
		admin := v1.Group("/admin")
		admin.Use(authHandler.RequireAuth())
		admin.Use(authHandler.RequireRole("admin"))
		{
			admin.GET("/users", authHandler.ListUsers)
			admin.GET("/users/:id", authHandler.GetUser)
			admin.PUT("/users/:id", authHandler.UpdateUser)
			admin.DELETE("/users/:id", authHandler.DeleteUser)
			admin.PUT("/users/:id/role", authHandler.UpdateUserRole)
		}
	}

	// Routes internes (réseau Docker uniquement, pas exposées via Kong)
	internal := router.Group("/internal")
	internal.Use(internalOnlyMiddleware())
	{
		cfg := internal.Group("/config")
		cfg.GET("/:service", configHandler.GetServiceConfig)
		cfg.GET("/:service/:key", configHandler.GetServiceKey)
		cfg.POST("/:service", configHandler.SetServiceConfigs)
	}

	return router
}

// internalOnlyMiddleware bloque les requêtes qui ne viennent pas du réseau Docker interne.
func internalOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		// Autorise localhost et le réseau Docker (172.x, 10.x)
		if len(ip) > 0 && (ip == "127.0.0.1" || ip == "::1" ||
			len(ip) >= 3 && (ip[:3] == "172" || ip[:2] == "10")) {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
