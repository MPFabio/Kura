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
	Type      string      `json:"type,omitempty"`
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
	Dependencies  []Dependency          `json:"dependencies,omitempty"`
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
