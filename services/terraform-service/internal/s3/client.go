package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/modulops/terraform-service/internal/storage"
)

// Client encapsule le client S3 AWS et implémente storage.Client.
type Client struct {
	s3Client *s3.Client
}

// NewClient crée un nouveau client S3.
func NewClient(region, endpoint, accessKeyID, secretAccessKey string) (*Client, error) {
	ctx := context.Background()
	
	var cfg aws.Config
	var err error
	
	// Si des credentials sont fournis, les utiliser
	if accessKeyID != "" && secretAccessKey != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		)
	} else {
		// Sinon, utiliser les credentials par défaut (IAM role, env vars, etc.)
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}
	
	if err != nil {
		return nil, fmt.Errorf("erreur lors du chargement de la configuration AWS: %w", err)
	}
	
	s3Client := s3.NewFromConfig(cfg)
	
	// Si un endpoint personnalisé est fourni (pour S3-compatible), le configurer
	if endpoint != "" {
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}
	
	return &Client{
		s3Client: s3Client,
	}, nil
}

// GetStateFile récupère un fichier tfstate depuis S3.
func (c *Client) GetStateFile(ctx context.Context, bucket, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	
	result, err := c.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'objet S3: %w", err)
	}
	defer result.Body.Close()
	
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du contenu: %w", err)
	}
	
	return data, nil
}

// GetStateFileMetadata récupère les métadonnées d'un fichier tfstate depuis S3.
func (c *Client) GetStateFileMetadata(ctx context.Context, bucket, key string) (*storage.StateFileMetadata, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	
	result, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des métadonnées: %w", err)
	}
	
	metadata := &storage.StateFileMetadata{
		ETag:         aws.ToString(result.ETag),
		LastModified: aws.ToTime(result.LastModified),
		Size:         aws.ToInt64(result.ContentLength),
		VersionID:    aws.ToString(result.VersionId),
	}
	
	return metadata, nil
}

// ListStateFiles liste les fichiers tfstate dans un bucket S3 (optionnel, avec préfixe).
func (c *Client) ListStateFiles(ctx context.Context, bucket, prefix string) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}
	
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(c.s3Client, input)
	
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la liste des objets: %w", err)
		}
		
		for _, obj := range page.Contents {
			// Filtrer uniquement les fichiers .tfstate
			key := aws.ToString(obj.Key)
			if len(key) > 8 && key[len(key)-8:] == ".tfstate" {
				keys = append(keys, key)
			}
		}
	}
	
	return keys, nil
}

// TestConnection teste la connexion à S3.
func (c *Client) TestConnection(ctx context.Context, bucket string) error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}
	
	_, err := c.s3Client.HeadBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("impossible d'accéder au bucket %s: %w", bucket, err)
	}
	
	return nil
}
