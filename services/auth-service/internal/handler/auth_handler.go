package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/modulops/auth-service/internal/config"
	"github.com/modulops/auth-service/internal/models"
	"github.com/modulops/auth-service/internal/service"
)

type AuthHandler struct {
	authService    *service.AuthService
	projectService *service.ProjectService
	cfg            *config.Config
}

func NewAuthHandler(authService *service.AuthService, projectService *service.ProjectService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		projectService: projectService,
		cfg:            cfg,
	}
}

// Register gère l'inscription d'un nouvel utilisateur
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Utilisateur créé avec succès",
		"user":    user,
	})
}

// Login gère la connexion d'un utilisateur
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken gère le rafraîchissement d'un token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout gère la déconnexion d'un utilisateur
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Si aucun refresh token n'est fourni, on considère que c'est OK
		c.JSON(http.StatusOK, gin.H{"message": "Déconnexion réussie"})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Déconnexion réussie"})
}

// GetCurrentUser récupère l'utilisateur actuel
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	user, err := h.authService.GetUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateCurrentUser met à jour l'utilisateur actuel
func (h *AuthHandler) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.UpdateUser(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword change le mot de passe de l'utilisateur actuel
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ChangePassword(userID.(string), req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mot de passe modifié avec succès"})
}

// ListUsers récupère la liste des utilisateurs (admin seulement)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := h.authService.ListUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"limit": limit,
		"offset": offset,
	})
}

// GetUser récupère un utilisateur par son ID (admin seulement)
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.authService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser met à jour un utilisateur (admin seulement)
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.UpdateUser(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser supprime un utilisateur (admin seulement)
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.authService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur supprimé avec succès"})
}

// UpdateUserRole met à jour les rôles d'un utilisateur (admin seulement)
func (h *AuthHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.UpdateUserRole(userID, req.Roles)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// RequireAuth est un middleware qui vérifie l'authentification
func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token d'authentification manquant"})
			c.Abort()
			return
		}

		// Extraire le token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "format de token invalide"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Valider le token
		claims, err := h.authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalide ou expiré"})
			c.Abort()
			return
		}

		// Ajouter les informations de l'utilisateur au contexte
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)

		c.Next()
	}
}

// GetPermissions retourne les rôles de l'utilisateur authentifié et, si un
// project_id est fourni, ses scopes par module dans ce projet.
func (h *AuthHandler) GetPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	roles, _ := c.Get("user_roles")
	rolesSlice, _ := roles.([]string)

	resp := gin.H{
		"user_id": userID,
		"roles":   rolesSlice,
	}

	if projectID := c.Query("project_id"); projectID != "" {
		modules, err := h.projectService.GetProjectPermissions(userID.(string), projectID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		resp["project_id"] = projectID
		resp["modules"] = modules
	}

	c.JSON(http.StatusOK, resp)
}

// Authorize vérifie que l'utilisateur authentifié dispose d'un scope suffisant
// sur un module d'un projet. Utilisé par les services métier comme point de
// décision central (200 = autorisé, 403 = refusé).
func (h *AuthHandler) Authorize(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Query("project_id")
	module := c.Query("module")
	scope := c.DefaultQuery("scope", "read")

	// Sans contexte projet, le token valide suffit (routes non rattachées à un projet)
	if projectID == "" || module == "" {
		c.JSON(http.StatusOK, gin.H{"allowed": true})
		return
	}

	allowed, err := h.projectService.UserHasPermission(userID.(string), projectID, module, scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"allowed": false, "error": "scope " + scope + " requis sur le module " + module})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allowed": true})
}

// RequireRole est un middleware qui vérifie qu'un utilisateur a un rôle spécifique
func (h *AuthHandler) RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        roles, exists := c.Get("user_roles")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "rôles non disponibles"})
            c.Abort()
            return
        }

        rolesSlice, ok := roles.([]string)
        if !ok {
            c.JSON(http.StatusForbidden, gin.H{"error": "format de rôles invalide"})
            c.Abort()
            return
        }

        hasRole := false
        for _, r := range rolesSlice {
            if r == role {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{"error": "accès refusé: rôle requis: " + role})
            c.Abort()
            return
        }

        c.Next()
    }
}
