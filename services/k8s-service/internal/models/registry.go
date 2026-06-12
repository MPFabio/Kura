package models

// RegistryRepository représente un dépôt (repository) du registre OCI interne (Zot).
type RegistryRepository struct {
	Name     string `json:"name"`
	TagCount int    `json:"tag_count"`
}

// RegistryTag représente un tag d'un dépôt, avec son statut de signature Cosign.
type RegistryTag struct {
	Name      string `json:"name"`
	Digest    string `json:"digest"`
	MediaType string `json:"media_type"`
	SizeBytes int64  `json:"size_bytes"`
	Signed    bool   `json:"signed"`
	Type      string `json:"type"`
}

// RegistryRepositoryDetail représente le détail d'un dépôt avec la liste de ses tags.
type RegistryRepositoryDetail struct {
	Name string        `json:"name"`
	Tags []RegistryTag `json:"tags"`
}
