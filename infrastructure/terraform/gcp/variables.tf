variable "gcp_project" {
  description = "Identifiant du projet GCP qui héberge la plateforme."
  type        = string
}

variable "gcp_region" {
  description = "Région GCP pour les ressources (ex: europe-west1)."
  type        = string
  default     = "europe-west1"
}

variable "gcp_zone" {
  description = "Zone GCP pour la VM (ex: europe-west1-b)."
  type        = string
  default     = "europe-west1-b"
}

variable "instance_name" {
  description = "Nom de la VM qui héberge ModulOps."
  type        = string
  default     = "kura-platform"
}

variable "instance_machine_type" {
  description = "Type de machine GCE (e2-standard-4 conseillé pour docker compose)."
  type        = string
  default     = "e2-standard-4"
}

variable "instance_disk_size_gb" {
  description = "Taille du disque de la VM."
  type        = number
  default     = 200
}

variable "instance_user" {
  description = "Utilisateur Linux utilisé pour la connexion SSH et pour docker."
  type        = string
  default     = "ubuntu"
}

variable "instance_ssh_public_key" {
  description = "Clé publique SSH (format 'ssh-rsa AAA... user@host') autorisée sur la VM."
  type        = string
}

variable "allowed_ssh_cidr" {
  description = "Plages CIDR autorisées à se connecter en SSH."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "allowed_http_cidr" {
  description = "Plages CIDR autorisées à accéder aux ports HTTP/HTTPS/Kong."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "docker_compose_version" {
  description = "Version de Docker Compose installée sur la VM."
  type        = string
  default     = "v2.24.2"
}
