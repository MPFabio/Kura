terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
}

data "google_project" "current" {
  project_id = var.gcp_project
}

# Clé SSH pour le déploiement (générée par Terraform, attachée à la VM)
resource "tls_private_key" "deploy" {
  algorithm = "ED25519"
}

# Artifact Registry pour les images Docker
resource "google_artifact_registry_repository" "kura" {
  location      = var.gcp_region
  repository_id = var.artifact_registry_repo
  description   = "Images Docker Kura/ModulOps"
  format        = "DOCKER"
}

# Compte de service pour GitHub Actions (push images)
resource "google_service_account" "github_actions" {
  account_id   = "kura-github-actions"
  display_name = "Kura GitHub Actions Deploy"
}

resource "google_artifact_registry_repository_iam_member" "github_writer" {
  project    = var.gcp_project
  location   = google_artifact_registry_repository.kura.location
  repository = google_artifact_registry_repository.kura.repository_id
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.github_actions.email}"
}

# Permettre à la VM de tirer les images (lecteur sur Artifact Registry)
resource "google_artifact_registry_repository_iam_member" "vm_reader" {
  project    = var.gcp_project
  location   = google_artifact_registry_repository.kura.location
  repository = google_artifact_registry_repository.kura.repository_id
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${data.google_project.current.number}-compute@developer.gserviceaccount.com"
}

# VM Compute Engine
resource "google_compute_instance" "kura" {
  name         = "kura-vm"
  machine_type = var.machine_type
  zone         = var.gcp_zone

  tags = ["kura", "http-server"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 20
    }
  }

  network_interface {
    network = "default"
    access_config {}
  }

  metadata = {
    ssh-keys = "ubuntu:${tls_private_key.deploy.public_key_openssh}"
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
    set -e
    apt-get update && apt-get install -y apt-transport-https ca-certificates curl gnupg
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg
    apt-get update && apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin google-cloud-cli
    usermod -aG docker ubuntu
    # Swap pour e2-micro (1GB RAM)
    fallocate -l 1G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    mkdir -p /opt/kura
  EOT

  service_account {
    scopes = ["cloud-platform"]
  }
}

# Pare-feu : SSH, HTTP, API Kong
resource "google_compute_firewall" "allow_ssh" {
  name    = "kura-allow-ssh"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["kura"]
}

resource "google_compute_firewall" "allow_http" {
  name    = "kura-allow-http"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["80", "8000", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["kura", "http-server"]
}
