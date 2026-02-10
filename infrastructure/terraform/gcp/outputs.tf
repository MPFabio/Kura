output "vm_external_ip" {
  description = "IP externe de la VM"
  value       = google_compute_instance.kura.network_interface[0].access_config[0].nat_ip
}

output "artifact_registry_url" {
  description = "URL du dépôt Artifact Registry"
  value       = "${var.gcp_region}-docker.pkg.dev/${var.gcp_project}/${var.artifact_registry_repo}"
}

output "ssh_command" {
  description = "Commande SSH pour se connecter à la VM"
  value       = "gcloud compute ssh ubuntu@${google_compute_instance.kura.name} --zone=${var.gcp_zone} --project=${var.gcp_project}"
}

output "github_actions_sa_email" {
  description = "Email du compte de service pour GitHub Actions (créer une clé dans la console GCP)"
  value       = google_service_account.github_actions.email
}

output "deploy_private_key" {
  description = "Clé SSH privée pour GitHub Actions (secret GCP_SSH_PRIVATE_KEY)"
  value       = tls_private_key.deploy.private_key_openssh
  sensitive   = true
}
