package models

// ProjectMapping reflète le mapping projet <-> ressources externes exposé par l'auth-service.
type ProjectMapping struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	// GitHubRepository : conservé mais désactivé en prod (remplacé par ForgejoRepository).
	GitHubRepository  string `json:"github_repository,omitempty"`
	ForgejoRepository string `json:"forgejo_repository,omitempty"`
	ForgejoGitOpsRepository string `json:"forgejo_gitops_repository,omitempty"`
	TerraformStateID  string `json:"terraform_state_id,omitempty"`
	TerraformSourceID string `json:"terraform_source_id,omitempty"`
	ClusterID         string `json:"cluster_id,omitempty"`
	ClusterNamespace  string `json:"cluster_namespace,omitempty"`
}

// Repository représente un dépôt Forgejo/Codeberg lié à un projet.
type Repository struct {
	MappingID string `json:"mapping_id"`
	FullName  string `json:"full_name"` // "owner/repo"
}
