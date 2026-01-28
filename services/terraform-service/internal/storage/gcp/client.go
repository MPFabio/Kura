package gcp

import (
	"context"
	"fmt"
	"io"
	"strings"

	gcpStorage "cloud.google.com/go/storage"
	"google.golang.org/api/option"
	storageInterface "github.com/modulops/terraform-service/internal/storage"
)

// Client encapsule le client GCP Cloud Storage.
type Client struct {
	storageClient *gcpStorage.Client
}

// NewClient crée un nouveau client GCP Cloud Storage.
func NewClient(credentialsJSON string) (*Client, error) {
	ctx := context.Background()
	
	var opts []option.ClientOption
	
	// Si des credentials JSON sont fournis, les utiliser
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}
	// Sinon, utiliser les credentials par défaut (Application Default Credentials)
	
	storageClient, err := gcpStorage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client GCP Cloud Storage: %w", err)
	}
	
	return &Client{
		storageClient: storageClient,
	}, nil
}

// GetStateFile récupère un fichier tfstate depuis GCP Cloud Storage.
func (c *Client) GetStateFile(ctx context.Context, bucket, objectName string) ([]byte, error) {
	obj := c.storageClient.Bucket(bucket).Object(objectName)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'objet: %w", err)
	}
	defer reader.Close()
	
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du contenu: %w", err)
	}
	
	return data, nil
}

// GetStateFileMetadata récupère les métadonnées d'un fichier tfstate depuis GCP Cloud Storage.
func (c *Client) GetStateFileMetadata(ctx context.Context, bucket, objectName string) (*storageInterface.StateFileMetadata, error) {
	obj := c.storageClient.Bucket(bucket).Object(objectName)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des métadonnées: %w", err)
	}
	
	// Convertir MD5 en string hex
	etag := ""
	if len(attrs.MD5) > 0 {
		etag = fmt.Sprintf("%x", attrs.MD5)
	}
	
	metadata := &storageInterface.StateFileMetadata{
		ETag:         etag,
		LastModified: attrs.Updated,
		Size:         attrs.Size,
		VersionID:    fmt.Sprintf("%d", attrs.Generation),
	}
	
	return metadata, nil
}

// ListStateFiles liste les fichiers tfstate dans un bucket GCP Cloud Storage.
func (c *Client) ListStateFiles(ctx context.Context, bucket, prefix string) ([]string, error) {
	bkt := c.storageClient.Bucket(bucket)
	query := &gcpStorage.Query{
		Prefix: prefix,
	}
	
	var keys []string
	it := bkt.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la liste des objets: %w", err)
		}
		
		// Filtrer uniquement les fichiers .tfstate
		if strings.HasSuffix(attrs.Name, ".tfstate") {
			keys = append(keys, attrs.Name)
		}
	}
	
	return keys, nil
}

// TestConnection teste la connexion à GCP Cloud Storage.
func (c *Client) TestConnection(ctx context.Context, bucket string) error {
	bkt := c.storageClient.Bucket(bucket)
	
	// Vérifier que le bucket existe et est accessible
	_, err := bkt.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("impossible d'accéder au bucket %s: %w", bucket, err)
	}
	
	return nil
}
