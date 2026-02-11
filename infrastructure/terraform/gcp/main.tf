terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.12.0"
    }
  }
}

provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
  zone    = var.gcp_zone
}

terraform {
  backend "gcs" {
    bucket = "kura-ynov"
    prefix = "projet-ynov/kura/state"
  }
}
resource "google_compute_network" "kura" {
  name                    = "${var.instance_name}-net"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "kura" {
  name          = "${var.instance_name}-subnet"
  ip_cidr_range = "10.10.0.0/24"
  region        = var.gcp_region
  network       = google_compute_network.kura.id
}

resource "google_compute_firewall" "ssh" {
  name    = "${var.instance_name}-allow-ssh"
  network = google_compute_network.kura.id

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  direction     = "INGRESS"
  source_ranges = var.allowed_ssh_cidr
  target_tags   = ["${var.instance_name}-vm"]
}

resource "google_compute_firewall" "http" {
  name    = "${var.instance_name}-allow-http"
  network = google_compute_network.kura.id

  allow {
    protocol = "tcp"
    ports    = ["80", "443", "8000", "8001"]
  }

  direction     = "INGRESS"
  source_ranges = var.allowed_http_cidr
  target_tags   = ["${var.instance_name}-vm"]
}

resource "google_service_account" "vm" {
  account_id   = "${var.instance_name}-sa"
  display_name = "ModulOps runtime"
}

resource "google_compute_address" "public_ip" {
  name   = "${var.instance_name}-ip"
  region = var.gcp_region
}

locals {
  startup_script = <<-EOT
    #!/bin/bash
    set -euxo pipefail

    export DEBIAN_FRONTEND=noninteractive

    apt-get update
    apt-get install -y ca-certificates curl gnupg git apt-transport-https lsb-release

    install -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg

    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list

    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin

    curl -SL "https://github.com/docker/compose/releases/download/${var.docker_compose_version}/docker-compose-linux-x86_64" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

    usermod -aG docker ${var.instance_user} || true

    mkdir -p /opt/kura/releases
    chown -R ${var.instance_user}:${var.instance_user} /opt/kura
  EOT
}

resource "google_compute_instance" "vm" {
  name         = var.instance_name
  machine_type = var.instance_machine_type
  zone         = var.gcp_zone
  tags         = ["${var.instance_name}-vm"]

  boot_disk {
    initialize_params {
      image = "projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts"
      size  = var.instance_disk_size_gb
      type  = "pd-balanced"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.kura.id

    access_config {
      nat_ip = google_compute_address.public_ip.address
    }
  }

  service_account {
    email  = google_service_account.vm.email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata = {
    "ssh-keys" = "${var.instance_user}:${var.instance_ssh_public_key}"
  }

  metadata_startup_script = local.startup_script
}
