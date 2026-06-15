package models

import "time"

// ArgoApplication représente une Application ArgoCD (vue résumée).
type ArgoApplication struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Project        string `json:"project"`
	SyncStatus     string `json:"sync_status"`
	HealthStatus   string `json:"health_status"`
	RepoURL        string `json:"repo_url"`
	Path           string `json:"path"`
	TargetRevision string `json:"target_revision"`
	DestNamespace  string `json:"dest_namespace"`
	DestServer     string `json:"dest_server"`
}

// ArgoHistoryEntry représente une entrée de l'historique de déploiement d'une Application.
type ArgoHistoryEntry struct {
	ID         int64     `json:"id"`
	RevisionID string    `json:"revision_id"`
	DeployedAt time.Time `json:"deployed_at"`
	Source     string    `json:"source"`
}

// ArgoApplicationDetail représente le détail complet d'une Application, historique inclus.
type ArgoApplicationDetail struct {
	ArgoApplication
	History []ArgoHistoryEntry `json:"history"`
}

// CreateApplicationRequest représente une demande de création d'Application ArgoCD.
// SourceType vaut "git" (défaut, manifests/Path dans un dépôt Git) ou "helm"
// (chart Helm depuis un dépôt de charts, ex: catalogue ArtifactHub).
type CreateApplicationRequest struct {
	Name                string `json:"name" binding:"required"`
	Project             string `json:"project"`
	SourceType          string `json:"source_type"`
	RepoURL             string `json:"repo_url" binding:"required"`
	Path                string `json:"path"`
	Chart               string `json:"chart"`
	HelmValues          string `json:"helm_values"`
	TargetRevision      string `json:"target_revision"`
	DestNamespace       string `json:"dest_namespace" binding:"required"`
	DestServer          string `json:"dest_server"`
	SyncPolicyAutomated bool   `json:"sync_policy_automated"`
	Prune               bool   `json:"prune"`
	SelfHeal            bool   `json:"self_heal"`

	// Branch est la branche du dépôt GitOps sur laquelle commiter le manifest de
	// cette Application avant qu'ArgoCD ne la crée (flux GitOps "push avant pull").
	Branch string `json:"branch" binding:"required"`
	// CreateBranchFrom : si non vide, Branch est créée à partir de cette branche source.
	CreateBranchFrom string `json:"create_branch_from,omitempty"`
}

// InstallArgoCDRequest représente une demande d'installation d'ArgoCD, avec le choix
// de branche du dépôt GitOps pour le bootstrap de self-management.
type InstallArgoCDRequest struct {
	Branch           string `json:"branch" binding:"required"`
	CreateBranchFrom string `json:"create_branch_from,omitempty"`
}

// HelmChartSummary représente un chart Helm du catalogue (issu d'ArtifactHub).
type HelmChartSummary struct {
	PackageID   string `json:"package_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	LogoURL     string `json:"logo_url"`
	RepoURL     string `json:"repo_url"`
	RepoName    string `json:"repo_name"`
	Official    bool   `json:"official"`
	CNCF        bool   `json:"cncf"`
	Stars       int    `json:"stars"`
	HomeURL     string `json:"home_url"`
}

// RollbackRequest représente une demande de rollback vers une entrée d'historique donnée.
type RollbackRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// UpdateValuesRequest représente une demande de mise à jour des values Helm d'une Application.
type UpdateValuesRequest struct {
	Values string `json:"values"`
}

// ArgoCDStatus représente l'état d'installation et de disponibilité d'ArgoCD sur le cluster actif.
type ArgoCDStatus struct {
	Installed   bool   `json:"installed"`
	ServerReady bool   `json:"server_ready"`
	SelfManaged bool   `json:"self_managed"`
	Version     string `json:"version,omitempty"`
}
