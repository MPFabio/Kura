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

// RegistryImageLayer décrit une couche de l'image et l'instruction qui l'a
// produite.
type RegistryImageLayer struct {
	// Instruction est la commande de construction, telle que conservée dans
	// l'historique de l'image (l'équivalent d'une ligne de `docker history`).
	Instruction string `json:"instruction"`
	SizeBytes   int64  `json:"size_bytes"`
	CreatedAt   string `json:"created_at,omitempty"`
	// Empty distingue les étapes qui ne produisent aucune couche (ENV, LABEL,
	// EXPOSE…) de celles qui écrivent sur le système de fichiers.
	Empty bool `json:"empty"`
}

// RegistryImageDetail décrit le contenu d'une image, tel qu'un registre OCI
// peut le restituer.
//
// Un registre ne conserve pas les sources : le Dockerfile n'y est pas, et ne
// peut pas en être extrait. Ce qu'il détient, et qui sert à auditer une image,
// c'est sa configuration d'exécution et l'historique des instructions ayant
// produit chaque couche.
type RegistryImageDetail struct {
	Repository   string               `json:"repository"`
	Tag          string               `json:"tag"`
	Digest       string               `json:"digest"`
	Architecture string               `json:"architecture,omitempty"`
	OS           string               `json:"os,omitempty"`
	CreatedAt    string               `json:"created_at,omitempty"`
	Entrypoint   []string             `json:"entrypoint,omitempty"`
	Cmd          []string             `json:"cmd,omitempty"`
	WorkingDir   string               `json:"working_dir,omitempty"`
	ExposedPorts []string             `json:"exposed_ports,omitempty"`
	Env          []string             `json:"env,omitempty"`
	Labels       map[string]string    `json:"labels,omitempty"`
	Layers       []RegistryImageLayer `json:"layers,omitempty"`
	TotalBytes   int64                `json:"total_bytes"`
}
