variable "gcp_project" {
  description = "ID du projet GCP"
  type        = string
}

variable "gcp_region" {
  description = "Région GCP (ex: europe-west1)"
  type        = string
  default     = "europe-west1"
}

variable "gcp_zone" {
  description = "Zone GCP (ex: europe-west1-b)"
  type        = string
  default     = "europe-west1-b"
}

variable "machine_type" {
  description = "Type de machine (e2-micro = gratuit, e2-small = plus de RAM)"
  type        = string
  default     = "e2-micro"
}

variable "artifact_registry_repo" {
  description = "Nom du dépôt Artifact Registry"
  type        = string
  default     = "kura"
}
