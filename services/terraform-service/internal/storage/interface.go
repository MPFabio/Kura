package storage

import (
	"context"
	"time"
)

// StateFileMetadata contient les métadonnées d'un fichier tfstate.
type StateFileMetadata struct {
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
	VersionID    string    `json:"version_id,omitempty"`
}

// Client définit l'interface commune pour tous les clients de stockage cloud.
type Client interface {
	// GetStateFile récupère un fichier tfstate depuis le stockage.
	GetStateFile(ctx context.Context, bucket, key string) ([]byte, error)

	// GetStateFileMetadata récupère les métadonnées d'un fichier tfstate.
	GetStateFileMetadata(ctx context.Context, bucket, key string) (*StateFileMetadata, error)

	// ListStateFiles liste les fichiers tfstate dans un bucket/container.
	ListStateFiles(ctx context.Context, bucket, prefix string) ([]string, error)

	// TestConnection teste la connexion au stockage.
	TestConnection(ctx context.Context, bucket string) error
}

// BackendWriter permet d'écrire un tfstate dans un backend (ex. S3).
// Utilisé pour persister les états uploadés lorsque TERRAFORM_STATE_BACKEND=s3.
type BackendWriter interface {
	PutStateFile(ctx context.Context, bucket, key string, data []byte) error
}
