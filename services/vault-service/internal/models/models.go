package models

import "time"

type SecretMetadata struct {
	Path        string    `json:"path"`
	Version     int       `json:"version,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Destroyed   bool      `json:"destroyed,omitempty"`
}

type Secret struct {
	Path     string                 `json:"path"`
	Data     map[string]interface{} `json:"data"`
	Metadata SecretMetadata         `json:"metadata,omitempty"`
}

// SecretWriteRequest payload pour créer/mettre à jour un secret.
type SecretWriteRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// AuditEntry représente une entrée dans l'audit trail de Vault.
type AuditEntry struct {
	Time      string `json:"time"`
	Type      string `json:"type"`
	Auth      Auth   `json:"auth"`
	Request   Req    `json:"request"`
}

type Auth struct {
	DisplayName string `json:"display_name"`
	TokenType   string `json:"token_type"`
}

type Req struct {
	ID        string `json:"id"`
	Operation string `json:"operation"`
	Path      string `json:"path"`
}

// VaultStatus représente l'état de santé de Vault.
type VaultStatus struct {
	Initialized bool   `json:"initialized"`
	Sealed      bool   `json:"sealed"`
	Version     string `json:"version"`
	ClusterName string `json:"cluster_name,omitempty"`
}
