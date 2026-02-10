package models

import (
	"time"

	"github.com/google/uuid"
)

// Project représente un projet/workspace dans le système
type Project struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"` // ID du créateur du projet
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Members     []*ProjectMember `json:"members,omitempty"` // Membres du projet (chargé séparément)
}

// ProjectMember représente un membre d'un projet avec son rôle
type ProjectMember struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"` // "owner", "admin", "member"
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	User      *User     `json:"user,omitempty"` // Utilisateur associé (chargé séparément)
}

// NewProject crée un nouveau projet avec un ID généré
func NewProject(name, description, ownerID string) *Project {
	return &Project{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateProjectRequest représente une requête de création de projet
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description,omitempty" max:"500"`
}

// UpdateProjectRequest représente une requête de mise à jour de projet
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// AddProjectMemberRequest représente une requête d'ajout de membre
type AddProjectMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member"` // "owner" ne peut pas être assigné
}

// UpdateProjectMemberRequest représente une requête de mise à jour de membre
type UpdateProjectMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}
