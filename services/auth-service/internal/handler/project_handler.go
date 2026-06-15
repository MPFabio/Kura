package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modulops/auth-service/internal/config"
	"github.com/modulops/auth-service/internal/models"
	"github.com/modulops/auth-service/internal/service"
)

type ProjectHandler struct {
	projectService *service.ProjectService
	cfg            *config.Config
}

func NewProjectHandler(projectService *service.ProjectService, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		cfg:            cfg,
	}
}

// CreateProject crée un nouveau projet
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProject récupère un projet par son ID
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}

	project, err := h.projectService.GetProject(userID.(string), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// ListProjects récupère tous les projets de l'utilisateur
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projects, err := h.projectService.ListProjects(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": projects,
	})
}

// UpdateProject met à jour un projet
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.UpdateProject(userID.(string), projectID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject supprime un projet
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}

	if err := h.projectService.DeleteProject(userID.(string), projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "projet supprimé avec succès"})
}

// AddProjectMember ajoute un membre à un projet
func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}

	var req models.AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.projectService.AddProjectMember(userID.(string), projectID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// ListProjectMembers récupère tous les membres d'un projet
func (h *ProjectHandler) ListProjectMembers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}

	members, err := h.projectService.ListProjectMembers(userID.(string), projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": members,
	})
}

// UpdateProjectMember met à jour le rôle d'un membre
func (h *ProjectHandler) UpdateProjectMember(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Param("id")
	targetUserID := c.Param("user_id")
	if projectID == "" || targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet et id de l'utilisateur requis"})
		return
	}

	var req models.UpdateProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectService.UpdateProjectMember(userID.(string), projectID, targetUserID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "membre mis à jour avec succès"})
}

// RemoveProjectMember supprime un membre d'un projet
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}

	projectID := c.Param("id")
	targetUserID := c.Param("user_id")
	if projectID == "" || targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet et id de l'utilisateur requis"})
		return
	}

	if err := h.projectService.RemoveProjectMember(userID.(string), projectID, targetUserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "membre supprimé avec succès"})
}

// ListProjectMappings récupère les mappings d'un projet
func (h *ProjectHandler) ListProjectMappings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}
	mappings, err := h.projectService.ListProjectMappings(userID.(string), projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": mappings})
}

// CreateProjectMapping crée un mapping
func (h *ProjectHandler) CreateProjectMapping(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}
	var req models.CreateProjectMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mapping, err := h.projectService.CreateProjectMapping(userID.(string), projectID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, mapping)
}

// SetMappingGitOpsRepository met à jour le dépôt GitOps Forgejo associé à un mapping
func (h *ProjectHandler) SetMappingGitOpsRepository(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}
	projectID := c.Param("id")
	mappingID := c.Param("mapping_id")
	if projectID == "" || mappingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet et du mapping requis"})
		return
	}
	var body struct {
		ForgejoGitOpsRepository string `json:"forgejo_gitops_repository"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mapping, err := h.projectService.SetMappingGitOpsRepository(userID.(string), projectID, mappingID, body.ForgejoGitOpsRepository)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapping)
}

// DeleteProjectMapping supprime un mapping
func (h *ProjectHandler) DeleteProjectMapping(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}
	projectID := c.Param("id")
	mappingID := c.Param("mapping_id")
	if projectID == "" || mappingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet et du mapping requis"})
		return
	}
	if err := h.projectService.DeleteProjectMapping(userID.(string), projectID, mappingID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "mapping supprimé avec succès"})
}

// ListProjectPermissions récupère les permissions granulaires d'un projet
func (h *ProjectHandler) ListProjectPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}
	perms, err := h.projectService.GetProjectPermissions(userID.(string), projectID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"permissions": perms})
}

// CreateProjectPermission crée une permission granulaire
func (h *ProjectHandler) CreateProjectPermission(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "utilisateur non authentifié"})
		return
	}
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id du projet requis"})
		return
	}
	var req models.CreateProjectPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pp, err := h.projectService.CreateProjectPermission(userID.(string), projectID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pp)
}
