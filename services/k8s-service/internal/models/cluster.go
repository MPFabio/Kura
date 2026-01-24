package models

import "time"

// Cluster représente un cluster Kubernetes configuré.
type Cluster struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Endpoint    string    `json:"endpoint"` // URL du serveur API Kubernetes
	Kubeconfig  string    `json:"kubeconfig"` // Contenu du kubeconfig (base64 ou texte)
	ProjectID   string    `json:"project_id"` // ID du projet auquel appartient le cluster
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ClusterStatus représente le statut de connexion d'un cluster.
type ClusterStatus struct {
	ClusterID   string    `json:"cluster_id"`
	Connected   bool      `json:"connected"`
	Version     string    `json:"version,omitempty"`
	NodesCount  int       `json:"nodes_count,omitempty"`
	LastChecked time.Time `json:"last_checked,omitempty"`
	Error       string    `json:"error,omitempty"`
}
