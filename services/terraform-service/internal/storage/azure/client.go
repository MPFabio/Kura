package azure

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/modulops/terraform-service/internal/storage"
)

// Client encapsule le client Azure Blob Storage.
type Client struct {
	blobClient *azblob.Client
}

// NewClient crée un nouveau client Azure Blob Storage.
func NewClient(accountName, accountKey, connectionString string) (*Client, error) {
	var blobClient *azblob.Client
	var err error
	
	// Priorité : connection string > account key > default credentials
	if connectionString != "" {
		blobClient, err = azblob.NewClientFromConnectionString(connectionString, nil)
	} else if accountName != "" && accountKey != "" {
		accountURL := fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
		credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la création des credentials: %w", err)
		}
		blobClient, err = azblob.NewClientWithSharedKeyCredential(accountURL, credential, nil)
	} else {
		// Utiliser les credentials par défaut (Managed Identity, Azure CLI, etc.)
		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du chargement des credentials Azure: %w", err)
		}
		// Nécessite le nom du compte de stockage
		if accountName == "" {
			return nil, fmt.Errorf("account_name requis pour l'authentification par défaut")
		}
		accountURL := fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
		blobClient, err = azblob.NewClient(accountURL, credential, nil)
	}
	
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client Azure Blob Storage: %w", err)
	}
	
	return &Client{
		blobClient: blobClient,
	}, nil
}

// GetStateFile récupère un fichier tfstate depuis Azure Blob Storage.
func (c *Client) GetStateFile(ctx context.Context, container, blobName string) ([]byte, error) {
	resp, err := c.blobClient.DownloadStream(ctx, container, blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du blob: %w", err)
	}
	defer resp.Body.Close()
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du contenu: %w", err)
	}
	
	return data, nil
}

// GetStateFileMetadata récupère les métadonnées d'un fichier tfstate depuis Azure Blob Storage.
func (c *Client) GetStateFileMetadata(ctx context.Context, container, blobName string) (*storage.StateFileMetadata, error) {
	blobClient := c.blobClient.ServiceClient().NewContainerClient(container).NewBlobClient(blobName)
	resp, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des métadonnées: %w", err)
	}
	
	metadata := &storage.StateFileMetadata{
		ETag:         string(*resp.ETag),
		LastModified: *resp.LastModified,
		Size:         *resp.ContentLength,
		VersionID:    "", // Azure Blob Storage versioning est géré différemment
	}
	
	return metadata, nil
}

// ListStateFiles liste les fichiers tfstate dans un container Azure Blob Storage.
func (c *Client) ListStateFiles(ctx context.Context, container, prefix string) ([]string, error) {
	pager := c.blobClient.NewListBlobsFlatPager(container, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})
	
	var keys []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la liste des blobs: %w", err)
		}
		
		for _, blob := range page.Segment.BlobItems {
			name := *blob.Name
			// Filtrer uniquement les fichiers .tfstate
			if strings.HasSuffix(name, ".tfstate") {
				keys = append(keys, name)
			}
		}
	}
	
	return keys, nil
}

// TestConnection teste la connexion à Azure Blob Storage.
func (c *Client) TestConnection(ctx context.Context, container string) error {
	// Lister les blobs (limite 1) pour tester la connexion
	pager := c.blobClient.NewListBlobsFlatPager(container, &azblob.ListBlobsFlatOptions{
		MaxResults: func() *int32 { n := int32(1); return &n }(),
	})
	
	if !pager.More() {
		// Container vide, mais connexion OK
		return nil
	}
	
	_, err := pager.NextPage(ctx)
	if err != nil {
		return fmt.Errorf("impossible d'accéder au container %s: %w", container, err)
	}
	
	return nil
}
