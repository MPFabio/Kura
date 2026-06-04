package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modulops/auth-service/internal/models"
)

// CreateProject crée un nouveau projet
func (r *Repository) CreateProject(project *models.Project) error {
	query := `INSERT INTO projects (id, name, description, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query,
		project.ID,
		project.Name,
		project.Description,
		project.OwnerID,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("erreur lors de la création du projet: %w", err)
	}

	// Ajouter le propriétaire comme membre avec le rôle "owner"
	member := &models.ProjectMember{
		ID:        uuid.New().String(),
		ProjectID: project.ID,
		UserID:    project.OwnerID,
		Role:      "owner",
		JoinedAt:  time.Now(),
	}

	if err := r.AddProjectMember(member); err != nil {
		return fmt.Errorf("erreur lors de l'ajout du propriétaire comme membre: %w", err)
	}

	return nil
}

// GetProjectByID récupère un projet par son ID
func (r *Repository) GetProjectByID(id string) (*models.Project, error) {
	query := `SELECT id, name, description, owner_id, created_at, updated_at
		FROM projects WHERE id = $1`

	project := &models.Project{}
	err := r.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("projet non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du projet: %w", err)
	}

	return project, nil
}

// GetProjectsByUserID récupère tous les projets auxquels un utilisateur appartient
func (r *Repository) GetProjectsByUserID(userID string) ([]*models.Project, error) {
	query := `SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1
		ORDER BY p.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des projets: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.OwnerID,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan du projet: %w", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des projets: %w", err)
	}

	return projects, nil
}

// UpdateProject met à jour un projet
func (r *Repository) UpdateProject(project *models.Project) error {
	query := `UPDATE projects 
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1`

	_, err := r.db.Exec(query,
		project.ID,
		project.Name,
		project.Description,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du projet: %w", err)
	}

	return nil
}

// DeleteProject supprime un projet
func (r *Repository) DeleteProject(id string) error {
	// Les membres seront supprimés automatiquement grâce à ON DELETE CASCADE
	query := `DELETE FROM projects WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du projet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification des lignes affectées: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("projet non trouvé")
	}

	return nil
}

// AddProjectMember ajoute un membre à un projet
func (r *Repository) AddProjectMember(member *models.ProjectMember) error {
	query := `INSERT INTO project_members (id, project_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role`

	_, err := r.db.Exec(query,
		member.ID,
		member.ProjectID,
		member.UserID,
		member.Role,
		member.JoinedAt,
	)

	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout du membre: %w", err)
	}

	return nil
}

// GetProjectMembers récupère tous les membres d'un projet
func (r *Repository) GetProjectMembers(projectID string) ([]*models.ProjectMember, error) {
	query := `SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,
		u.id, u.email, u.username, u.roles, u.first_name, u.last_name, u.active, u.created_at, u.updated_at, u.last_login
		FROM project_members pm
		INNER JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at ASC`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des membres: %w", err)
	}
	defer rows.Close()

	var members []*models.ProjectMember
	for rows.Next() {
		member := &models.ProjectMember{}
		user := &models.User{}
		var roles models.StringArray
		var lastLogin sql.NullTime

		err := rows.Scan(
			&member.ID,
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&user.ID,
			&user.Email,
			&user.Username,
			&roles,
			&user.FirstName,
			&user.LastName,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan du membre: %w", err)
		}

		user.Roles = roles
		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}
		user.Password = "" // Ne jamais exposer le mot de passe

		member.User = user
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des membres: %w", err)
	}

	return members, nil
}

// GetProjectMember récupère un membre spécifique d'un projet
func (r *Repository) GetProjectMember(projectID, userID string) (*models.ProjectMember, error) {
	query := `SELECT id, project_id, user_id, role, joined_at
		FROM project_members WHERE project_id = $1 AND user_id = $2`

	member := &models.ProjectMember{}
	err := r.db.QueryRow(query, projectID, userID).Scan(
		&member.ID,
		&member.ProjectID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("membre non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du membre: %w", err)
	}

	return member, nil
}

// UpdateProjectMember met à jour le rôle d'un membre
func (r *Repository) UpdateProjectMember(projectID, userID, role string) error {
	query := `UPDATE project_members SET role = $3 WHERE project_id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, projectID, userID, role)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du membre: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification des lignes affectées: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("membre non trouvé")
	}

	return nil
}

// RemoveProjectMember supprime un membre d'un projet
func (r *Repository) RemoveProjectMember(projectID, userID string) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, projectID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du membre: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification des lignes affectées: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("membre non trouvé")
	}

	return nil
}

