# Héberger la plateforme pour des testeurs

Ce guide décrit comment mettre la solution en ligne pour la faire tester par d’autres personnes.

## Option recommandée : VM GCP (déjà prévue dans le projet)

Le Terraform GCP du projet crée une VM avec Docker et Docker Compose. C’est la voie la plus simple pour un environnement de démo/test partagé.

### 1. Créer la VM (si ce n’est pas déjà fait)

```bash
cd infrastructure/terraform/gcp
terraform init
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

Récupère l’IP publique affichée en sortie : `vm_external_ip`.

### 2. Ouvrir le port du frontend

Le firewall autorise déjà les ports 80, 443, 8000, 8001 et **5173** (interface utilisateur). Aucune modification nécessaire si tu utilises le Terraform à jour.

### 3. Déployer sur la VM

Sur ta machine, définis l’URL publique (remplace `VM_IP` par l’IP de la VM) :

```bash
export VM_IP="<vm_external_ip>"
export VITE_API_BASE_URL="http://${VM_IP}:8000"
export VITE_AUTH_URL="http://${VM_IP}:8000"
export VITE_PIPELINE_URL="http://${VM_IP}:8000"
export VITE_SOCKET_URL="http://${VM_IP}:8000"
```

Puis sur la VM (après `ssh ubuntu@${VM_IP}`) :

```bash
cd /opt/kura   # ou l’endroit où tu clones le repo
git clone <ton-repo> .   # ou copie du code

# Les variables VITE_* doivent être passées au build. Option 1 : fichier .env sur la VM
echo "VITE_API_BASE_URL=http://${VM_IP}:8000" > .env
echo "VITE_AUTH_URL=http://${VM_IP}:8000" >> .env
echo "VITE_PIPELINE_URL=http://${VM_IP}:8000" >> .env
echo "VITE_SOCKET_URL=http://${VM_IP}:8000" >> .env

docker compose up -d --build
```

Ou option 2 : build en local avec les bonnes variables, puis déploiement des images sur la VM (CI/CD ou manuel).

### 4. URLs à donner aux testeurs

- **Interface (frontend)** : `http://<VM_IP>:5173`
- **API (Kong)** : `http://<VM_IP>:8000`

Les testeurs ouvrent l’interface dans le navigateur ; le frontend appellera l’API sur la même IP.

---

## Autres options d’hébergement

| Option | Idéal pour | À noter |
|--------|------------|--------|
| **VPS (Hetzner, DigitalOcean, OVH, etc.)** | Full stack, coût fixe | Même principe : VM + Docker Compose. Installer Docker/Docker Compose, cloner le repo, définir les `VITE_*` avec l’URL publique du serveur, puis `docker compose up -d --build`. |
| **Netlify + backend ailleurs** | Frontend uniquement sur Netlify | Déploie le frontend sur Netlify (build Vite, dossier `dist`). Configure en variables de build : `VITE_API_BASE_URL`, `VITE_AUTH_URL`, etc. vers l’URL de ton backend (ex. la VM GCP ci‑dessus). Les testeurs utilisent l’URL Netlify ; les appels partent vers ton backend. |
| **Kubernetes (GKE, EKS, etc.)** | Environnement proche de la prod | Utilise les manifests dans `infrastructure/k8s/`. Il faut exposer les services (Ingress/LoadBalancer) et configurer le frontend avec l’URL publique de l’API. |

---

## Résumé

- **Oui**, tu peux héberger la solution ailleurs pour la faire tester.
- **Recommandation** : utiliser la VM GCP déjà prévue (Terraform + Docker Compose), avec les variables `VITE_*` pointant vers `http://<IP_VM>:8000`, et donner aux testeurs l’URL `http://<IP_VM>:5173`.
- Pour une démo plus “propre”, tu peux plus tard mettre un nom de domaine et un reverse proxy (Nginx/Caddy) sur le port 80/443 de la même VM.
