package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modulops/auth-service/internal/models"
	"github.com/modulops/auth-service/internal/repository"
)

type ProjectService struct {
	repo *repository.Repository
}

func NewProjectService(repo *repository.Repository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

// CreateProject crée un nouveau projet
func (s *ProjectService) CreateProject(userID string, req *models.CreateProjectRequest) (*models.Project, error) {
	// Vérifier que l'utilisateur existe
	_, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("utilisateur non trouvé")
	}

	project := models.NewProject(req.Name, req.Description, userID)

	if err := s.repo.CreateProject(project); err != nil {
		return nil, fmt.Errorf("erreur lors de la création du projet: %w", err)
	}

	return project, nil
}

// GetProject récupère un projet par son ID
func (s *ProjectService) GetProject(userID, projectID string) (*models.Project, error) {
	// Vérifier que l'utilisateur a accès au projet
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la vérification de l'accès: %w", err)
	}
	if !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}

	project, err := s.repo.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	// Charger les membres
	members, err := s.repo.GetProjectMembers(projectID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du chargement des membres: %w", err)
	}
	project.Members = members

	return project, nil
}

// ListProjects récupère tous les projets d'un utilisateur
func (s *ProjectService) ListProjects(userID string) ([]*models.Project, error) {
	projects, err := s.repo.GetProjectsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des projets: %w", err)
	}

	return projects, nil
}

// UpdateProject met à jour un projet
func (s *ProjectService) UpdateProject(userID, projectID string, req *models.UpdateProjectRequest) (*models.Project, error) {
	// Vérifier que l'utilisateur a accès au projet
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la vérification de l'accès: %w", err)
	}
	if !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}

	// Vérifier que l'utilisateur est owner ou admin du projet
	member, err := s.repo.GetProjectMember(projectID, userID)
	if err != nil {
		return nil, errors.New("vous n'êtes pas membre de ce projet")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, errors.New("seuls les propriétaires et administrateurs peuvent modifier le projet")
	}

	project, err := s.repo.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	// Mettre à jour les champs fournis
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}

	if err := s.repo.UpdateProject(project); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise à jour du projet: %w", err)
	}

	return project, nil
}

// DeleteProject supprime un projet
func (s *ProjectService) DeleteProject(userID, projectID string) error {
	// Vérifier que l'utilisateur est le propriétaire
	project, err := s.repo.GetProjectByID(projectID)
	if err != nil {
		return err
	}

	if project.OwnerID != userID {
		return errors.New("seul le propriétaire peut supprimer le projet")
	}

	return s.repo.DeleteProject(projectID)
}

// AddProjectMember ajoute un membre à un projet
func (s *ProjectService) AddProjectMember(userID, projectID string, req *models.AddProjectMemberRequest) (*models.ProjectMember, error) {
	// Vérifier que l'utilisateur a accès au projet
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la vérification de l'accès: %w", err)
	}
	if !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}

	// Vérifier que l'utilisateur est owner ou admin du projet
	member, err := s.repo.GetProjectMember(projectID, userID)
	if err != nil {
		return nil, errors.New("vous n'êtes pas membre de ce projet")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, errors.New("seuls les propriétaires et administrateurs peuvent ajouter des membres")
	}

	// Vérifier que l'utilisateur à ajouter existe
	_, err = s.repo.GetUserByID(req.UserID)
	if err != nil {
		return nil, errors.New("utilisateur non trouvé")
	}

	// Créer le membre
	newMember := &models.ProjectMember{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
		JoinedAt:  time.Now(),
	}

	if err := s.repo.AddProjectMember(newMember); err != nil {
		return nil, fmt.Errorf("erreur lors de l'ajout du membre: %w", err)
	}

	// Charger l'utilisateur pour le retourner
	user, err := s.repo.GetUserByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du chargement de l'utilisateur: %w", err)
	}
	newMember.User = user

	return newMember, nil
}

// ListProjectMembers récupère tous les membres d'un projet
func (s *ProjectService) ListProjectMembers(userID, projectID string) ([]*models.ProjectMember, error) {
	// Vérifier que l'utilisateur a accès au projet
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la vérification de l'accès: %w", err)
	}
	if !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}

	members, err := s.repo.GetProjectMembers(projectID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des membres: %w", err)
	}

	return members, nil
}

// UpdateProjectMember met à jour le rôle d'un membre
func (s *ProjectService) UpdateProjectMember(userID, projectID, targetUserID string, req *models.UpdateProjectMemberRequest) error {
	// Vérifier que l'utilisateur a accès au projet
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'accès: %w", err)
	}
	if !hasAccess {
		return errors.New("accès refusé à ce projet")
	}

	// Vérifier que l'utilisateur est owner ou admin du projet
	member, err := s.repo.GetProjectMember(projectID, userID)
	if err != nil {
		return errors.New("vous n'êtes pas membre de ce projet")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return errors.New("seuls les propriétaires et administrateurs peuvent modifier les membres")
	}

	// Ne pas permettre de modifier le rôle du propriétaire
	targetMember, err := s.repo.GetProjectMember(projectID, targetUserID)
	if err != nil {
		return errors.New("membre non trouvé")
	}
	if targetMember.Role == "owner" {
		return errors.New("le rôle du propriétaire ne peut pas être modifié")
	}

	return s.repo.UpdateProjectMember(projectID, targetUserID, req.Role)
}

