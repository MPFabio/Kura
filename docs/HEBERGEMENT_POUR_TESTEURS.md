# Héberger la plateforme pour des testeurs

Ce guide décrit comment mettre la solution en ligne pour la faire tester par d’autres personnes.

## Option recommandée : VM GCP (repo Kuro)

L'infrastructure (VMs GCP, réseau, firewall) et le bootstrap (cloud-init Docker Compose) sont gérés dans le repo séparé [Kuro](https://codeberg.org/MPFabio/Kuro) (Terraform/OpenTofu). Se référer à son README pour la provision et le cycle stop/start des VMs.

Une fois la VM `kuro-kura` provisionnée et démarrée, le cloud-init clone ce repo (Kura) et lance `docker compose up -d` automatiquement.

### URLs à donner aux testeurs

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
