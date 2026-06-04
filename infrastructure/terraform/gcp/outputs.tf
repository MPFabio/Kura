output "vm_external_ip" {
  description = "Adresse IP publique exposée."
  value       = google_compute_address.public_ip.address
}

output "ssh_command" {
  description = "Commande SSH pour accéder à la VM."
  value       = "ssh ${var.instance_user}@${google_compute_address.public_ip.address}"
}

output "deployment_directory" {
  description = "Chemin cible attendu par le workflow GitHub Actions."
  value       = "/opt/kura"
}
