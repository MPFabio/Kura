# Kura

[![Version](https://img.shields.io/badge/version-1.3.0-blue.svg)](VERSION)
[![GitHub release](https://img.shields.io/github/v/release/MPFabio/Kura?label=release)](https://github.com/MPFabio/Kura/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Console d'opération DevOps unifiée — Kubernetes, Terraform, Ansible, Pipelines CI/CD, Métriques et Alertes dans une seule interface.

## Pourquoi Kura

Les équipes Ops jonglent entre Lens, Terraform Cloud, Ansible Tower, GitHub Actions et Grafana. Kura regroupe ces surfaces dans un seul portail avec une authentification centralisée et des événements corrélés entre systèmes.

Différence avec Backstage/Port : Kura est orienté **opération active** (exécuter, modifier, surveiller) plutôt que catalogue/documentation.

## Architecture

```
Frontend React/TS → Kong API Gateway → auth-service
                                     → k8s-service
                                     → terraform-service
                                     → ansible-service
                                     → pipeline-service
                                     → metrics-service

Kafka ← événements des services → alert-service
Redis  — cache distribué
PostgreSQL — persistance
Prometheus + Grafana — observabilité
```

Voir [docs/architecture.md](docs/architecture.md) pour les diagrammes détaillés et les flux Kafka.

## Prérequis

- Docker 20.10+
- Docker Compose 2.0+

## Démarrage rapide

```bash
cp .env.example .env
# Remplir les variables dans .env

docker compose -p kura --env-file .env -f docker-compose.yml up -d
```

## Déploiement production (GCP)

Le déploiement est automatisé via GitHub Actions sur push sur `master`.

**Secrets GitHub requis :**

| Secret | Description |
|--------|-------------|
| `GCP_SA_KEY` | Clé JSON du service account Terraform |
| `GCP_SA_JSON` | Clé JSON du service account runtime |
| `GCP_PROJECT_ID` | ID du projet GCP |
| `GCP_SSH_PRIVATE_KEY` | Clé SSH pour accéder à la VM |
| `GCP_VM_IP` | IP de la VM de déploiement |
| `GCP_VM_USER` | Utilisateur SSH |
| `PROD_ENV_FILE` | Fichier `.env` encodé en base64 |
| `KUBECONFIG_ANSIBLE` | Kubeconfig pour Ansible/Semaphore |

**Relancer la stack sur la VM après redémarrage :**

```bash
cd /opt/kura/current && sudo docker compose -p kura --env-file .env -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 5173 | Interface React/TypeScript |
| Kong Gateway | 8000 | API Gateway (point d'entrée unique) |
| Auth Service | 8080 | Authentification JWT |
| K8s Service | 8081 | Gestion clusters Kubernetes |
| Terraform Service | 8082 | États Terraform + détection de drift |
| Ansible Service | 8083 | Intégration Semaphore/AWX |
| Pipeline Service | 8084 | Pipelines CI/CD (GitHub Actions) |
| Metrics Service | 8086 | Agrégation Prometheus/Grafana |
| Grafana | 3000 | Dashboards métriques |
| Prometheus | 9090 | Collecte métriques |
| Semaphore | 3001 | UI Ansible |

## Fonctionnalités

### Kubernetes
- Gestion namespaces, pods, deployments, services
- Terminal interactif via WebSocket
- Actions en masse, recherche et filtrage

### Terraform
- Upload et visualisation d'états
- Synchronisation depuis S3, Azure Blob Storage, GCP Cloud Storage
- Détection de drift en temps réel via APIs cloud (GCP, AWS, Azure)

### Ansible
- Intégration Semaphore (UI incluse sur `/ansible`)
- Exécution de playbooks, inventaires

### Pipelines
- Synchronisation GitHub Actions
- Suivi des runs CI/CD

### Observabilité
- Dashboards Grafana accessibles sur `/grafana`
- Alertes via Kafka (`terraform.drift.detected`, `k8s.deployment.changed`)

## Structure

```
Kura/
├── services/
│   ├── auth-service/       # Go — Authentification JWT
│   ├── k8s-service/        # Go — Kubernetes
│   ├── terraform-service/  # Go — Terraform + drift
│   ├── ansible-service/    # Python — Ansible/Semaphore
│   ├── pipeline-service/   # Go — CI/CD
│   ├── metrics-service/    # Go — Métriques
│   └── alert-service/      # Go — Alertes
├── frontend/               # React + TypeScript + Vite
├── infrastructure/
│   ├── docker/             # Kong, Caddy, Prometheus config
│   ├── k8s/                # Manifests Kubernetes
│   └── terraform/gcp/      # Infrastructure GCP
├── docs/                   # Architecture, guides
└── scripts/                # Scripts utilitaires
```

## Commandes utiles

```bash
# Logs d'un service
docker compose -p kura logs -f auth-service

# Statut des conteneurs
docker compose -p kura ps

# Arrêter la stack
docker compose -p kura down

# Rebuild d'un service
docker compose -p kura build --parallel && docker compose -p kura up -d
```

## Licence

MIT
