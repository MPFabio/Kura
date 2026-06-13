package models

// ProjectMapping reflète le mapping projet <-> ressources externes exposé par l'auth-service.
type ProjectMapping struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	GitHubRepository  string `json:"github_repository,omitempty"`
	TerraformStateID  string `json:"terraform_state_id,omitempty"`
	TerraformSourceID string `json:"terraform_source_id,omitempty"`
	ClusterID         string `json:"cluster_id,omitempty"`
	ClusterNamespace  string `json:"cluster_namespace,omitempty"`
}

// Repository représente un dépôt GitHub lié à un projet.
type Repository struct {
	MappingID string `json:"mapping_id"`
	FullName  string `json:"full_name"` // "owner/repo"
}