// RemoveProjectMember supprime un membre d'un projet
func (s *ProjectService) RemoveProjectMember(userID, projectID, targetUserID string) error {
	// Vérifier que l'utilisateur a accès au projet
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'accès: %w", err)
	}
	if !hasAccess {
		return errors.New("accès refusé à ce projet")
	}

	// Vérifier que l'utilisateur est owner ou admin du projet
	member, err := s.repo.GetProjectMember(projectID, userID)
	if err != nil {
		return errors.New("vous n'êtes pas membre de ce projet")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return errors.New("seuls les propriétaires et administrateurs peuvent supprimer des membres")
	}

	// Ne pas permettre de supprimer le propriétaire
	targetMember, err := s.repo.GetProjectMember(projectID, targetUserID)
	if err != nil {
		return errors.New("membre non trouvé")
	}
	if targetMember.Role == "owner" {
		return errors.New("le propriétaire ne peut pas être supprimé")
	}

	return s.repo.RemoveProjectMember(projectID, targetUserID)
}

// ListProjectMappings récupère les mappings d'un projet
func (s *ProjectService) ListProjectMappings(userID, projectID string) ([]*models.ProjectMapping, error) {
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil || !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}
	return s.repo.ListProjectMappings(projectID)
}

// CreateProjectMapping crée un mapping
func (s *ProjectService) CreateProjectMapping(userID, projectID string, req *models.CreateProjectMappingRequest) (*models.ProjectMapping, error) {
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil || !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}
	member, err := s.repo.GetProjectMember(projectID, userID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		return nil, errors.New("seuls les propriétaires et administrateurs peuvent gérer les mappings")
	}

	m := &models.ProjectMapping{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if req.GitHubRepository != nil {
		m.GitHubRepository = *req.GitHubRepository
	}
	if req.TerraformStateID != nil {
		m.TerraformStateID = *req.TerraformStateID
	}
	if req.TerraformSourceID != nil {
		m.TerraformSourceID = *req.TerraformSourceID
	}
	if req.ClusterID != nil {
		m.ClusterID = *req.ClusterID
	}
	if req.ClusterNamespace != nil {
		m.ClusterNamespace = *req.ClusterNamespace
	}

	if m.GitHubRepository == "" && m.TerraformStateID == "" && m.ClusterID == "" {
		return nil, errors.New("au moins un champ requis : github_repository, terraform_state_id ou cluster_id")
	}

	if err := s.repo.CreateProjectMapping(m); err != nil {
		return nil, fmt.Errorf("erreur lors de la création du mapping: %w", err)
	}
	return m, nil
}

// DeleteProjectMapping supprime un mapping
func (s *ProjectService) DeleteProjectMapping(userID, projectID, mappingID string) error {
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil || !hasAccess {
		return errors.New("accès refusé à ce projet")
	}
	member, err := s.repo.GetProjectMember(projectID, userID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		return errors.New("seuls les propriétaires et administrateurs peuvent gérer les mappings")
	}
	return s.repo.DeleteProjectMapping(projectID, mappingID)
}

// scopeLevel retourne un entier pour comparer les scopes (admin=3, write=2, read=1)
func scopeLevel(scope string) int {
	switch scope {
	case "admin":
		return 3
	case "write":
		return 2
	case "read":
		return 1
	default:
		return 0
	}
}

// UserHasPermission vérifie si un utilisateur a au moins le scope demandé pour un module dans un projet
func (s *ProjectService) UserHasPermission(userID, projectID, module, requiredScope string) (bool, error) {
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil || !hasAccess {
		return false, nil
	}

	pp, err := s.repo.GetProjectPermission(projectID, userID, module)
	if err != nil {
		return false, err
	}

	if pp != nil {
		return scopeLevel(pp.Scope) >= scopeLevel(requiredScope), nil
	}

	// Fallback sur project_members : owner/admin = admin, member = read
	member, err := s.repo.GetProjectMember(projectID, userID)
	if err != nil {
		return false, nil
	}
	fallbackScope := "read"
	if member.Role == "owner" || member.Role == "admin" {
		fallbackScope = "admin"
	}
	return scopeLevel(fallbackScope) >= scopeLevel(requiredScope), nil
}

// GetProjectPermissions récupère les permissions pour un projet (pour un user ou toutes)
func (s *ProjectService) GetProjectPermissions(userID, projectID string) (map[string]string, error) {
	hasAccess, err := s.repo.UserHasAccessToProject(userID, projectID)
	if err != nil || !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}

	perms := make(map[string]string)
	for _, mod := range []string{"k8s", "terraform", "ansible", "pipeline"} {
		pp, err := s.repo.GetProjectPermission(projectID, userID, mod)
		if err != nil {
			return nil, err
		}
		if pp != nil {
			perms[mod] = pp.Scope
		} else {
			member, err := s.repo.GetProjectMember(projectID, userID)
			if err == nil && (member.Role == "owner" || member.Role == "admin") {
				perms[mod] = "admin"
			} else {
				perms[mod] = "read"
			}
		}
	}
	return perms, nil
}

// CreateProjectPermission crée une permission granulaire (réservé aux owners/admins du projet)
func (s *ProjectService) CreateProjectPermission(requesterID, projectID string, req *models.CreateProjectPermissionRequest) (*models.ProjectPermission, error) {
	hasAccess, err := s.repo.UserHasAccessToProject(requesterID, projectID)
	if err != nil || !hasAccess {
		return nil, errors.New("accès refusé à ce projet")
	}
	member, err := s.repo.GetProjectMember(projectID, requesterID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		return nil, errors.New("seuls les propriétaires et administrateurs peuvent gérer les permissions")
	}

	pp := &models.ProjectPermission{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    req.UserID,
		Module:    req.Module,
		Scope:     req.Scope,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.CreateProjectPermission(pp); err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la permission: %w", err)
	}
	return pp, nil
}