// UserHasAccessToProject vérifie si un utilisateur a accès à un projet
func (r *Repository) UserHasAccessToProject(userID, projectID string) (bool, error) {
	query := `SELECT COUNT(*) FROM project_members WHERE project_id = $1 AND user_id = $2`

	var count int
	err := r.db.QueryRow(query, projectID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur lors de la vérification de l'accès: %w", err)
	}

	return count > 0, nil
}

// CreateProjectMapping crée un mapping projet <-> ressource externe
func (r *Repository) CreateProjectMapping(m *models.ProjectMapping) error {
	query := `INSERT INTO project_mappings (id, project_id, github_repository, terraform_state_id, terraform_source_id, cluster_id, cluster_namespace, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(query,
		m.ID, m.ProjectID, nullIfEmpty(m.GitHubRepository), nullIfEmpty(m.TerraformStateID),
		nullIfEmpty(m.TerraformSourceID), nullIfEmpty(m.ClusterID), nullIfEmpty(m.ClusterNamespace),
		m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// ListProjectMappings récupère tous les mappings d'un projet
func (r *Repository) ListProjectMappings(projectID string) ([]*models.ProjectMapping, error) {
	query := `SELECT id, project_id, COALESCE(github_repository,''), COALESCE(terraform_state_id,''), COALESCE(terraform_source_id,''), COALESCE(cluster_id,''), COALESCE(cluster_namespace,''), created_at, updated_at
		FROM project_mappings WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []*models.ProjectMapping
	for rows.Next() {
		m := &models.ProjectMapping{}
		err := rows.Scan(&m.ID, &m.ProjectID, &m.GitHubRepository, &m.TerraformStateID, &m.TerraformSourceID, &m.ClusterID, &m.ClusterNamespace, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}
	return mappings, rows.Err()
}

// DeleteProjectMapping supprime un mapping par ID
func (r *Repository) DeleteProjectMapping(projectID, mappingID string) error {
	query := `DELETE FROM project_mappings WHERE project_id = $1 AND id = $2`
	result, err := r.db.Exec(query, projectID, mappingID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("mapping non trouvé")
	}
	return nil
}

// GetProjectMappingByID récupère un mapping par ID
func (r *Repository) GetProjectMappingByID(projectID, mappingID string) (*models.ProjectMapping, error) {
	query := `SELECT id, project_id, COALESCE(github_repository,''), COALESCE(terraform_state_id,''), COALESCE(terraform_source_id,''), COALESCE(cluster_id,''), COALESCE(cluster_namespace,''), created_at, updated_at
		FROM project_mappings WHERE project_id = $1 AND id = $2`
	m := &models.ProjectMapping{}
	err := r.db.QueryRow(query, projectID, mappingID).Scan(&m.ID, &m.ProjectID, &m.GitHubRepository, &m.TerraformStateID, &m.TerraformSourceID, &m.ClusterID, &m.ClusterNamespace, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mapping non trouvé")
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetProjectPermission récupère la permission d'un utilisateur pour un projet et un module
func (r *Repository) GetProjectPermission(projectID, userID, module string) (*models.ProjectPermission, error) {
	query := `SELECT id, project_id, user_id, module, scope, created_at, updated_at
		FROM project_permissions WHERE project_id = $1 AND user_id = $2 AND module = $3`
	pp := &models.ProjectPermission{}
	err := r.db.QueryRow(query, projectID, userID, module).Scan(&pp.ID, &pp.ProjectID, &pp.UserID, &pp.Module, &pp.Scope, &pp.CreatedAt, &pp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return pp, nil
}

// ListProjectPermissions récupère toutes les permissions pour un projet
func (r *Repository) ListProjectPermissions(projectID string) ([]*models.ProjectPermission, error) {
	query := `SELECT id, project_id, user_id, module, scope, created_at, updated_at
		FROM project_permissions WHERE project_id = $1 ORDER BY user_id, module`
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.ProjectPermission
	for rows.Next() {
		pp := &models.ProjectPermission{}
		if err := rows.Scan(&pp.ID, &pp.ProjectID, &pp.UserID, &pp.Module, &pp.Scope, &pp.CreatedAt, &pp.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, pp)
	}
	return list, nil
}

// CreateProjectPermission crée une permission
func (r *Repository) CreateProjectPermission(pp *models.ProjectPermission) error {
	query := `INSERT INTO project_permissions (id, project_id, user_id, module, scope, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, pp.ID, pp.ProjectID, pp.UserID, pp.Module, pp.Scope, pp.CreatedAt, pp.UpdatedAt)
	return err
}
