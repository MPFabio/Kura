package models

import "time"

// TerraformState représente un fichier tfstate.
type TerraformState struct {
	Version          int                      `json:"version"`
	TerraformVersion string                   `json:"terraform_version,omitempty"`
	Serial           int64                    `json:"serial"`
	Lineage          string                   `json:"lineage,omitempty"`
	Outputs          map[string]Output        `json:"outputs,omitempty"`
	Resources        []Resource               `json:"resources,omitempty"`
	CheckResults     []CheckResult            `json:"check_results,omitempty"`
}

// Output représente une sortie Terraform.
type Output struct {
	Value     interface{} `json:"value"`
	Type      interface{} `json:"type,omitempty"` // Peut être string ou array (ex: ["list", "string"])
	Sensitive bool        `json:"sensitive,omitempty"`
}

// Resource représente une ressource Terraform dans le tfstate.
type Resource struct {
	Module    string                 `json:"module,omitempty"`
	Mode      string                 `json:"mode"` // "managed" ou "data"
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Provider  string                 `json:"provider"`
	Instances []ResourceInstance     `json:"instances,omitempty"`
}

// ResourceInstance représente une instance de ressource.
type ResourceInstance struct {
	SchemaVersion int64                  `json:"schema_version,omitempty"`
	Attributes    map[string]interface{} `json:"attributes"`
	SensitiveAttributes []interface{}    `json:"sensitive_attributes,omitempty"`
	Dependencies  interface{}           `json:"dependencies,omitempty"` // Peut être []Dependency, []string, ou string
}

// Dependency représente une dépendance entre ressources.
type Dependency struct {
	Resource string   `json:"resource"`
	Attrs    []string `json:"attrs,omitempty"`
}

// CheckResult représente un résultat de vérification (drift detection).
type CheckResult struct {
	ObjectKind    string    `json:"object_kind,omitempty"`
	ObjectType    string    `json:"object_type,omitempty"`
	ObjectName    string    `json:"object_name,omitempty"`
	Status        string    `json:"status"` // "pass", "fail", "unknown"
	FailureReason string    `json:"failure_reason,omitempty"`
}

// StateFile représente un fichier tfstate avec métadonnées.
type StateFile struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	State       *TerraformState `json:"state"`
	UploadedAt  time.Time      `json:"uploaded_at"`
	LastChecked time.Time      `json:"last_checked,omitempty"`
	DriftResults []*DriftResult `json:"drift_results,omitempty"` // Résultats de détection de drift
}

// DriftResult représente un résultat de détection de drift.
type DriftResult struct {
	ResourceAddress string                 `json:"resource_address"`
	ResourceType    string                 `json:"resource_type"`
	Status          string                 `json:"status"` // "in_sync", "drifted", "missing", "unknown"
	Differences     []DriftDifference      `json:"differences,omitempty"`
	DetectedAt      time.Time              `json:"detected_at"`
	Message         string                 `json:"message,omitempty"`
}

// DriftDifference représente une différence détectée entre l'état Terraform et l'état réel.
type DriftDifference struct {
	Attribute   string      `json:"attribute"`
	Expected    interface{} `json:"expected"`
	Actual      interface{} `json:"actual"`
	ChangeType  string      `json:"change_type"` // "added", "removed", "modified"
}

// StateSummary représente un résumé d'un état Terraform.
type StateSummary struct {
	ResourceCount  int       `json:"resource_count"`
	OutputCount    int       `json:"output_count"`
	LastModified   time.Time `json:"last_modified,omitempty"`
	DriftCount     int       `json:"drift_count"`
}

// StateSource représente une source pour un état Terraform (S3, etc.).
type StateSource struct {
	ID          string    `json:"id"`
	StateFileID string   `json:"state_file_id"` // ID de l'état associé
	Type        string   `json:"type"`          // "s3", "azure", "gcp", "local", "terraform_cloud"
	Config      SourceConfig `json:"config"`
	Enabled     bool     `json:"enabled"`
	LastSync    time.Time `json:"last_sync,omitempty"`
	NextSync    time.Time `json:"next_sync,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SourceConfig contient la configuration spécifique à chaque type de source.
type SourceConfig struct {
	// Pour S3 (AWS)
	S3Bucket    string `json:"s3_bucket,omitempty"`
	S3Key       string `json:"s3_key,omitempty"`
	S3Region    string `json:"s3_region,omitempty"`
	S3Endpoint  string `json:"s3_endpoint,omitempty"` // Pour S3-compatible (MinIO, etc.)
	AWSAccessKeyID string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `json:"aws_secret_access_key,omitempty"` // Chiffré en base
	
	// Pour Azure Blob Storage
	AzureAccountName string `json:"azure_account_name,omitempty"`
	AzureAccountKey string `json:"azure_account_key,omitempty"` // Chiffré en base
	AzureConnectionString string `json:"azure_connection_string,omitempty"` // Chiffré en base
	AzureContainer string `json:"azure_container,omitempty"`
	AzureBlobName string `json:"azure_blob_name,omitempty"`
	
	// Pour GCP Cloud Storage
	GCPBucket string `json:"gcp_bucket,omitempty"`
	GCPObjectName string `json:"gcp_object_name,omitempty"`
	GCPCredentialsJSON string `json:"gcp_credentials_json,omitempty"` // Chiffré en base
	
	// Pour Terraform Cloud
	TerraformCloudOrg      string `json:"terraform_cloud_org,omitempty"`
	TerraformCloudWorkspace string `json:"terraform_cloud_workspace,omitempty"`
	TerraformCloudToken    string `json:"terraform_cloud_token,omitempty"` // Chiffré en base
	
	// Pour synchronisation
	SyncInterval string `json:"sync_interval,omitempty"` // "5m", "15m", "1h", etc.
	AutoSync     bool   `json:"auto_sync"`
}

// SyncJob représente un job de synchronisation.
type SyncJob struct {
	ID          string    `json:"id"`
	StateFileID string   `json:"state_file_id"`
	SourceID    string   `json:"source_id"`
	Status      string   `json:"status"` // "pending", "running", "success", "failed"
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string   `json:"error,omitempty"`
	Message     string   `json:"message,omitempty"`
}
